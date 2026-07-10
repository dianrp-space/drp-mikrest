package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/drp-mikrest/backend/internal/service"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

type Scheduler struct {
	cron       *cron.Cron
	expireSvc  *service.ExpirationService
	settingSvc *service.SettingService
	auditSvc   *service.AuditService
	entryID    cron.EntryID
	mu         sync.Mutex
}

func New(expireSvc *service.ExpirationService, settingSvc *service.SettingService, auditSvc *service.AuditService) *Scheduler {
	return &Scheduler{
		cron:       cron.New(cron.WithLocation(time.Local)),
		expireSvc:  expireSvc,
		settingSvc: settingSvc,
		auditSvc:   auditSvc,
	}
}

func (s *Scheduler) Start() {
	interval := s.readInterval()
	s.schedule(interval)
	// cleanup harian jam 03:00
	_, err := s.cron.AddFunc("0 3 * * *", s.cleanupExpired)
	if err != nil {
		log.Error().Err(err).Msg("scheduler: add cleanup job")
	}
	// cleanup log harian jam 03:30
	_, err = s.cron.AddFunc("30 3 * * *", s.cleanupOldLogs)
	if err != nil {
		log.Error().Err(err).Msg("scheduler: add log cleanup job")
	}
	s.cron.Start()
	log.Info().Int("interval_min", interval).Msg("scheduler started")
}

func (s *Scheduler) Stop(ctx context.Context) {
	stopCtx := s.cron.Stop()
	select {
	case <-stopCtx.Done():
	case <-ctx.Done():
	}
	log.Info().Msg("scheduler stopped")
}

func (s *Scheduler) Reload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	interval := s.readInterval()
	s.cron.Remove(s.entryID)
	s.schedule(interval)
	log.Info().Int("interval_min", interval).Msg("scheduler reloaded")
}

func (s *Scheduler) CurrentInterval() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := s.cron.Entry(s.entryID)
	if entry.Valid() {
		next := entry.Next
		now := time.Now()
		if next.After(now) {
			return int(next.Sub(now).Minutes()) + 1
		}
	}
	return s.readInterval()
}

func (s *Scheduler) readInterval() int {
	settings, err := s.settingSvc.GetAll(context.Background())
	if err != nil {
		return 1
	}
	raw := settings["cron_interval"]
	if raw == "" {
		return 1
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 1
	}
	return n
}

func (s *Scheduler) schedule(interval int) {
	expr := fmt.Sprintf("*/%d * * * *", interval)
	id, err := s.cron.AddFunc(expr, s.checkExpired)
	if err != nil {
		log.Error().Err(err).Msg("scheduler: add job")
		return
	}
	s.entryID = id
}

func (s *Scheduler) checkExpired() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	n := s.expireSvc.RunExpiredCheck(ctx)
	if n > 0 {
		log.Info().Int("expired_count", n).Msg("scheduler: voucher expired diproses")
	}
}

func (s *Scheduler) cleanupExpired() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	n := s.expireSvc.CleanupExpired(ctx, 24*time.Hour)
	if n > 0 {
		log.Info().Int("removed", n).Msg("scheduler: expired voucher dibersihkan")
	}
}

func (s *Scheduler) readLogRetentionDays() int {
	settings, err := s.settingSvc.GetAll(context.Background())
	if err != nil {
		return 30
	}
	raw := settings["auto_delete_log_days"]
	if raw == "" {
		return 30
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 30
	}
	return n
}

func (s *Scheduler) cleanupOldLogs() {
	days := s.readLogRetentionDays()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	n, err := s.auditSvc.DeleteOlderThan(ctx, time.Duration(days)*24*time.Hour)
	if err != nil {
		log.Error().Err(err).Msg("scheduler: cleanup old logs gagal")
		return
	}
	if n > 0 {
		log.Info().Int64("deleted", n).Int("days", days).Msg("scheduler: log lama dihapus")
	}
}
