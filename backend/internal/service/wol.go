package service

import (
	"context"
	"fmt"

	"github.com/DRP-MikREST/backend/internal/models"
	"github.com/DRP-MikREST/backend/internal/repository"
	"github.com/google/uuid"
)

type WOLService struct {
	repo     *repository.WOLRepository
	serverSvc *ServerService
}

func NewWOLService(repo *repository.WOLRepository, serverSvc *ServerService) *WOLService {
	return &WOLService{repo: repo, serverSvc: serverSvc}
}

func (s *WOLService) Create(ctx context.Context, in models.WOLTargetInput) (*models.WOLTarget, error) {
	serverID, err := uuid.Parse(in.ServerID)
	if err != nil {
		return nil, fmt.Errorf("server_id tidak valid")
	}
	t := &models.WOLTarget{
		ServerID:      serverID,
		InterfaceName: in.InterfaceName,
		MACAddress:    in.MACAddress,
		Name:          in.Name,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *WOLService) ListByServer(ctx context.Context, serverID uuid.UUID) ([]models.WOLTarget, error) {
	return s.repo.ListByServer(ctx, serverID)
}

func (s *WOLService) ListAll(ctx context.Context) ([]models.WOLTarget, error) {
	return s.repo.ListAll(ctx)
}

func (s *WOLService) SendWOL(ctx context.Context, id uuid.UUID) error {
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if t == nil {
		return fmt.Errorf("wol target tidak ditemukan")
	}

	cl, err := s.serverSvc.GetClient(ctx, t.ServerID)
	if err != nil {
		return fmt.Errorf("gagal konek ke server: %w", err)
	}

	_, err = cl.Run("/tool/wol", fmt.Sprintf("=interface=%s", t.InterfaceName), fmt.Sprintf("=mac=%s", t.MACAddress))
	if err != nil {
		return fmt.Errorf("gagal kirim WOL: %w", err)
	}
	return nil
}

func (s *WOLService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
