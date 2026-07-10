package service

import (
	"context"
	"time"

	"github.com/drp-mikrest/backend/internal/repository"
)

type AuditService struct {
	repo *repository.AuditRepository
}

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) List(ctx context.Context, f repository.AuditFilter) ([]repository.AuditRow, int, error) {
	return s.repo.List(ctx, f)
}

func (s *AuditService) DeleteOlderThan(ctx context.Context, age time.Duration) (int64, error) {
	return s.repo.DeleteOlderThan(ctx, age)
}
