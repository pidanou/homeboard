CREATE TABLE invites (
    token      TEXT PRIMARY KEY,
    family_id  TEXT NOT NULL REFERENCES families(id) ON DELETE CASCADE,
    created_by TEXT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ
);
