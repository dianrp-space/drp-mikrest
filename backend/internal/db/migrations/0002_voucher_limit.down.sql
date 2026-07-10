-- 0002_voucher_limit.down.sql
DROP INDEX IF EXISTS idx_vouchers_expires_active;
ALTER TABLE vouchers DROP COLUMN IF EXISTS expires_at;
ALTER TABLE vouchers DROP COLUMN IF EXISTS limit_uptime;
