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
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users *repository.UserRepository
	jwt   *util.JWTManager
}

func NewAuthService(users *repository.UserRepository, jwt *util.JWTManager) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

func (s *AuthService) Register(ctx context.Context, email, password, role string) (*models.User, error) {
	if role == "" {
		role = "admin"
	}
	existing, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email sudah terdaftar")
	}
	if len(password) < 8 {
		return nil, errors.New("password minimal 8 karakter")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	return s.users.Create(ctx, email, string(hash), role)
}

type TokenPair struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	TokenType   string    `json:"token_type"`
	User        *models.User `json:"user"`
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("email atau password salah")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("email atau password salah")
	}
	tok, exp, err := s.jwt.Generate(u.ID.String(), u.Email, u.Role)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken: tok,
		ExpiresAt:   exp,
		TokenType:   "Bearer",
		User:        u,
	}, nil
}

func (s *AuthService) ChangeEmail(ctx context.Context, userID uuid.UUID, password, newEmail string) error {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("user tidak ditemukan")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return errors.New("password salah")
	}
	return s.users.UpdateEmail(ctx, userID, newEmail)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("user tidak ditemukan")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("password lama salah")
	}
	if len(newPassword) < 8 {
		return errors.New("password baru minimal 8 karakter")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.users.UpdatePassword(ctx, userID, string(hash))
}

// EnsureSeedUser membuat admin awal jika tabel users masih kosong.
func (s *AuthService) EnsureSeedUser(ctx context.Context, email, password string) error {
	n, err := s.users.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err = s.Register(ctx, email, password, "admin")
	return err
}
