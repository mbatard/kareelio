ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMPTZ NULL;

UPDATE users SET email_verified_at = created_at WHERE email_verified_at IS NULL AND is_active = true;
