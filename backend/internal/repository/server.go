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

type ServerRepository struct{ pool *pgxpool.Pool }

func NewServerRepository(pool *pgxpool.Pool) *ServerRepository {
	return &ServerRepository{pool: pool}
}

func (r *ServerRepository) Create(ctx context.Context, s *models.Server) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO servers (name, host, api_port, username, password_enc, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`,
		s.Name, s.Host, s.APIPort, s.Username, s.PasswordEnc, s.Status, s.CreatedBy,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *ServerRepository) List(ctx context.Context) ([]models.Server, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, host, api_port, username, password_enc, status, last_checked_at,
		       created_by, created_at, updated_at
		FROM servers ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list servers: %w", err)
	}
	defer rows.Close()
	var out []models.Server
	for rows.Next() {
		var s models.Server
		if err := rows.Scan(&s.ID, &s.Name, &s.Host, &s.APIPort, &s.Username, &s.PasswordEnc,
			&s.Status, &s.LastCheckedAt, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *ServerRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Server, error) {
	s := &models.Server{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, host, api_port, username, password_enc, status, last_checked_at,
		       created_by, created_at, updated_at
		FROM servers WHERE id = $1`, id,
	).Scan(&s.ID, &s.Name, &s.Host, &s.APIPort, &s.Username, &s.PasswordEnc,
		&s.Status, &s.LastCheckedAt, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find server: %w", err)
	}
	return s, nil
}

func (r *ServerRepository) Update(ctx context.Context, s *models.Server) error {
	return r.pool.QueryRow(ctx, `
		UPDATE servers SET name=$1, host=$2, api_port=$3, username=$4, password_enc=$5,
		                   updated_at=now()
		WHERE id=$6
		RETURNING updated_at`,
		s.Name, s.Host, s.APIPort, s.Username, s.PasswordEnc, s.ID,
	).Scan(&s.UpdatedAt)
}

func (r *ServerRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE servers SET status=$1, last_checked_at=now(), updated_at=now() WHERE id=$2`,
		status, id)
	return err
}

func (r *ServerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM servers WHERE id=$1`, id)
	return err
}
