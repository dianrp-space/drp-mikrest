package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// contextKey untuk auth source (web/api/system).
type ctxKey string

const authSourceKey ctxKey = "auth_source"

// WithAuthSource membungkus context dengan source.
func WithAuthSource(ctx context.Context, source string) context.Context {
	return context.WithValue(ctx, authSourceKey, source)
}

// AuthSource mengambil source dari context, default "system".
func AuthSource(ctx context.Context) string {
	v, _ := ctx.Value(authSourceKey).(string)
	if v == "" {
		return "system"
	}
	return v
}

type AuditRepository struct{ pool *pgxpool.Pool }

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

func (r *AuditRepository) Log(ctx context.Context, userID *uuid.UUID, serverID *uuid.UUID, action, target string, detail []byte) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs (user_id, server_id, action, target, detail, source)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		userID, serverID, action, target, detail, AuthSource(ctx))
	return err
}

type AuditFilter struct {
	Action    string
	ServerID  *uuid.UUID
	UserID    *uuid.UUID
	DateFrom  *time.Time
	DateTo    *time.Time
	Search    string
	Limit     int
	Offset    int
}

type AuditRow struct {
	ID         int64      `json:"id"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	ServerID   *uuid.UUID `json:"server_id,omitempty"`
	Action     string     `json:"action"`
	Target     string     `json:"target"`
	Detail     *string    `json:"detail,omitempty"`
	Source     string     `json:"source"`
	CreatedAt  time.Time  `json:"created_at"`
	UserEmail  string     `json:"user_email"`
	ServerName string     `json:"server_name"`
}

const auditSelect = `
	SELECT a.id, a.user_id, a.server_id, a.action, a.target, a.detail, a.source, a.created_at,
	       COALESCE(u.email, '') AS user_email,
	       COALESCE(s.name, '') AS server_name
`

func (r *AuditRepository) List(ctx context.Context, f AuditFilter) ([]AuditRow, int, error) {
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 50
	}

	q := auditSelect + ` FROM audit_logs a
		LEFT JOIN users u ON u.id = a.user_id
		LEFT JOIN servers s ON s.id = a.server_id
		WHERE 1=1`
	countQ := `SELECT count(*) FROM audit_logs a WHERE 1=1`
	args := []any{}
	i := 1

	addFilter := func(cond string, val any) {
		q += " AND " + cond
		countQ += " AND " + cond
		args = append(args, val)
		i++
	}

	if f.Action != "" {
		addFilter(fmt.Sprintf("a.action=$%d", i), f.Action)
	}
	if f.ServerID != nil {
		addFilter(fmt.Sprintf("a.server_id=$%d", i), *f.ServerID)
	}
	if f.UserID != nil {
		addFilter(fmt.Sprintf("a.user_id=$%d", i), *f.UserID)
	}
	if f.DateFrom != nil {
		addFilter(fmt.Sprintf("a.created_at>=$%d", i), *f.DateFrom)
	}
	if f.DateTo != nil {
		addFilter(fmt.Sprintf("a.created_at<=$%d", i), *f.DateTo)
	}
	if f.Search != "" {
		addFilter(fmt.Sprintf("(a.action ILIKE $%d OR a.target ILIKE $%d)", i, i), "%"+f.Search+"%")
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit: %w", err)
	}

	q += fmt.Sprintf(" ORDER BY a.created_at DESC LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit: %w", err)
	}
	defer rows.Close()

	var out []AuditRow
	for rows.Next() {
		var row AuditRow
		var detailBytes []byte
		if err := rows.Scan(
			&row.ID, &row.UserID, &row.ServerID, &row.Action, &row.Target,
			&detailBytes, &row.Source, &row.CreatedAt, &row.UserEmail, &row.ServerName,
		); err != nil {
			return nil, 0, err
		}
		if detailBytes != nil {
			s := string(detailBytes)
			row.Detail = &s
		}
		out = append(out, row)
	}
	return out, total, rows.Err()
}

// DeleteOlderThan menghapus audit log yang lebih lama dari durasi tertentu.
func (r *AuditRepository) DeleteOlderThan(ctx context.Context, age time.Duration) (int64, error) {
	res, err := r.pool.Exec(ctx, `
		DELETE FROM audit_logs WHERE created_at < now() - $1::interval`,
		age.String())
	if err != nil {
		return 0, fmt.Errorf("delete old audit logs: %w", err)
	}
	return res.RowsAffected(), nil
}
