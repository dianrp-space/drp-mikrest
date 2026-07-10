package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/drp-mikrest/backend/internal/crypto"
	"github.com/drp-mikrest/backend/internal/models"
	"github.com/drp-mikrest/backend/internal/repository"
	"github.com/drp-mikrest/backend/internal/routeros"
	"github.com/google/uuid"
)

type ServerService struct {
	repos  *repository.ServerRepository
	audit  *repository.AuditRepository
	cryptoKey []byte

	mu     sync.Mutex
	clients map[uuid.UUID]*routeros.Client
}

func NewServerService(repos *repository.ServerRepository, audit *repository.AuditRepository, appSecret string) *ServerService {
	return &ServerService{
		repos:     repos,
		audit:     audit,
		cryptoKey: crypto.DeriveKey(appSecret),
		clients:   make(map[uuid.UUID]*routeros.Client),
	}
}

type ServerInput struct {
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Host     string `json:"host" validate:"required,min=1,max=255"`
	APIPort  int    `json:"api_port" validate:"min=1,max=65535"`
	Username string `json:"username" validate:"required,min=1,max=100"`
	Password string `json:"password" validate:"required,min=1,max=255"`
}

func (s *ServerService) Create(ctx context.Context, in ServerInput, createdBy uuid.UUID) (*models.Server, error) {
	enc, err := crypto.Encrypt(s.cryptoKey, in.Password)
	if err != nil {
		return nil, err
	}
	srv := &models.Server{
		Name:        in.Name,
		Host:        in.Host,
		APIPort:     in.APIPort,
		Username:    in.Username,
		PasswordEnc: enc,
		Status:      "unknown",
		CreatedBy:   &createdBy,
	}
	if err := s.repos.Create(ctx, srv); err != nil {
		return nil, err
	}
	_ = s.audit.Log(ctx, &createdBy, &srv.ID, "server.create", srv.Name, nil)
	return srv, nil
}

func (s *ServerService) List(ctx context.Context) ([]models.Server, error) {
	return s.repos.List(ctx)
}

func (s *ServerService) Get(ctx context.Context, id uuid.UUID) (*models.Server, error) {
	return s.repos.FindByID(ctx, id)
}

func (s *ServerService) Update(ctx context.Context, id uuid.UUID, in ServerInput, userID uuid.UUID) (*models.Server, error) {
	srv, err := s.repos.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, fmt.Errorf("server tidak ditemukan")
	}
	srv.Name = in.Name
	srv.Host = in.Host
	srv.APIPort = in.APIPort
	srv.Username = in.Username
	if in.Password != "" {
		enc, err := crypto.Encrypt(s.cryptoKey, in.Password)
		if err != nil {
			return nil, err
		}
		srv.PasswordEnc = enc
	}
	if err := s.repos.Update(ctx, srv); err != nil {
		return nil, err
	}
	s.closeClient(id)
	_ = s.audit.Log(ctx, &userID, &srv.ID, "server.update", srv.Name, nil)
	return srv, nil
}

func (s *ServerService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	s.closeClient(id)
	if err := s.repos.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &userID, &id, "server.delete", "", nil)
	return nil
}

// TestConnection membuka koneksi ke router dan update status.
func (s *ServerService) TestConnection(ctx context.Context, id uuid.UUID) (string, error) {
	srv, err := s.repos.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	if srv == nil {
		return "", fmt.Errorf("server tidak ditemukan")
	}
	pass, err := crypto.Decrypt(s.cryptoKey, srv.PasswordEnc)
	if err != nil {
		return "", err
	}
	cl, err := routeros.Dial(srv.Host, srv.APIPort, srv.Username, pass)
	if err != nil {
		_ = s.repos.UpdateStatus(ctx, id, "offline")
		return "", err
	}
	defer cl.Close()
	name, err := cl.SystemIdentity()
	if err != nil {
		_ = s.repos.UpdateStatus(ctx, id, "offline")
		return "", err
	}
	_ = s.repos.UpdateStatus(ctx, id, "online")
	return name, nil
}

// GetClient mengembalikan (atau membuat) koneksi routeros untuk server.
// Koneksi di-cache per server dan di-reuse.
func (s *ServerService) GetClient(ctx context.Context, id uuid.UUID) (*routeros.Client, error) {
	s.mu.Lock()
	if cl, ok := s.clients[id]; ok {
		s.mu.Unlock()
		return cl, nil
	}
	s.mu.Unlock()

	srv, err := s.repos.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, fmt.Errorf("server tidak ditemukan")
	}
	pass, err := crypto.Decrypt(s.cryptoKey, srv.PasswordEnc)
	if err != nil {
		return nil, err
	}
	cl, err := routeros.Dial(srv.Host, srv.APIPort, srv.Username, pass)
	if err != nil {
		_ = s.repos.UpdateStatus(ctx, id, "offline")
		return nil, err
	}
	s.mu.Lock()
	s.clients[id] = cl
	s.mu.Unlock()
	_ = s.repos.UpdateStatus(ctx, id, "online")
	return cl, nil
}

func (s *ServerService) closeClient(id uuid.UUID) {
	s.mu.Lock()
	cl, ok := s.clients[id]
	if ok {
		delete(s.clients, id)
	}
	s.mu.Unlock()
	if ok {
		cl.Close()
	}
}

// CloseAll menutup semua koneksi (dipanggil saat shutdown).
func (s *ServerService) CloseAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, cl := range s.clients {
		cl.Close()
		delete(s.clients, id)
	}
}

// DecryptPassword helper untuk service lain.
func (s *ServerService) DecryptPassword(srv *models.Server) (string, error) {
	return crypto.Decrypt(s.cryptoKey, srv.PasswordEnc)
}

var _ = time.Now
