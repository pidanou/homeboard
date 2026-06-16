ALTER TABLE tasks RENAME COLUMN end_date TO due_date;
ALTER TABLE tasks DROP COLUMN start_date;
