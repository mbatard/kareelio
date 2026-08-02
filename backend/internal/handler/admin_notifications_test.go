package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/kareelio/backend/internal/middleware"
	"github.com/user/kareelio/backend/internal/model"
)

func TestAdminNotificationCountSuccess(t *testing.T) {
	repo := &fakeAdminNotificationRepo{count: 4}
	h := &AdminNotificationHandler{notificationRepo: repo}

	var buf bytes.Buffer
	oldFlags := log.Flags()
	oldPrefix := log.Prefix()
	oldOutput := log.Writer()
	t.Cleanup(func() {
		log.SetFlags(oldFlags)
		log.SetPrefix(oldPrefix)
		log.SetOutput(oldOutput)
	})
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(&buf)

	req := newAdminNotificationRequest(http.MethodGet, "/api/admin/notifications", &model.User{ID: "admin-1", Email: "admin@example.com", Role: model.RoleAdmin})
	rec := httptest.NewRecorder()

	h.GetUserRegistrations(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if repo.countAdminID != "admin-1" {
		t.Fatalf("unexpected count admin id: %s", repo.countAdminID)
	}
	var got map[string]int
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got["new_user_registrations"] != 4 {
		t.Fatalf("unexpected body: %#v", got)
	}
	logs := buf.String()
	if !strings.Contains(logs, "event=admin_notifications.count request_id=req-123 admin_user_id=admin-1 new_user_registrations=4") {
		t.Fatalf("missing count log: %s", logs)
	}
	if strings.Contains(logs, "admin@example.com") {
		t.Fatalf("log leaked admin email: %s", logs)
	}
}

func TestAdminNotificationAckSuccess(t *testing.T) {
	repo := &fakeAdminNotificationRepo{}
	h := &AdminNotificationHandler{notificationRepo: repo}

	req := newAdminNotificationRequest(http.MethodPost, "/api/admin/notifications/user-registrations/ack", &model.User{ID: "admin-2", Email: "admin@example.com", Role: model.RoleAdmin})
	rec := httptest.NewRecorder()

	h.AcknowledgeUserRegistrations(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if repo.ackAdminID != "admin-2" {
		t.Fatalf("unexpected ack admin id: %s", repo.ackAdminID)
	}
}

func TestAdminNotificationCountUnauthorized(t *testing.T) {
	h := &AdminNotificationHandler{notificationRepo: &fakeAdminNotificationRepo{count: 1}}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/notifications", nil)
	rec := httptest.NewRecorder()

	h.GetUserRegistrations(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAdminNotificationAckError(t *testing.T) {
	repo := &fakeAdminNotificationRepo{ackErr: context.DeadlineExceeded}
	h := &AdminNotificationHandler{notificationRepo: repo}
	req := newAdminNotificationRequest(http.MethodPost, "/api/admin/notifications/user-registrations/ack", &model.User{ID: "admin-3", Email: "admin@example.com", Role: model.RoleAdmin})
	rec := httptest.NewRecorder()

	h.AcknowledgeUserRegistrations(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if repo.ackAdminID != "admin-3" {
		t.Fatalf("unexpected ack admin id: %s", repo.ackAdminID)
	}
}

type fakeAdminNotificationRepo struct {
	count        int
	countErr     error
	ackErr       error
	countAdminID string
	ackAdminID   string
}

func (f *fakeAdminNotificationRepo) CountNewUserRegistrations(ctx context.Context, adminUserID string) (int, error) {
	f.countAdminID = adminUserID
	return f.count, f.countErr
}

func (f *fakeAdminNotificationRepo) AcknowledgeUserRegistrations(ctx context.Context, adminUserID string) error {
	f.ackAdminID = adminUserID
	return f.ackErr
}

func newAdminNotificationRequest(method, path string, actor *model.User) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("X-Request-ID", "req-123")
	ctx := context.WithValue(req.Context(), middleware.UserKey, actor)
	return req.WithContext(ctx)
}
