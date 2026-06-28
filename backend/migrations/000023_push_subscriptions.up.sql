CREATE TABLE push_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    family_id TEXT NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    auth TEXT NOT NULL,
    p256dh TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, endpoint)
);
