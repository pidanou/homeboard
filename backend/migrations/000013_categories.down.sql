CREATE TABLE task_labels (task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE, label_id TEXT NOT NULL REFERENCES categories(id) ON DELETE CASCADE, PRIMARY KEY (task_id, label_id));
CREATE TABLE event_labels (event_id TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE, label_id TEXT NOT NULL REFERENCES categories(id) ON DELETE CASCADE, PRIMARY KEY (event_id, label_id));
INSERT INTO task_labels (task_id, label_id) SELECT id, category_id FROM tasks WHERE category_id IS NOT NULL;
INSERT INTO event_labels (event_id, label_id) SELECT id, category_id FROM events WHERE category_id IS NOT NULL;
ALTER TABLE events DROP COLUMN category_id;
ALTER TABLE tasks DROP COLUMN category_id;
ALTER TABLE categories RENAME TO labels;
