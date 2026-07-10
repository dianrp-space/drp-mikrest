package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/drp-mikrest/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepository struct{ pool *pgxpool.Pool }

func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

// profileSelect mendefinisikan kolom SELECT untuk profile.
// Note: tidak ada kolom `comment` (profile di RouterOS tidak punya property comment).
const profileSelect = `
	SELECT id, server_id, name, rate_limit, session_timeout, idle_timeout, shared_users,
	       keepalive_timeout, login_by, is_local, created_at, updated_at
`

func (r *ProfileRepository) Create(ctx context.Context, p *models.HotspotProfile) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO hotspot_profiles
		  (server_id, name, rate_limit, session_timeout, idle_timeout, shared_users,
		   keepalive_timeout, login_by, is_local)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, created_at, updated_at`,
		p.ServerID, p.Name, p.RateLimit, p.SessionTimeout, p.IdleTimeout, p.SharedUsers,
		p.KeepaliveTimeout, p.LoginBy, p.IsLocal,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *ProfileRepository) ListByServer(ctx context.Context, serverID uuid.UUID) ([]models.HotspotProfile, error) {
	rows, err := r.pool.Query(ctx, profileSelect+` FROM hotspot_profiles WHERE server_id=$1 ORDER BY name`, serverID)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	defer rows.Close()
	var out []models.HotspotProfile
	for rows.Next() {
		var p models.HotspotProfile
		if err := rows.Scan(&p.ID, &p.ServerID, &p.Name, &p.RateLimit, &p.SessionTimeout,
			&p.IdleTimeout, &p.SharedUsers, &p.KeepaliveTimeout, &p.LoginBy, &p.IsLocal,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProfileRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.HotspotProfile, error) {
	p := &models.HotspotProfile{}
	err := r.pool.QueryRow(ctx, profileSelect+` FROM hotspot_profiles WHERE id=$1`, id,
	).Scan(&p.ID, &p.ServerID, &p.Name, &p.RateLimit, &p.SessionTimeout, &p.IdleTimeout,
		&p.SharedUsers, &p.KeepaliveTimeout, &p.LoginBy, &p.IsLocal,
		&p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find profile: %w", err)
	}
	return p, nil
}

func (r *ProfileRepository) FindByName(ctx context.Context, serverID uuid.UUID, name string) (*models.HotspotProfile, error) {
	p := &models.HotspotProfile{}
	err := r.pool.QueryRow(ctx, profileSelect+` FROM hotspot_profiles WHERE server_id=$1 AND name=$2`, serverID, name,
	).Scan(&p.ID, &p.ServerID, &p.Name, &p.RateLimit, &p.SessionTimeout, &p.IdleTimeout,
		&p.SharedUsers, &p.KeepaliveTimeout, &p.LoginBy, &p.IsLocal,
		&p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find profile by name: %w", err)
	}
	return p, nil
}

func (r *ProfileRepository) UpsertFromRouter(ctx context.Context, p *models.HotspotProfile) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO hotspot_profiles
		  (server_id, name, rate_limit, session_timeout, idle_timeout, shared_users,
		   keepalive_timeout, login_by, is_local, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,false,now())
		ON CONFLICT (server_id, name) DO UPDATE SET
		  rate_limit=EXCLUDED.rate_limit,
		  session_timeout=EXCLUDED.session_timeout,
		  idle_timeout=EXCLUDED.idle_timeout,
		  shared_users=EXCLUDED.shared_users,
		  keepalive_timeout=EXCLUDED.keepalive_timeout,
		  login_by=EXCLUDED.login_by,
		  updated_at=now()`,
		p.ServerID, p.Name, p.RateLimit, p.SessionTimeout, p.IdleTimeout, p.SharedUsers,
		p.KeepaliveTimeout, p.LoginBy)
	return err
}

func (r *ProfileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM hotspot_profiles WHERE id=$1`, id)
	return err
}
