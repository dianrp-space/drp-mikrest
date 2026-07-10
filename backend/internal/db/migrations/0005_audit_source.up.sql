ALTER TABLE audit_logs ADD COLUMN source TEXT NOT NULL DEFAULT 'web'
  CHECK (source IN ('web','api','system'));
CREATE INDEX idx_audit_logs_source ON audit_logs(source);
