CREATE TABLE events (
    id          TEXT PRIMARY KEY,
    family_id   TEXT NOT NULL REFERENCES families(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    description TEXT,
    start_at    TIMESTAMPTZ NOT NULL,
    end_at      TIMESTAMPTZ NOT NULL,
    all_day     BOOLEAN NOT NULL DEFAULT FALSE,
    created_by  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
