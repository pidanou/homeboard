ALTER TABLE tasks DROP COLUMN important;
ALTER TABLE tasks ADD COLUMN priority varchar(20) NOT NULL DEFAULT 'medium';
