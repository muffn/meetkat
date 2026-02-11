ALTER TABLE polls ADD COLUMN admin_id TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX idx_polls_admin_id ON polls(admin_id) WHERE admin_id != '';
