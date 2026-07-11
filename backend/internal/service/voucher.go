package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DRP-MikREST/backend/internal/models"
	"github.com/DRP-MikREST/backend/internal/repository"
	"github.com/DRP-MikREST/backend/internal/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// timeNow ambil waktu sekarang (helper, bisa di-override di test).
func timeNow() time.Time { return time.Now() }
var _ = util.NewUUID

type VoucherService struct {
	vouchers *repository.VoucherRepository
	batches  *repository.BatchRepository
	profiles *repository.ProfileRepository
	servers  *ServerService
	audit    *repository.AuditRepository
}

func NewVoucherService(
	vouchers *repository.VoucherRepository,
	batches *repository.BatchRepository,
	profiles *repository.ProfileRepository,
	servers *ServerService,
	audit *repository.AuditRepository,
) *VoucherService {
	return &VoucherService{
		vouchers: vouchers, batches: batches, profiles: profiles,
		servers: servers, audit: audit,
	}
}

type GenerateInput struct {
	ServerID     uuid.UUID  `json:"server_id" validate:"required"`
	ProfileID    *uuid.UUID `json:"profile_id"`
	Count        int        `json:"count" validate:"min=1,max=500"`
	Pattern      string     `json:"pattern"`
	Prefix       string     `json:"prefix"`
	UsernameMode string     `json:"username_mode"`
	Comment      string     `json:"comment"`
	LimitUptime  string     `json:"limit_uptime"`
	LimitBytes   string     `json:"limit_bytes"`
}

