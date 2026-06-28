ALTER TABLE push_subscriptions ADD COLUMN family_id UUID REFERENCES households(id) ON DELETE CASCADE;
