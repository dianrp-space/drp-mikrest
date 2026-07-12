package service

import (
	"context"
	"sync"
	"time"

	"github.com/DRP-MikREST/backend/internal/repository"
	"github.com/DRP-MikREST/backend/internal/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ExpirationService auto-kick & auto-hapus voucher/member yang sudah lewat masa aktifnya.
// Dijalankan periodik oleh scheduler (cron).
// Expiry dihitung dari comment (timestamp first login via OnLogin script di RouterOS)
// + limit-uptime yang tercatat di voucher/member.
type ExpirationService struct {
	vouchers *repository.VoucherRepository
	servers  *ServerService
	audit    *repository.AuditRepository

	mu      sync.Mutex
	busy    map[uuid.UUID]bool // serverID -> sedang diproses
}

func NewExpirationService(v *repository.VoucherRepository, srv *ServerService, a *repository.AuditRepository) *ExpirationService {
	return &ExpirationService{vouchers: v, servers: srv, audit: a, busy: make(map[uuid.UUID]bool)}
}

// CleanupExpired menghapus voucher/member dengan status 'expired' dari RouterOS & DB.
// Delay: hanya hapus yang sudah expired lebih dari cleanupDelay (mis. 24 jam).
func (e *ExpirationService) CleanupExpired(ctx context.Context, cleanupDelay time.Duration) int {
	list, err := e.vouchers.ListExpiredForCleanup(ctx, cleanupDelay, 200)
	if err != nil {
		log.Error().Err(err).Msg("cleanup: list expired")
		return 0
	}
	if len(list) == 0 {
		return 0
	}
	removed := 0
	for _, v := range list {
		if v.RouterOSID != "" {
			if cl, err := e.servers.GetClient(ctx, v.ServerID); err == nil {
				_ = cl.RemoveHotspotUser(v.RouterOSID)
			}
		}
		if err := e.vouchers.Delete(ctx, v.ID); err != nil {
			log.Warn().Err(err).Str("id", v.ID.String()).Msg("cleanup: gagal hapus dari DB")
			continue
		}
		_ = e.audit.Log(ctx, nil, &v.ServerID, "voucher.cleanup_deleted", v.Username, nil)
		removed++
	}
	if removed > 0 {
		log.Info().Int("removed", removed).Msg("cleanup: voucher expired dihapus")
	}
	return removed
}

// RunExpiredCheck membaca comment + limit-uptime dari RouterOS untuk menentukan
// apakah voucher/member sudah expired (now >= comment_time + limit_uptime).
// Jika expired:
//   1. kick dari /ip/hotspot/active
//   2. remove dari /ip/hotspot/user
//   3. delete dari DB
//   4. catat audit log
func (e *ExpirationService) RunExpiredCheck(ctx context.Context) int {
	servers, err := e.servers.List(ctx)
	if err != nil {
		log.Error().Err(err).Msg("expiration: list server gagal")
		return 0
	}

	processed := 0
	for _, srv := range servers {
		e.mu.Lock()
		if e.busy[srv.ID] {
			e.mu.Unlock()
			continue
		}
		e.busy[srv.ID] = true
		e.mu.Unlock()

		func() {
			defer func() {
				e.mu.Lock()
				delete(e.busy, srv.ID)
				e.mu.Unlock()
			}()

			cl, err := e.servers.GetClient(ctx, srv.ID)
			if err != nil {
				log.Warn().Err(err).Str("server", srv.Name).Msg("expiration: skip server offline")
				return
			}

			users, err := cl.ListHotspotUsers()
			if err != nil {
				log.Warn().Err(err).Str("server", srv.Name).Msg("expiration: gagal list users")
				return
			}

			active, _ := cl.ListActiveUsers()
			activeByID := make(map[string]bool, len(active))
			for _, a := range active {
				activeByID[a.ID] = true
			}

			for _, u := range users {
				if u.Name == "default-trial" || u.Name == "admin" || u.Dynamic {
					continue
				}
				if u.Comment == "" {
					continue
				}

				firstLogin, err := time.ParseInLocation("2006-01-02 15:04:05", u.Comment, time.Local)
				if err != nil {
					continue
				}

				v, err := e.vouchers.FindByUsername(ctx, srv.ID, u.Name)
				if err != nil || v == nil {
					continue
				}

				// sync comment & used_at dari RouterOS (first login timestamp)
				if u.Comment != v.Comment {
					_ = e.vouchers.UpdateComment(ctx, v.ID, u.Comment)
				}
				_ = e.vouchers.UpdateUsedAt(ctx, v.ID, firstLogin)

				// proses expiry hanya jika ada limit-uptime
				if u.LimitUptime == "" {
					continue
				}

				limitDur, err := util.ParseRouterOSDuration(u.LimitUptime)
				if err != nil {
					continue
				}

				expiresAt := firstLogin.Add(limitDur)

				_ = e.vouchers.UpdateExpiresAt(ctx, v.ID, expiresAt)

				if time.Now().Before(expiresAt) {
					continue
				}

				// kick jika online
				if activeByID[u.ID] {
					if err := cl.KickActiveUser(u.ID); err != nil {
						log.Warn().Err(err).Str("user", u.Name).Msg("expiration: kick gagal")
					}
				}

				// remove dari RouterOS
				if u.ID != "" {
					if err := cl.RemoveHotspotUser(u.ID); err != nil {
						log.Warn().Err(err).Str("user", u.Name).Msg("expiration: remove gagal")
					}
				}

				// delete dari DB
				if err := e.vouchers.Delete(ctx, v.ID); err != nil {
					log.Warn().Err(err).Str("user", u.Name).Msg("expiration: gagal hapus DB")
					continue
				}

				// audit log
				action := "voucher.expired"
				if v.Username != v.Password {
					action = "member.expired"
				}
				_ = e.audit.Log(ctx, nil, &srv.ID, action, u.Name, nil)
				processed++
				log.Info().Str("user", u.Name).Str("server", srv.Name).Msg("expired: dihapus")
			}
		}()
	}
	return processed
}
