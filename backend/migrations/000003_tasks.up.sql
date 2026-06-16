CREATE TABLE tasks (
    id          TEXT PRIMARY KEY,
    family_id   TEXT NOT NULL REFERENCES families(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'todo',
    assigned_to TEXT REFERENCES users(id) ON DELETE SET NULL,
    due_date    TIMESTAMPTZ,
    created_by  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
