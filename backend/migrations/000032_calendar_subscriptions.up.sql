CREATE TABLE calendar_subscriptions (
    id              TEXT PRIMARY KEY,
    family_id       TEXT NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    url             TEXT NOT NULL,
    created_by      TEXT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_synced_at  TIMESTAMPTZ,
    last_sync_error TEXT
);

ALTER TABLE events
  ADD COLUMN subscription_id TEXT REFERENCES calendar_subscriptions(id) ON DELETE CASCADE,
  ADD COLUMN external_uid TEXT;

CREATE UNIQUE INDEX events_subscription_external_uid_uidx
  ON events (subscription_id, external_uid)
  WHERE subscription_id IS NOT NULL;
