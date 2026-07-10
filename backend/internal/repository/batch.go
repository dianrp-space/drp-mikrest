package repository

import (
	"context"

	"github.com/drp-mikrest/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BatchRepository struct{ pool *pgxpool.Pool }

func NewBatchRepository(pool *pgxpool.Pool) *BatchRepository {
	return &BatchRepository{pool: pool}
}

func (r *BatchRepository) Create(ctx context.Context, b *models.VoucherBatch) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO voucher_batches (server_id, profile_id, count, pattern, prefix, username_mode, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, created_at`,
		b.ServerID, b.ProfileID, b.Count, b.Pattern, b.Prefix, b.UsernameMode, b.CreatedBy,
	).Scan(&b.ID, &b.CreatedAt)
}

func (r *BatchRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.VoucherBatch, error) {
	b := &models.VoucherBatch{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, server_id, profile_id, count, pattern, prefix, username_mode, created_by, created_at
		FROM voucher_batches WHERE id=$1`, id,
	).Scan(&b.ID, &b.ServerID, &b.ProfileID, &b.Count, &b.Pattern, &b.Prefix,
		&b.UsernameMode, &b.CreatedBy, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BatchRepository) ListByServer(ctx context.Context, serverID uuid.UUID, limit int) ([]models.VoucherBatch, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, server_id, profile_id, count, pattern, prefix, username_mode, created_by, created_at
		FROM voucher_batches WHERE server_id=$1 ORDER BY created_at DESC LIMIT $2`, serverID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.VoucherBatch
	for rows.Next() {
		var b models.VoucherBatch
		if err := rows.Scan(&b.ID, &b.ServerID, &b.ProfileID, &b.Count, &b.Pattern,
			&b.Prefix, &b.UsernameMode, &b.CreatedBy, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
