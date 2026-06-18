ALTER TABLE labels RENAME TO categories;
ALTER TABLE tasks ADD COLUMN category_id TEXT REFERENCES categories(id) ON DELETE SET NULL;
ALTER TABLE events ADD COLUMN category_id TEXT REFERENCES categories(id) ON DELETE SET NULL;

UPDATE tasks t SET category_id = (SELECT label_id FROM task_labels WHERE task_id = t.id LIMIT 1);
UPDATE events e SET category_id = (SELECT label_id FROM event_labels WHERE event_id = e.id LIMIT 1);

DROP TABLE task_labels;
DROP TABLE event_labels;