// UnmarshalJSON menerima count sebagai number (10) atau string ("10").
func (in *GenerateInput) UnmarshalJSON(data []byte) error {
	type alias GenerateInput
	aux := &struct {
		Count json.RawMessage `json:"count"`
		*alias
	}{alias: (*alias)(in)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if aux.Count != nil {
		var n int
		if err := json.Unmarshal(aux.Count, &n); err == nil {
			in.Count = n
			return nil
		}
		var s string
		if err := json.Unmarshal(aux.Count, &s); err != nil {
			return fmt.Errorf("count: %w", err)
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("count: %w", err)
		}
		in.Count = n
	}
	return nil
}

type GenerateResult struct {
	Batch   *models.VoucherBatch `json:"batch"`
	Created []models.Voucher      `json:"vouchers"`
	Failed  []FailedVoucher       `json:"failed"`
}

type FailedVoucher struct {
	Username string `json:"username"`
	Error    string `json:"error"`
}

// Generate membuat batch voucher: insert DB + push ke RouterOS.
// Strategy: commit DB dulu, lalu push RouterOS. Jika gagal, mark per voucher.
//
// `limit_uptime` dikirim ke RouterOS untuk auto-disconnect setelah user pertama
// kali login selama durasi tersebut. expires_at di DB di-set saat first-login
// (lihat SyncFromRouter) — voucher yang belum pernah dipakai TIDAK expire.
func (s *VoucherService) Generate(ctx context.Context, in GenerateInput, userID uuid.UUID) (*GenerateResult, error) {
	if in.Count <= 0 {
		in.Count = 1
	}
	if in.UsernameMode == "" {
		in.UsernameMode = "random"
	}

	// validasi limit_uptime (jika diisi)
	if in.LimitUptime != "" {
		if _, err := util.ParseRouterOSDuration(in.LimitUptime); err != nil {
			return nil, fmt.Errorf("limit_uptime tidak valid (%q): %w", in.LimitUptime, err)
		}
	}

	// ambil nama profile untuk dikirim ke RouterOS
	profileName := ""
	if in.ProfileID != nil {
		p, err := s.profiles.FindByID(ctx, *in.ProfileID)
		if err != nil {
			return nil, fmt.Errorf("cari profile: %w", err)
		}
		if p == nil {
			return nil, fmt.Errorf("profile tidak ditemukan")
		}
		profileName = p.Name
	}

	// buat batch
	batch := &models.VoucherBatch{
		ServerID:     in.ServerID,
		ProfileID:    in.ProfileID,
		Count:        in.Count,
		Pattern:      in.Pattern,
		Prefix:       in.Prefix,
		UsernameMode: in.UsernameMode,
		CreatedBy:    &userID,
	}
	if err := s.batches.Create(ctx, batch); err != nil {
		return nil, fmt.Errorf("buat batch: %w", err)
	}

	res := &GenerateResult{Batch: batch}

	// dapatkan koneksi router
	cl, err := s.servers.GetClient(ctx, in.ServerID)
	if err != nil {
		// tidak bisa konek router: tetap simpan voucher DB sebagai failed
		for i := 0; i < in.Count; i++ {
			user, pass, gerr := s.genCredential(in, i)
			if gerr != nil {
				return nil, gerr
			}
			v := &models.Voucher{
				ServerID: in.ServerID, ProfileID: in.ProfileID, BatchID: &batch.ID,
				Username: user, Password: pass, Comment: in.Comment,
				LimitUptime: in.LimitUptime,
				Status:      "failed", CreatedBy: &userID,
			}
			if derr := s.vouchers.Create(ctx, v); derr != nil {
				log.Error().Err(derr).Msg("simpan voucher failed")
				continue
			}
			res.Failed = append(res.Failed, FailedVoucher{Username: user, Error: "router tidak terhubung: " + err.Error()})
		}
		return res, nil
	}

	// urut pakai untuk mode prefix
	seq := 0
	for i := 0; i < in.Count; i++ {
		user, pass, gerr := s.genCredential(in, i)
		if gerr != nil {
			return nil, gerr
		}

		// cek duplikasi username di DB
		exists, err := s.vouchers.ExistsByUsername(ctx, in.ServerID, user)
		if err != nil {
			return nil, err
		}
		if exists {
			// regenerate sekali
			user, pass, _ = s.genCredential(in, i+1000)
			exists, _ = s.vouchers.ExistsByUsername(ctx, in.ServerID, user)
			if exists {
				user = user + util.NewUUID()[:4]
			}
		}

		v := &models.Voucher{
			ServerID: in.ServerID, ProfileID: in.ProfileID, BatchID: &batch.ID,
			Username: user, Password: pass, Comment: in.Comment,
			LimitUptime: in.LimitUptime,
			Status:      "active", CreatedBy: &userID,
		}
		if err := s.vouchers.Create(ctx, v); err != nil {
			res.Failed = append(res.Failed, FailedVoucher{Username: user, Error: err.Error()})
			continue
		}

		// push ke RouterOS dengan limit-uptime & limit-bytes-total
		ru, perr := cl.AddHotspotUser(user, pass, profileName, in.LimitUptime, in.LimitBytes, in.Comment)
		if perr != nil {
			_ = s.vouchers.MarkFailed(ctx, v.ID)
			res.Failed = append(res.Failed, FailedVoucher{Username: user, Error: perr.Error()})
			continue
		}
		if ru.ID != "" {
			_ = s.vouchers.UpdateRouterOSID(ctx, v.ID, ru.ID)
		}
		res.Created = append(res.Created, *v)
		seq++
	}

	voucherCodes := make([]string, len(res.Created))
	for i, cv := range res.Created {
		voucherCodes[i] = cv.Username
	}
	_ = s.audit.Log(ctx, &userID, &in.ServerID, "voucher.generate",
		fmt.Sprintf("batch=%s count=%d ok=%d fail=%d vouchers=%s limit_uptime=%s limit_bytes=%s comment=%s", batch.ID, in.Count, len(res.Created), len(res.Failed), strings.Join(voucherCodes, ","), in.LimitUptime, in.LimitBytes, in.Comment), nil)

	return res, nil
}

// genCredential menghasilkan (username, password) untuk voucher.
// Username = password selalu, karena voucher = username==password.
// Kalau username!=password berarti member.
func (s *VoucherService) genCredential(in GenerateInput, idx int) (string, string, error) {
	switch in.UsernameMode {
	case "prefix":
		user := util.PrefixedName(in.Prefix, idx+1, 3)
		return user, user, nil
	case "same":
		cred, err := util.ApplyPattern(in.Pattern, 8)
		if err != nil {
			return "", "", err
		}
		return cred, cred, nil
	default: // random
		user, err := util.ApplyPattern(in.Pattern, 8)
		if err != nil {
			return "", "", err
		}
		return user, user, nil
	}
}

func (s *VoucherService) List(ctx context.Context, f repository.VoucherFilter) ([]models.Voucher, int, error) {
	return s.vouchers.List(ctx, f)
}

func (s *VoucherService) Get(ctx context.Context, id uuid.UUID) (*models.Voucher, error) {
	return s.vouchers.FindByID(ctx, id)
}

func (s *VoucherService) Disable(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	v, err := s.vouchers.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if v == nil {
		return fmt.Errorf("voucher tidak ditemukan")
	}
	cl, err := s.servers.GetClient(ctx, v.ServerID)
	if err != nil {
		return err
	}
	if v.RouterOSID != "" {
		if err := cl.DisableHotspotUser(v.RouterOSID); err != nil {
			return err
		}
	}
	if err := s.vouchers.UpdateStatus(ctx, id, "disabled"); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &userID, &v.ServerID, auditAction("member.disable", "voucher.disable", v), v.Username, nil)
	return nil
}

