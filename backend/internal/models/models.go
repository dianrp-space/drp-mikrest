package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type APIToken struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Label       string     `json:"label"`
	TokenHash   string     `json:"-"`
	TokenPrefix string     `json:"token_prefix"`
	Scopes      []string   `json:"scopes"`
	RateLimit   int        `json:"rate_limit"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Server struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Host          string     `json:"host"`
	APIPort       int        `json:"api_port"`
	Username      string     `json:"username"`
	PasswordEnc   string     `json:"-"` // base64 ciphertext
	Status        string     `json:"status"`
	LastCheckedAt *time.Time `json:"last_checked_at,omitempty"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type HotspotProfile struct {
	ID               uuid.UUID  `json:"id"`
	ServerID         uuid.UUID  `json:"server_id"`
	Name             string     `json:"name"`
	RateLimit        string     `json:"rate_limit"`
	SessionTimeout   string     `json:"session_timeout"`
	IdleTimeout      string     `json:"idle_timeout"`
	SharedUsers      int        `json:"shared_users"`
	KeepaliveTimeout string     `json:"keepalive_timeout"`
	LoginBy          []string   `json:"login_by"`
	IsLocal          bool       `json:"is_local"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Voucher struct {
	ID          uuid.UUID  `json:"id"`
	ServerID    uuid.UUID  `json:"server_id"`
	ProfileID   *uuid.UUID `json:"profile_id,omitempty"`
	BatchID     *uuid.UUID `json:"batch_id,omitempty"`
	Username    string     `json:"username"`
	Password    string     `json:"password"`
	Comment     string     `json:"comment"`
	Status      string     `json:"status"`
	Uptime      string     `json:"uptime"`
	BytesIn     int64      `json:"bytes_in"`
	BytesOut    int64      `json:"bytes_out"`
	LimitUptime string     `json:"limit_uptime"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	UsedAt      *time.Time `json:"used_at,omitempty"`
	DisabledAt  *time.Time `json:"disabled_at,omitempty"`
	RouterOSID  string     `json:"routeros_id,omitempty"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type VoucherBatch struct {
	ID           uuid.UUID  `json:"id"`
	ServerID     uuid.UUID  `json:"server_id"`
	ProfileID    *uuid.UUID `json:"profile_id,omitempty"`
	Count        int        `json:"count"`
	Pattern      string     `json:"pattern"`
	Prefix       string     `json:"prefix"`
	UsernameMode string     `json:"username_mode"`
	CreatedBy    *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuditLog struct {
	ID        int64     `json:"id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	ServerID  *uuid.UUID `json:"server_id,omitempty"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Detail    []byte    `json:"detail,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
