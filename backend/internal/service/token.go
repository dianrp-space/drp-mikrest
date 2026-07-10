package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/drp-mikrest/backend/internal/models"
	"github.com/drp-mikrest/backend/internal/repository"
	"github.com/drp-mikrest/backend/internal/util"
	"github.com/google/uuid"
)

type TokenService struct {
	repos *repository.TokenRepository
	users *repository.UserRepository
}

func NewTokenService(repos *repository.TokenRepository, users *repository.UserRepository) *TokenService {
	return &TokenService{repos: repos, users: users}
}

type TokenCreateInput struct {
	Label     string     `json:"label" validate:"required,min=1,max=50"`
	Scopes    []string   `json:"scopes"`
	RateLimit int        `json:"rate_limit"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type TokenCreated struct {
	Token   *models.APIToken `json:"token"`
	Plain   string           `json:"plain_token"`
}

func (s *TokenService) Create(ctx context.Context, userID uuid.UUID, in TokenCreateInput) (*TokenCreated, error) {
	if in.Label == "" {
		return nil, errors.New("label wajib diisi")
	}
	if len(in.Scopes) == 0 {
		in.Scopes = []string{"vouchers:rw", "servers:ro"}
	}
	if in.RateLimit == 0 {
		in.RateLimit = 60
	}
	plain, hash, prefix, err := util.NewAPIToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	t := &models.APIToken{
		UserID:      userID,
		Label:       in.Label,
		TokenHash:   hash,
		TokenPrefix: prefix,
		Scopes:      in.Scopes,
		RateLimit:   in.RateLimit,
		ExpiresAt:   in.ExpiresAt,
	}
	if err := s.repos.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("simpan token: %w", err)
	}
	return &TokenCreated{Token: t, Plain: plain}, nil
}

func (s *TokenService) List(ctx context.Context, userID uuid.UUID) ([]models.APIToken, error) {
	return s.repos.ListByUser(ctx, userID)
}

// Delete menghapus token dari DB secara permanen (hard delete).
// Token yang dihapus tidak akan bisa dipakai untuk request API (hash hilang)
// dan hilang dari daftar token user.
func (s *TokenService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return s.repos.HardDelete(ctx, id, userID)
}

// AuthByToken memverifikasi token plain dan mengembalikan user + token.
func (s *TokenService) AuthByToken(ctx context.Context, plain string) (*models.User, *models.APIToken, error) {
	if plain == "" {
		return nil, nil, errors.New("token kosong")
	}
	hash := util.HashToken(plain)
	t, err := s.repos.FindByHash(ctx, hash)
	if err != nil {
		return nil, nil, err
	}
	if t == nil {
		return nil, nil, errors.New("token tidak valid atau sudah dicabut")
	}
	if t.ExpiresAt != nil && t.ExpiresAt.Before(time.Now()) {
		return nil, nil, errors.New("token kedaluwarsa")
	}
	u, err := s.users.FindByID(ctx, t.UserID)
	if err != nil {
		return nil, nil, err
	}
	if u == nil {
		return nil, nil, errors.New("user token tidak ditemukan")
	}
	s.repos.TouchLastUsed(ctx, t.ID)
	return u, t, nil
}