func (s *VoucherService) Enable(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	v, err := s.vouchers.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if v == nil {
		return fmt.Errorf("voucher tidak ditemukan")
	}
	cl, err := s.servers.GetClient(ctx, v.ServerID)
	if err != nil {
		return err
	}
	if v.RouterOSID != "" {
		if err := cl.EnableHotspotUser(v.RouterOSID); err != nil {
			return err
		}
	}
	if err := s.vouchers.UpdateStatus(ctx, id, "active"); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &userID, &v.ServerID, auditAction("member.enable", "voucher.enable", v), v.Username, nil)
	return nil
}

func (s *VoucherService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	v, err := s.vouchers.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if v == nil {
		return fmt.Errorf("voucher tidak ditemukan")
	}
	cl, err := s.servers.GetClient(ctx, v.ServerID)
	if err == nil && v.RouterOSID != "" {
		_ = cl.RemoveHotspotUser(v.RouterOSID)
	}
	if err := s.vouchers.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &userID, &v.ServerID, auditAction("member.delete", "voucher.delete", v), v.Username, nil)
	return nil
}

func auditAction(memberAction, voucherAction string, v *models.Voucher) string {
	if v.Username != v.Password {
		return memberAction
	}
	return voucherAction
}

type MemberInput struct {
	ServerID    uuid.UUID  `json:"server_id" validate:"required"`
	Username    string     `json:"username" validate:"required"`
	Password    string     `json:"password" validate:"required"`
	ProfileID   *uuid.UUID `json:"profile_id"`
	Comment     string     `json:"comment"`
	LimitUptime string     `json:"limit_uptime"`
	LimitBytes  string     `json:"limit_bytes"`
}

func (s *VoucherService) CreateMember(ctx context.Context, in MemberInput, userID uuid.UUID) (*models.Voucher, error) {
	if in.Username == "" || in.Password == "" {
		return nil, fmt.Errorf("username dan password wajib diisi")
	}
	if in.Username == in.Password {
		return nil, fmt.Errorf("username dan password harus berbeda untuk member")
	}
	if in.LimitUptime != "" {
		if _, err := util.ParseRouterOSDuration(in.LimitUptime); err != nil {
			return nil, fmt.Errorf("limit_uptime tidak valid (%q): %w", in.LimitUptime, err)
		}
	}

	// cek duplikasi username
	exists, err := s.vouchers.ExistsByUsername(ctx, in.ServerID, in.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username '%s' sudah ada di server ini", in.Username)
	}

	// ambil nama profile
	profileName := ""
	if in.ProfileID != nil {
		p, err := s.profiles.FindByID(ctx, *in.ProfileID)
		if err != nil {
			return nil, fmt.Errorf("cari profile: %w", err)
		}
		if p != nil {
			profileName = p.Name
		}
	}

	// dapatkan koneksi router
	cl, err := s.servers.GetClient(ctx, in.ServerID)
	if err != nil {
		return nil, fmt.Errorf("router tidak terhubung: %w", err)
	}

	// push ke RouterOS dulu
	ru, perr := cl.AddHotspotUser(in.Username, in.Password, profileName, in.LimitUptime, in.LimitBytes, in.Comment)
	if perr != nil {
		return nil, fmt.Errorf("gagal push ke router: %w", perr)
	}

	v := &models.Voucher{
		ServerID:    in.ServerID,
		ProfileID:   in.ProfileID,
		Username:    in.Username,
		Password:    in.Password,
		Comment:     in.Comment,
		LimitUptime: in.LimitUptime,
		Status:      "active",
		CreatedBy:   &userID,
	}
	if ru.ID != "" {
		v.RouterOSID = ru.ID
	}
	if err := s.vouchers.Create(ctx, v); err != nil {
		return nil, fmt.Errorf("simpan ke DB: %w", err)
	}

	_ = s.audit.Log(ctx, &userID, &in.ServerID, "member.create",
		fmt.Sprintf("user=%s pass=%s limit_uptime=%s limit_bytes=%s comment=%s", v.Username, v.Password, in.LimitUptime, in.LimitBytes, in.Comment), nil)

	return v, nil
}

