CREATE TABLE calendar_export_tokens (
    token      TEXT PRIMARY KEY,
    family_id  TEXT NOT NULL UNIQUE REFERENCES households(id) ON DELETE CASCADE,
    created_by TEXT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
