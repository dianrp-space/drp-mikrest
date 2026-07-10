package service

import (
	"context"
	"fmt"

	"github.com/drp-mikrest/backend/internal/repository"
	"github.com/drp-mikrest/backend/internal/routeros"
	"github.com/google/uuid"
)

type HotspotService struct {
	servers *ServerService
	audit   *repository.AuditRepository
}

func NewHotspotService(servers *ServerService, audit *repository.AuditRepository) *HotspotService {
	return &HotspotService{servers: servers, audit: audit}
}

func (s *HotspotService) ActiveUsers(ctx context.Context, serverID uuid.UUID) ([]routeros.ActiveUser, error) {
	cl, err := s.servers.GetClient(ctx, serverID)
	if err != nil {
		return nil, err
	}
	return cl.ListActiveUsers()
}

func (s *HotspotService) Kick(ctx context.Context, serverID uuid.UUID, rosID string, userID uuid.UUID) error {
	cl, err := s.servers.GetClient(ctx, serverID)
	if err != nil {
		return err
	}
	if err := cl.KickActiveUser(rosID); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &userID, &serverID, "hotspot.kick", rosID, nil)
	return nil
}

func (s *HotspotService) SystemResource(ctx context.Context, serverID uuid.UUID) (routeros.SystemResource, error) {
	cl, err := s.servers.GetClient(ctx, serverID)
	if err != nil {
		return routeros.SystemResource{}, err
	}
	return cl.SystemResource()
}

func (s *HotspotService) SystemIdentity(ctx context.Context, serverID uuid.UUID) (string, error) {
	cl, err := s.servers.GetClient(ctx, serverID)
	if err != nil {
		return "", err
	}
	return cl.SystemIdentity()
}

var _ = fmt.Sprintf
