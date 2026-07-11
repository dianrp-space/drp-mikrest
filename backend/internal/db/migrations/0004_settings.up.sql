-- 0004_settings.up.sql
-- Application settings key-value store

CREATE TABLE settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO settings (key, value) VALUES
    ('app_name', 'DRP-MikREST'),
    ('logo_path', ''),
    ('favicon_path', ''),
    ('app_url', 'http://localhost:8080')
ON CONFLICT (key) DO NOTHING;
