-- 0002_voucher_limit.up.sql
-- Tambah field limit-uptime (jangka waktu voucher) & expires_at (kapan voucher expired).
-- limit_uptime mengikuti format RouterOS: "1h", "30m", "1d", "1d12h", dst.
-- expires_at diisi saat generate (created_at + limit_uptime), NULL = tanpa batas.

ALTER TABLE vouchers
  ADD COLUMN limit_uptime TEXT,
  ADD COLUMN expires_at   TIMESTAMPTZ;

CREATE INDEX idx_vouchers_expires_active
  ON vouchers(expires_at)
  WHERE status = 'active' AND expires_at IS NOT NULL;
