DROP TABLE IF EXISTS oidc_identities;
-- Fails if any OIDC-only users exist (password_hash IS NULL) — set one before rolling back.
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
