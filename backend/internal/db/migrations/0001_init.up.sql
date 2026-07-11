-- 0001_init.up.sql
-- Skema awal DRP-MikREST

-- users aplikasi
CREATE TABLE users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email         TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  role          TEXT NOT NULL DEFAULT 'admin'
                CHECK (role IN ('admin','operator')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- API token per user (untuk integrasi sistem lain)
CREATE TABLE api_tokens (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  label         TEXT NOT NULL,
  token_hash    TEXT NOT NULL UNIQUE,
  token_prefix  TEXT NOT NULL,
  scopes        TEXT[] NOT NULL DEFAULT '{vouchers:rw,servers:ro}',
  rate_limit    INT  NOT NULL DEFAULT 60,
  last_used_at  TIMESTAMPTZ,
  expires_at    TIMESTAMPTZ,
  revoked_at    TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_api_tokens_user_id ON api_tokens(user_id);
CREATE INDEX idx_api_tokens_hash    ON api_tokens(token_hash) WHERE revoked_at IS NULL;

-- server router yang dikelola
CREATE TABLE servers (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name            TEXT NOT NULL,
  host            TEXT NOT NULL,
  api_port        INT  NOT NULL DEFAULT 8728,
  username        TEXT NOT NULL,
  password_enc    TEXT NOT NULL,                -- base64(AES-256-GCM ciphertext)
  status          TEXT NOT NULL DEFAULT 'unknown'
                  CHECK (status IN ('online','offline','unknown')),
  last_checked_at TIMESTAMPTZ,
  created_by      UUID REFERENCES users(id),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- profile voucher
CREATE TABLE hotspot_profiles (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  server_id        UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
  name             TEXT NOT NULL,
  rate_limit       TEXT,
  session_timeout  TEXT,
  idle_timeout     TEXT,
  shared_users     INT  NOT NULL DEFAULT 1,
  keepalive_timeout TEXT,
  login_by         TEXT[],
  comment          TEXT,
  is_local         BOOL NOT NULL DEFAULT false,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (server_id, name)
);

-- voucher / hotspot user
CREATE TABLE vouchers (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  server_id       UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
  profile_id      UUID REFERENCES hotspot_profiles(id) ON DELETE SET NULL,
  batch_id        UUID,
  username        TEXT NOT NULL,
  password        TEXT NOT NULL,
  comment         TEXT,
  status          TEXT NOT NULL DEFAULT 'active'
                  CHECK (status IN ('active','used','disabled','expired','failed')),
  uptime          TEXT,
  bytes_in        BIGINT NOT NULL DEFAULT 0,
  bytes_out       BIGINT NOT NULL DEFAULT 0,
  used_at         TIMESTAMPTZ,
  disabled_at     TIMESTAMPTZ,
  routeros_id     TEXT,
  created_by      UUID REFERENCES users(id),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (server_id, username)
);

-- batch generate voucher
CREATE TABLE voucher_batches (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  server_id       UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
  profile_id      UUID REFERENCES hotspot_profiles(id) ON DELETE SET NULL,
  count           INT NOT NULL,
  pattern         TEXT,
  prefix          TEXT,
  username_mode   TEXT NOT NULL DEFAULT 'random'
                  CHECK (username_mode IN ('random','prefix','same')),
  created_by      UUID REFERENCES users(id),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- foreign key dari vouchers ke voucher_batches (setelah tabelnya ada)
ALTER TABLE vouchers
  ADD CONSTRAINT vouchers_batch_id_fkey
  FOREIGN KEY (batch_id) REFERENCES voucher_batches(id) ON DELETE SET NULL;

-- audit log
CREATE TABLE audit_logs (
  id         BIGSERIAL PRIMARY KEY,
  user_id    UUID REFERENCES users(id),
  server_id  UUID REFERENCES servers(id),
  action     TEXT NOT NULL,
  target     TEXT,
  detail     JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vouchers_server_status   ON vouchers(server_id, status);
CREATE INDEX idx_vouchers_created_at      ON vouchers(created_at DESC);
CREATE INDEX idx_vouchers_batch_id        ON vouchers(batch_id);
CREATE INDEX idx_audit_logs_created_at    ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_user_id       ON audit_logs(user_id);
