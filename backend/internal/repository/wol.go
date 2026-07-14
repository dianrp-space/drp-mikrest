package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/DRP-MikREST/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WOLRepository struct{ pool *pgxpool.Pool }

func NewWOLRepository(pool *pgxpool.Pool) *WOLRepository {
	return &WOLRepository{pool: pool}
}

func (r *WOLRepository) Create(ctx context.Context, t *models.WOLTarget) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO wol_targets (server_id, interface_name, mac_address, name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`,
		t.ServerID, t.InterfaceName, t.MACAddress, t.Name,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *WOLRepository) ListByServer(ctx context.Context, serverID uuid.UUID) ([]models.WOLTarget, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, server_id, interface_name, mac_address, name, created_at, updated_at
		FROM wol_targets WHERE server_id = $1 ORDER BY created_at DESC`, serverID)
	if err != nil {
		return nil, fmt.Errorf("list wol: %w", err)
	}
	defer rows.Close()
	var out []models.WOLTarget
	for rows.Next() {
		var t models.WOLTarget
		if err := rows.Scan(&t.ID, &t.ServerID, &t.InterfaceName, &t.MACAddress, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *WOLRepository) ListAll(ctx context.Context) ([]models.WOLTarget, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, server_id, interface_name, mac_address, name, created_at, updated_at
		FROM wol_targets ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list all wol: %w", err)
	}
	defer rows.Close()
	var out []models.WOLTarget
	for rows.Next() {
		var t models.WOLTarget
		if err := rows.Scan(&t.ID, &t.ServerID, &t.InterfaceName, &t.MACAddress, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *WOLRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.WOLTarget, error) {
	t := &models.WOLTarget{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, server_id, interface_name, mac_address, name, created_at, updated_at
		FROM wol_targets WHERE id = $1`, id,
	).Scan(&t.ID, &t.ServerID, &t.InterfaceName, &t.MACAddress, &t.Name, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find wol: %w", err)
	}
	return t, nil
}

func (r *WOLRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM wol_targets WHERE id = $1`, id)
	return err
}
