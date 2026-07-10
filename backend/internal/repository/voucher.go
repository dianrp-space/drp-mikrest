package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/drp-mikrest/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VoucherRepository struct{ pool *pgxpool.Pool }

func NewVoucherRepository(pool *pgxpool.Pool) *VoucherRepository {
	return &VoucherRepository{pool: pool}
}

// Field-list SELECT untuk voucher (semua query pakai ini).
const voucherSelect = `
	SELECT id, server_id, profile_id, batch_id, username, password,
	       COALESCE(comment,'') AS comment, status,
	       COALESCE(uptime,'') AS uptime, bytes_in, bytes_out, used_at, disabled_at,
	       COALESCE(routeros_id,'') AS routeros_id,
	       COALESCE(limit_uptime,'') AS limit_uptime, expires_at,
	       created_by, created_at, updated_at
`

func (r *VoucherRepository) Create(ctx context.Context, v *models.Voucher) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO vouchers
		  (server_id, profile_id, batch_id, username, password, comment, status,
		   routeros_id, limit_uptime, expires_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, created_at, updated_at`,
		v.ServerID, v.ProfileID, v.BatchID, v.Username, v.Password, v.Comment, v.Status,
		v.RouterOSID, v.LimitUptime, v.ExpiresAt, v.CreatedBy,
	).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
}

// ExistsByUsername cek apakah username sudah dipakai untuk server ini.
func (r *VoucherRepository) ExistsByUsername(ctx context.Context, serverID uuid.UUID, username string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM vouchers WHERE server_id=$1 AND username=$2)`,
		serverID, username).Scan(&exists)
	return exists, err
}

// FindByUsername mencari voucher berdasarkan server + username.
func (r *VoucherRepository) FindByUsername(ctx context.Context, serverID uuid.UUID, username string) (*models.Voucher, error) {
	v := &models.Voucher{}
	err := r.pool.QueryRow(ctx, voucherSelect+` FROM vouchers WHERE server_id=$1 AND username=$2`, serverID, username,
	).Scan(&v.ID, &v.ServerID, &v.ProfileID, &v.BatchID, &v.Username, &v.Password,
		&v.Comment, &v.Status, &v.Uptime, &v.BytesIn, &v.BytesOut, &v.UsedAt,
		&v.DisabledAt, &v.RouterOSID, &v.LimitUptime, &v.ExpiresAt,
		&v.CreatedBy, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find voucher by username: %w", err)
	}
	return v, nil
}

func (r *VoucherRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Voucher, error) {
	v := &models.Voucher{}
	err := r.pool.QueryRow(ctx, voucherSelect+` FROM vouchers WHERE id=$1`, id,
	).Scan(&v.ID, &v.ServerID, &v.ProfileID, &v.BatchID, &v.Username, &v.Password,
		&v.Comment, &v.Status, &v.Uptime, &v.BytesIn, &v.BytesOut, &v.UsedAt,
		&v.DisabledAt, &v.RouterOSID, &v.LimitUptime, &v.ExpiresAt,
		&v.CreatedBy, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find voucher: %w", err)
	}
	return v, nil
}

type VoucherFilter struct {
	ServerID *uuid.UUID
	Status   string
	Search   string
	BatchID  *uuid.UUID
	Type     string // "voucher" (username=password), "member" (username!=password), "" (semua)
	Limit    int
	Offset   int
}

func (r *VoucherRepository) List(ctx context.Context, f VoucherFilter) ([]models.Voucher, int, error) {
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 50
	}
	q := voucherSelect + ` FROM vouchers WHERE 1=1`
	countQ := `SELECT count(*) FROM vouchers WHERE 1=1`
	args := []any{}
	i := 1
	addFilter := func(cond string, val any) {
		q += " AND " + cond
		countQ += " AND " + cond
		args = append(args, val)
		i++
	}
	if f.ServerID != nil {
		addFilter(fmt.Sprintf("server_id=$%d", i), *f.ServerID)
	}
	if f.Status != "" {
		addFilter(fmt.Sprintf("status=$%d", i), f.Status)
	}
	if f.Search != "" {
		addFilter(fmt.Sprintf("(username ILIKE $%d OR comment ILIKE $%d)", i, i), "%"+f.Search+"%")
	}
	if f.BatchID != nil {
		addFilter(fmt.Sprintf("batch_id=$%d", i), *f.BatchID)
	}
	if f.Type == "voucher" {
		q += " AND username = password"
		countQ += " AND username = password"
	} else if f.Type == "member" {
		q += " AND username <> password"
		countQ += " AND username <> password"
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count vouchers: %w", err)
	}

	q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list vouchers: %w", err)
	}
	defer rows.Close()
	var out []models.Voucher
	for rows.Next() {
		var v models.Voucher
		if err := rows.Scan(&v.ID, &v.ServerID, &v.ProfileID, &v.BatchID, &v.Username,
			&v.Password, &v.Comment, &v.Status, &v.Uptime, &v.BytesIn, &v.BytesOut,
			&v.UsedAt, &v.DisabledAt, &v.RouterOSID, &v.LimitUptime, &v.ExpiresAt,
			&v.CreatedBy, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}

func (r *VoucherRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE vouchers SET status=$1, updated_at=now() WHERE id=$2`, status, id)
	return err
}

func (r *VoucherRepository) UpdateRouterOSID(ctx context.Context, id uuid.UUID, routerOSID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE vouchers SET routeros_id=$1, updated_at=now() WHERE id=$2`, routerOSID, id)
	return err
}