// SyncResult menyimpan ringkasan hasil sinkronisasi dari router.
type SyncResult struct {
	Updated  int `json:"updated"`  // voucher di-update (status, routeros_id, first-use)
	Imported int `json:"imported"` // voucher di-impor dari router
	Removed  int `json:"removed"`  // voucher di DB yang dihapus karena tidak ada di router
}

// SyncFromRouter melakukan sinkronisasi dua arah (router sebagai source of truth):
//   - update voucher yang sudah ada di DB (status, routeros_id, first-use)
//   - impor voucher dari router yang belum ada di DB
//   - HAPUS voucher di DB yang tidak ada lagi di router (mis. dihapus manual di router,
//     auto-expired scheduler, atau batch remove). Ini memastikan DB mirror router.
//
// Saat first-login (router uptime > 0) untuk voucher dengan limit_uptime:
//   used_at = now(), expires_at = now() + limit_uptime (atomic, hanya sekali).
func (s *VoucherService) SyncFromRouter(ctx context.Context, serverID uuid.UUID, userID *uuid.UUID) (*SyncResult, error) {
	cl, err := s.servers.GetClient(ctx, serverID)
	if err != nil {
		return nil, err
	}
	users, err := cl.ListHotspotUsers()
	if err != nil {
		return nil, err
	}

	// muat profile DB untuk mapping profile name -> id
	profileRows, err := s.profiles.ListByServer(ctx, serverID)
	if err != nil {
		profileRows = nil
	}
	profileMap := make(map[string]uuid.UUID, len(profileRows))
	for _, p := range profileRows {
		profileMap[p.Name] = p.ID
	}

	// bangun set username + map name->RouterOS ID dari router
	routerUsernames := make(map[string]bool, len(users))
	routerIDByName := make(map[string]string, len(users))
	for _, u := range users {
		// skip user bawaan router
		if u.Name == "default-trial" || u.Name == "admin" || u.Dynamic {
			continue
		}
		routerUsernames[u.Name] = true
		routerIDByName[u.Name] = u.ID
	}

	// muat semua voucher DB untuk server ini (untuk deteksi yang perlu dihapus)
	dbVouchers, err := s.vouchers.ListByServer(ctx, serverID, 1000)
	if err != nil {
		return nil, err
	}

	result := &SyncResult{}

	// ===== Pass 1: sync voucher DB yang ada di router (update) & impor yang belum ada =====
	for _, u := range users {
		if u.Name == "default-trial" || u.Name == "admin" || u.Dynamic {
			continue
		}

		v, err := s.vouchers.FindByUsername(ctx, serverID, u.Name)
		if err != nil {
			continue
		}

		hasNonZeroUptime := u.Uptime != "" && !strings.HasPrefix(u.Uptime, "0s") && u.Uptime != "00:00:00"
		newStatus := "active"
		if u.Disabled {
			newStatus = "disabled"
		} else if hasNonZeroUptime {
			newStatus = "used"
		}

		// cek apakah uptime sudah mencapai limit (voucher expired karena habis masa aktif)
		limitExhausted := false
		if v != nil && v.LimitUptime != "" && hasNonZeroUptime {
			if limitDur, err1 := util.ParseRouterOSDuration(v.LimitUptime); err1 == nil {
				if uptimeDur, err2 := util.ParseRouterOSUptime(u.Uptime); err2 == nil && uptimeDur >= limitDur {
					limitExhausted = true
				}
			}
		}
		if limitExhausted {
			newStatus = "expired"
		}

		// first-use: router menandakan user pernah login (uptime > 0)
		isFirstUse := hasNonZeroUptime

		if v != nil {
			// update voucher yang sudah ada
			needUpdate := false
			if v.Status != newStatus {
				_ = s.vouchers.UpdateStatus(ctx, v.ID, newStatus)
				needUpdate = true
			}
			// sync comment dari RouterOS (OnLogin script menaruh timestamp di comment)
			if u.Comment != "" && u.Comment != v.Comment {
				_ = s.vouchers.UpdateComment(ctx, v.ID, u.Comment)
				needUpdate = true
			}
			// disable di router kalau expired
			if limitExhausted && u.ID != "" && !u.Disabled {
				_ = cl.DisableHotspotUser(u.ID)
			}
			if u.ID != "" && v.RouterOSID == "" {
				_ = s.vouchers.UpdateRouterOSID(ctx, v.ID, u.ID)
				needUpdate = true
			}
			// first-use: set used_at & expires_at saat user pertama kali online
			if isFirstUse && v.ExpiresAt == nil && v.LimitUptime != "" {
				if d, derr := util.ParseRouterOSDuration(v.LimitUptime); derr == nil {
					now := timeNow()
					_ = s.vouchers.MarkFirstUse(ctx, v.ID, now, now.Add(d))
					needUpdate = true
				}
			}
			if needUpdate {
				result.Updated++
			}
		} else {
			// impor voucher dari router yang belum ada di DB
			var profileID *uuid.UUID
			if pid, ok := profileMap[u.Profile]; ok {
				profileID = &pid
			}
			importStatus := newStatus
			// cek apakah imported user sudah habis limitnya
			if importStatus != "expired" && u.LimitUptime != "" && hasNonZeroUptime {
				if limitDur, err1 := util.ParseRouterOSDuration(u.LimitUptime); err1 == nil {
					if uptimeDur, err2 := util.ParseRouterOSUptime(u.Uptime); err2 == nil && uptimeDur >= limitDur {
						importStatus = "expired"
					}
				}
			}
			newV := &models.Voucher{
				ServerID:    serverID,
				ProfileID:   profileID,
				Username:    u.Name,
				Password:    u.Password,
				Comment:     u.Comment,
				Status:      importStatus,
				Uptime:      u.Uptime,
				RouterOSID:  u.ID,
			}
			if isFirstUse && u.LimitUptime != "" {
				if d, derr := util.ParseRouterOSDuration(u.LimitUptime); derr == nil {
					now := timeNow()
					newV.UsedAt = &now
					exp := now.Add(d)
					newV.ExpiresAt = &exp
				}
			}
			if newV.LimitUptime == "" {
				newV.LimitUptime = u.LimitUptime
			}
			if err := s.vouchers.Create(ctx, newV); err != nil {
				// skip kalau gagal (mis. duplicate), lanjut ke user berikutnya
				continue
			}
			result.Imported++
		}
	}

	// ===== Pass 2: hapus voucher DB yang tidak ada di router =====
	// Skip voucher dengan status 'failed' (mungkin belum sempat push ke router,
	// jangan dihapus karena masih valid). Untuk status lain, hapus.
	for i := range dbVouchers {
		v := &dbVouchers[i]
		if v.Status == "failed" {
			continue
		}
		if routerUsernames[v.Username] {
			continue // masih ada di router
		}
		if err := s.vouchers.Delete(ctx, v.ID); err != nil {
			log.Warn().Err(err).Str("user", v.Username).Msg("sync: gagal hapus orphan voucher")
			continue
		}
		_ = s.audit.Log(ctx, userID, &serverID, "voucher.sync_remove", v.Username, nil)
		result.Removed++
	}

	return result, nil
}


