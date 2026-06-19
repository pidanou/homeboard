ALTER TABLE events
  DROP COLUMN cancelled,
  DROP COLUMN recurrence_date,
  DROP COLUMN recurrence_parent_id,
  DROP COLUMN recurrence_rule;
