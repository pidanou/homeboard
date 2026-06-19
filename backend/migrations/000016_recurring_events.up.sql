ALTER TABLE events
  ADD COLUMN recurrence_rule TEXT,
  ADD COLUMN recurrence_parent_id TEXT REFERENCES events(id) ON DELETE CASCADE,
  ADD COLUMN recurrence_date DATE,
  ADD COLUMN cancelled BOOLEAN NOT NULL DEFAULT FALSE;
