package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestAdminNotificationRepositoryCountNewUserRegistrations(t *testing.T) {
	db := &fakeAdminNotificationDB{count: 3}
	repo := &AdminNotificationRepository{db: db}

	got, err := repo.CountNewUserRegistrations(context.Background(), "admin-1")
	if err != nil {
		t.Fatalf("CountNewUserRegistrations() error = %v", err)
	}
	if got != 3 {
		t.Fatalf("CountNewUserRegistrations() = %d, want 3", got)
	}
	if db.queryRowUserID != "admin-1" {
		t.Fatalf("expected admin user id admin-1, got %s", db.queryRowUserID)
	}
	if !strings.Contains(db.queryRowSQL, "admin_notification_state") || !strings.Contains(db.queryRowSQL, "user_registered") || !strings.Contains(db.queryRowSQL, "target_type = 'user'") {
		t.Fatalf("unexpected count query: %s", db.queryRowSQL)
	}
}

func TestAdminNotificationRepositoryCountNewUserRegistrationsError(t *testing.T) {
	db := &fakeAdminNotificationDB{queryRowErr: errors.New("boom")}
	repo := &AdminNotificationRepository{db: db}

	_, err := repo.CountNewUserRegistrations(context.Background(), "admin-1")
	if err == nil || !strings.Contains(err.Error(), "unable to count new user registrations") {
		t.Fatalf("expected wrapped count error, got %v", err)
	}
}

func TestAdminNotificationRepositoryAcknowledgeUserRegistrations(t *testing.T) {
	db := &fakeAdminNotificationDB{}
	repo := &AdminNotificationRepository{db: db}

	if err := repo.AcknowledgeUserRegistrations(context.Background(), "admin-1"); err != nil {
		t.Fatalf("AcknowledgeUserRegistrations() error = %v", err)
	}
	if db.execUserID != "admin-1" {
		t.Fatalf("expected admin user id admin-1, got %s", db.execUserID)
	}
	if !strings.Contains(db.execSQL, "INSERT INTO admin_notification_state") || !strings.Contains(db.execSQL, "ON CONFLICT (admin_user_id) DO UPDATE") {
		t.Fatalf("unexpected ack query: %s", db.execSQL)
	}
}

func TestAdminNotificationRepositoryAcknowledgeUserRegistrationsError(t *testing.T) {
	db := &fakeAdminNotificationDB{execErr: errors.New("boom")}
	repo := &AdminNotificationRepository{db: db}

	if err := repo.AcknowledgeUserRegistrations(context.Background(), "admin-1"); err == nil || !strings.Contains(err.Error(), "unable to acknowledge new user registrations") {
		t.Fatalf("expected wrapped ack error, got %v", err)
	}
}

type fakeAdminNotificationDB struct {
	queryRowSQL    string
	queryRowUserID string
	queryRowErr    error
	count          int
	execSQL        string
	execUserID     string
	execErr        error
}

type fakeRow struct {
	count int
	err   error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) > 0 {
		switch d := dest[0].(type) {
		case *int:
			*d = r.count
		case *int64:
			*d = int64(r.count)
		default:
			return errors.New("unsupported scan destination")
		}
	}
	return nil
}

func (db *fakeAdminNotificationDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	db.queryRowSQL = sql
	if len(args) > 0 {
		if userID, ok := args[0].(string); ok {
			db.queryRowUserID = userID
		}
	}
	return fakeRow{count: db.count, err: db.queryRowErr}
}

func (db *fakeAdminNotificationDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	db.execSQL = sql
	if len(args) > 0 {
		if userID, ok := args[0].(string); ok {
			db.execUserID = userID
		}
	}
	return pgconn.CommandTag{}, db.execErr
}
