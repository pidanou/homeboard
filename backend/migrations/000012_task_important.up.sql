ALTER TABLE tasks DROP COLUMN priority;
ALTER TABLE tasks ADD COLUMN important boolean NOT NULL DEFAULT false;
