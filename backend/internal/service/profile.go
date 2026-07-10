package service

import (
	"context"
	"fmt"

	"github.com/drp-mikrest/backend/internal/models"
	"github.com/drp-mikrest/backend/internal/repository"
	"github.com/drp-mikrest/backend/internal/routeros"
	"github.com/google/uuid"
)

type ProfileService struct {
	repos   *repository.ProfileRepository
	servers *ServerService
	audit   *repository.AuditRepository
}

func NewProfileService(repos *repository.ProfileRepository, servers *ServerService, audit *repository.AuditRepository) *ProfileService {
	return &ProfileService{repos: repos, servers: servers, audit: audit}
}

func (s *ProfileService) List(ctx context.Context, serverID uuid.UUID) ([]models.HotspotProfile, error) {
	return s.repos.ListByServer(ctx, serverID)
}

// SyncFromRouter menarik profile dari RouterOS dan upsert ke DB.
// Profile yang ada di DB tapi tidak ada di router akan dihapus.
func (s *ProfileService) SyncFromRouter(ctx context.Context, serverID uuid.UUID) (int, error) {
	cl, err := s.servers.GetClient(ctx, serverID)
	if err != nil {
		return 0, err
	}
	profiles, err := cl.ListHotspotProfiles()
	if err != nil {
		return 0, err
	}
	routerNames := make(map[string]bool, len(profiles))
	for _, p := range profiles {
		routerNames[p.Name] = true
		dbP := &models.HotspotProfile{
			ServerID:         serverID,
			Name:             p.Name,
			RateLimit:        p.RateLimit,
			SessionTimeout:   p.SessionTimeout,
			IdleTimeout:      p.IdleTimeout,
			SharedUsers:      atoiSafe(p.SharedUsers),
			KeepaliveTimeout: p.KeepaliveTimeout,
			LoginBy:          p.LoginBy,
		}
		if err := s.repos.UpsertFromRouter(ctx, dbP); err != nil {
			return 0, fmt.Errorf("upsert profile %s: %w", p.Name, err)
		}
	}

	// hapus profile DB yang tidak ada di router (kecuali is_local)
	dbRows, err := s.repos.ListByServer(ctx, serverID)
	if err == nil {
		for _, dbP := range dbRows {
			if !dbP.IsLocal && !routerNames[dbP.Name] {
				_ = s.repos.Delete(ctx, dbP.ID)
			}
		}
	}
	return len(profiles), nil
}

type ProfileInput struct {
	Name             string   `json:"name" validate:"required,min=1,max=100"`
	RateLimit        string   `json:"rate_limit"`
	SessionTimeout   string   `json:"session_timeout"`
	IdleTimeout      string   `json:"idle_timeout"`
	SharedUsers      int      `json:"shared_users"`
	KeepaliveTimeout string   `json:"keepalive_timeout"`
	LoginBy          []string `json:"login_by"`
}

func (s *ProfileService) Create(ctx context.Context, serverID uuid.UUID, in ProfileInput, userID uuid.UUID) (*models.HotspotProfile, error) {
	cl, err := s.servers.GetClient(ctx, serverID)
	if err != nil {
		return nil, err
	}
	_, err = cl.AddHotspotProfile(routeros.HotspotProfile{
		Name:             in.Name,
		RateLimit:        in.RateLimit,
		SessionTimeout:   in.SessionTimeout,
		IdleTimeout:      in.IdleTimeout,
		SharedUsers:      fmt.Sprintf("%d", in.SharedUsers),
		KeepaliveTimeout: in.KeepaliveTimeout,
		LoginBy:          in.LoginBy,
	})
	if err != nil {
		return nil, err
	}
	p := &models.HotspotProfile{
		ServerID:         serverID,
		Name:             in.Name,
		RateLimit:        in.RateLimit,
		SessionTimeout:   in.SessionTimeout,
		IdleTimeout:      in.IdleTimeout,
		SharedUsers:      in.SharedUsers,
		KeepaliveTimeout: in.KeepaliveTimeout,
		LoginBy:          in.LoginBy,
		IsLocal:          true,
	}
	if err := s.repos.Create(ctx, p); err != nil {
		return nil, err
	}
	_ = s.audit.Log(ctx, &userID, &serverID, "profile.create", in.Name, nil)
	return p, nil
}

func (s *ProfileService) Delete(ctx context.Context, serverID, profileID, userID uuid.UUID) error {
	p, err := s.repos.FindByID(ctx, profileID)
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("profile tidak ditemukan")
	}
	cl, err := s.servers.GetClient(ctx, serverID)
	if err != nil {
		return err
	}
	users, err := cl.ListHotspotUsers()
	if err == nil {
		for _, u := range users {
			if u.Profile == p.Name {
				return fmt.Errorf("profile %s masih dipakai voucher %s", p.Name, u.Name)
			}
		}
	}
	// hapus dari router jika ada (cari .id by name)
	rows, err := cl.Print("/ip/hotspot/user/profile", "name", p.Name)
	if err == nil {
		for _, r := range rows {
			if id := r[".id"]; id != "" {
				_ = cl.RemoveHotspotProfile(id)
			}
		}
	}
	if err := s.repos.Delete(ctx, profileID); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &userID, &serverID, "profile.delete", p.Name, nil)
	return nil
}

func atoiSafe(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	if n == 0 {
		return 1
	}
	return n
}