func (r *VoucherRepository) UpdateExpiresAt(ctx context.Context, id uuid.UUID, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE vouchers SET expires_at=$1, updated_at=now() WHERE id=$2`, expiresAt, id)
	return err
}

func (r *VoucherRepository) MarkFailed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE vouchers SET status='failed', updated_at=now() WHERE id=$1`, id)
	return err
}

func (r *VoucherRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM vouchers WHERE id=$1`, id)
	return err
}

func (r *VoucherRepository) ListByServer(ctx context.Context, serverID uuid.UUID, limit int) ([]models.Voucher, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, voucherSelect+` FROM vouchers WHERE server_id=$1 ORDER BY created_at DESC LIMIT $2`, serverID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Voucher
	for rows.Next() {
		var v models.Voucher
		if err := rows.Scan(&v.ID, &v.ServerID, &v.ProfileID, &v.BatchID, &v.Username,
			&v.Password, &v.Comment, &v.Status, &v.Uptime, &v.BytesIn, &v.BytesOut,
			&v.UsedAt, &v.DisabledAt, &v.RouterOSID, &v.LimitUptime, &v.ExpiresAt,
			&v.CreatedBy, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// ListExpired mengembalikan voucher dengan expires_at < now dan status active.
// Digunakan scheduler untuk auto-kick & auto-disable.
func (r *VoucherRepository) ListExpired(ctx context.Context, limit int) ([]models.Voucher, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.pool.Query(ctx, voucherSelect+`
		FROM vouchers
		WHERE status = 'active' AND expires_at IS NOT NULL AND expires_at < now()
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Voucher
	for rows.Next() {
		var v models.Voucher
		if err := rows.Scan(&v.ID, &v.ServerID, &v.ProfileID, &v.BatchID, &v.Username,
			&v.Password, &v.Comment, &v.Status, &v.Uptime, &v.BytesIn, &v.BytesOut,
			&v.UsedAt, &v.DisabledAt, &v.RouterOSID, &v.LimitUptime, &v.ExpiresAt,
			&v.CreatedBy, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// MarkFirstUse menandai voucher sudah pernah dipakai (first login) dan
// set expires_at = used_at + limit_uptime. Atomic, hanya update jika used_at IS NULL
// (idempotent — tidak menimpa first-use time yang sudah ada).
func (r *VoucherRepository) MarkFirstUse(ctx context.Context, id uuid.UUID, usedAt time.Time, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE vouchers
		SET used_at = $2, expires_at = $3, updated_at = now()
		WHERE id = $1 AND used_at IS NULL`,
		id, usedAt, expiresAt)
	return err
}

// UpdateUsage mengupdate uptime & bytes dari router (tanpa mengubah first-use time).
func (r *VoucherRepository) UpdateUsage(ctx context.Context, id uuid.UUID, uptime string, bytesIn, bytesOut int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE vouchers
		SET uptime = $2, bytes_in = $3, bytes_out = $4, updated_at = now()
		WHERE id = $1`,
		id, uptime, bytesIn, bytesOut)
	return err
}

// ListExpiredForCleanup mengembalikan voucher dengan status 'expired' yang sudah expired
// lebih dari cleanupDelay (mis. 24 jam). Ini untuk cron cleanup hapus dari router & DB.
func (r *VoucherRepository) ListExpiredForCleanup(ctx context.Context, cleanupDelay time.Duration, limit int) ([]models.Voucher, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.pool.Query(ctx, voucherSelect+`
		FROM vouchers
		WHERE status = 'expired' AND updated_at < now() - $1::interval
		LIMIT $2`,
		cleanupDelay.String(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Voucher
	for rows.Next() {
		var v models.Voucher
		if err := rows.Scan(&v.ID, &v.ServerID, &v.ProfileID, &v.BatchID, &v.Username,
			&v.Password, &v.Comment, &v.Status, &v.Uptime, &v.BytesIn, &v.BytesOut,
			&v.UsedAt, &v.DisabledAt, &v.RouterOSID, &v.LimitUptime, &v.ExpiresAt,
			&v.CreatedBy, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
