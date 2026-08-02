package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminNotificationRepository struct {
	db adminNotificationDB
}

type adminNotificationDB interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func NewAdminNotificationRepository(db *pgxpool.Pool) *AdminNotificationRepository {
	return &AdminNotificationRepository{db: db}
}

func (r *AdminNotificationRepository) CountNewUserRegistrations(ctx context.Context, adminUserID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM audit_events
		WHERE action = 'user_registered'
		  AND target_type = 'user'
		  AND created_at > COALESCE(
			  (SELECT user_registrations_seen_at
			   FROM admin_notification_state
			   WHERE admin_user_id = $1),
			  'epoch'::timestamptz
		  )`, adminUserID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unable to count new user registrations: %w", err)
	}

	return count, nil
}

func (r *AdminNotificationRepository) AcknowledgeUserRegistrations(ctx context.Context, adminUserID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO admin_notification_state (admin_user_id, user_registrations_seen_at, updated_at)
		VALUES ($1, NOW(), NOW())
		ON CONFLICT (admin_user_id) DO UPDATE
		SET user_registrations_seen_at = EXCLUDED.user_registrations_seen_at,
		    updated_at = NOW()`, adminUserID)
	if err != nil {
		return fmt.Errorf("unable to acknowledge new user registrations: %w", err)
	}

	return nil
}
