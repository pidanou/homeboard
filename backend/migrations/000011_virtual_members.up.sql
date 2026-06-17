CREATE TABLE virtual_members (
    id             TEXT PRIMARY KEY,
    family_id      TEXT NOT NULL REFERENCES families(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    linked_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Allow virtual member IDs (not in users table) to be stored in tasks and events
ALTER TABLE tasks DROP CONSTRAINT tasks_assigned_to_fkey;
ALTER TABLE event_attendees DROP CONSTRAINT event_attendees_user_id_fkey;
