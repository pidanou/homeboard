CREATE INDEX idx_tasks_family_id ON tasks(family_id);
CREATE INDEX idx_events_family_id_start_at ON events(family_id, start_at);
CREATE INDEX idx_categories_family_id ON categories(family_id);
