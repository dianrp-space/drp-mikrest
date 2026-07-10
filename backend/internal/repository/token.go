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

type TokenRepository struct{ pool *pgxpool.Pool }

func NewTokenRepository(pool *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{pool: pool}
}

func (r *TokenRepository) Create(ctx context.Context, t *models.APIToken) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO api_tokens (user_id, label, token_hash, token_prefix, scopes, rate_limit, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, created_at`,
		t.UserID, t.Label, t.TokenHash, t.TokenPrefix, t.Scopes, t.RateLimit, t.ExpiresAt,
	).Scan(&t.ID, &t.CreatedAt)
}

func (r *TokenRepository) FindByHash(ctx context.Context, hash string) (*models.APIToken, error) {
	t := &models.APIToken{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, label, token_hash, token_prefix, scopes, rate_limit,
		       last_used_at, expires_at, revoked_at, created_at
		FROM api_tokens WHERE token_hash=$1 AND revoked_at IS NULL`, hash,
	).Scan(&t.ID, &t.UserID, &t.Label, &t.TokenHash, &t.TokenPrefix, &t.Scopes,
		&t.RateLimit, &t.LastUsedAt, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find token: %w", err)
	}
	return t, nil
}

func (r *TokenRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.APIToken, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, label, token_hash, token_prefix, scopes, rate_limit,
		       last_used_at, expires_at, revoked_at, created_at
		FROM api_tokens WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.APIToken
	for rows.Next() {
		var t models.APIToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.Label, &t.TokenHash, &t.TokenPrefix,
			&t.Scopes, &t.RateLimit, &t.LastUsedAt, &t.ExpiresAt, &t.RevokedAt,
			&t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TokenRepository) Revoke(ctx context.Context, id, userID uuid.UUID) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE api_tokens SET revoked_at=now() WHERE id=$1 AND user_id=$2 AND revoked_at IS NULL`,
		id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("token tidak ditemukan atau sudah di-revoke")
	}
	return nil
}

// HardDelete menghapus baris token dari DB secara permanen.
// Token yang dihapus tidak akan bisa dipakai (hash hilang) dan hilang dari daftar.
func (r *TokenRepository) HardDelete(ctx context.Context, id, userID uuid.UUID) error {
	ct, err := r.pool.Exec(ctx,
		`DELETE FROM api_tokens WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("token tidak ditemukan")
	}
	return nil
}

func (r *TokenRepository) TouchLastUsed(ctx context.Context, id uuid.UUID) {
	_, _ = r.pool.Exec(ctx, `UPDATE api_tokens SET last_used_at=now() WHERE id=$1`, id)
}
