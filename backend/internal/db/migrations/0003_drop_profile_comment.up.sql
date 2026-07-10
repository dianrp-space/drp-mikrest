-- 0003_drop_profile_comment.up.sql
-- /ip/hotspot/user/profile di RouterOS TIDAK punya property `comment`.
-- Hapus kolom dari DB untuk konsistensi.

ALTER TABLE hotspot_profiles DROP COLUMN IF EXISTS comment;
