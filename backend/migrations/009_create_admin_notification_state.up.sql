CREATE TABLE IF NOT EXISTS admin_notification_state (
    admin_user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    user_registrations_seen_at TIMESTAMPTZ NOT NULL DEFAULT 'epoch',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
