package service

import (
	"context"

	"github.com/DRP-MikREST/backend/internal/repository"
)

type SettingService struct {
	settings *repository.SettingRepository
}

func NewSettingService(settings *repository.SettingRepository) *SettingService {
	return &SettingService{settings: settings}
}

func (s *SettingService) GetAll(ctx context.Context) (map[string]string, error) {
	return s.settings.GetAll(ctx)
}

func (s *SettingService) Set(ctx context.Context, key, value string) error {
	return s.settings.Set(ctx, key, value)
}

func (s *SettingService) BatchSet(ctx context.Context, pairs map[string]string) error {
	for k, v := range pairs {
		if err := s.settings.Set(ctx, k, v); err != nil {
			return err
		}
	}
	return nil
}
