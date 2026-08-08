package handler

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/user/kareelio/backend/internal/middleware"
	"github.com/user/kareelio/backend/internal/model"
)

func TestAdminResendVerificationSuccess(t *testing.T) {
	userRepo := &fakeUserLookup{user: &model.User{ID: "user-1", Email: "user@example.com", Role: model.RoleUser, Language: "fr"}}
	evRepo := &fakeTokenStore{}
	mailer := &fakeResendMailer{}
	auditRepo := &fakeAuditRepo{}

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

	req := newAdminResendRequest("req-123", "user-1", &model.User{ID: "admin-1", Email: "admin@example.com", Role: model.RoleAdmin})
	status, body := adminResendVerification(req, userRepo, evRepo, mailer, auditRepo, 24)

	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if body["message"] != "verification email sent" {
		t.Fatalf("unexpected body: %#v", body)
	}
	if userRepo.gotID != "user-1" {
		t.Fatalf("unexpected user lookup id: %s", userRepo.gotID)
	}
	if evRepo.deletedUserID != "user-1" || evRepo.createdUserID != "user-1" {
		t.Fatalf("unexpected token store calls: %+v", evRepo)
	}
	if mailer.requestID != "req-123" || mailer.to != "user@example.com" || mailer.token == "" || mailer.language != "fr" {
		t.Fatalf("unexpected mailer calls: %+v", mailer)
	}
	if auditRepo.event == nil || auditRepo.event.Action != model.AuditActionEmailVerificationResent || auditRepo.event.TargetID != "user-1" {
		t.Fatalf("unexpected audit event: %+v", auditRepo.event)
	}
	if !strings.Contains(string(auditRepo.event.Metadata), "user@example.com") {
		t.Fatalf("expected target email metadata, got %s", string(auditRepo.event.Metadata))
	}
	if strings.Contains(buf.String(), mailer.token) {
		t.Fatalf("log output leaked token: %s", buf.String())
	}
}

func TestAdminResendVerificationAlreadyVerified(t *testing.T) {
	verifiedAt := time.Now()
	userRepo := &fakeUserLookup{user: &model.User{ID: "user-1", Email: "user@example.com", Role: model.RoleUser, EmailVerifiedAt: &verifiedAt}}
	evRepo := &fakeTokenStore{}
	mailer := &fakeResendMailer{}
	auditRepo := &fakeAuditRepo{}
	req := newAdminResendRequest("req-123", "user-1", &model.User{ID: "admin-1", Email: "admin@example.com", Role: model.RoleAdmin})

	status, body := adminResendVerification(req, userRepo, evRepo, mailer, auditRepo, 24)
	if status != 409 {
		t.Fatalf("expected 409, got %d", status)
	}
	if body["error"] != "user already verified" {
		t.Fatalf("unexpected body: %#v", body)
	}
	if evRepo.deletedUserID != "" || evRepo.createdUserID != "" || mailer.requestID != "" || auditRepo.event != nil {
		t.Fatalf("expected no side effects, got evRepo=%+v mailer=%+v audit=%+v", evRepo, mailer, auditRepo.event)
	}
}

func TestAdminResendVerificationMissingUser(t *testing.T) {
	userRepo := &fakeUserLookup{err: context.Canceled}
	evRepo := &fakeTokenStore{}
	mailer := &fakeResendMailer{}
	auditRepo := &fakeAuditRepo{}
	req := newAdminResendRequest("req-123", "missing-user", &model.User{ID: "admin-1", Email: "admin@example.com", Role: model.RoleAdmin})

	status, body := adminResendVerification(req, userRepo, evRepo, mailer, auditRepo, 24)
	if status != 404 {
		t.Fatalf("expected 404, got %d", status)
	}
	if body["error"] != "user not found" {
		t.Fatalf("unexpected body: %#v", body)
	}
	if evRepo.deletedUserID != "" || mailer.requestID != "" || auditRepo.event != nil {
		t.Fatalf("expected no side effects, got evRepo=%+v mailer=%+v audit=%+v", evRepo, mailer, auditRepo.event)
	}
}

func TestAdminResendVerificationMailError(t *testing.T) {
	userRepo := &fakeUserLookup{user: &model.User{ID: "user-1", Email: "user@example.com", Role: model.RoleUser}}
	evRepo := &fakeTokenStore{}
	mailer := &fakeResendMailer{err: context.DeadlineExceeded}
	auditRepo := &fakeAuditRepo{}
	req := newAdminResendRequest("req-123", "user-1", &model.User{ID: "admin-1", Email: "admin@example.com", Role: model.RoleAdmin})

	status, body := adminResendVerification(req, userRepo, evRepo, mailer, auditRepo, 24)
	if status != 502 {
		t.Fatalf("expected 502, got %d", status)
	}
	if body["error"] != "unable to send verification email" {
		t.Fatalf("unexpected body: %#v", body)
	}
	if evRepo.deletedUserID != "user-1" || evRepo.createdUserID != "user-1" {
		t.Fatalf("expected token reset before send error, got %+v", evRepo)
	}
	if auditRepo.event != nil {
		t.Fatalf("expected no audit on send error, got %+v", auditRepo.event)
	}
}

type fakeUserLookup struct {
	user  *model.User
	err   error
	gotID string
}

func (f *fakeUserLookup) GetByID(ctx context.Context, id string) (*model.User, error) {
	f.gotID = id
	return f.user, f.err
}

type fakeTokenStore struct {
	deletedUserID string
	createdUserID string
	createHash    string
	createExpiry  string
	deleteErr     error
	createErr     error
}

func (f *fakeTokenStore) DeleteForUser(ctx context.Context, userID string) error {
	f.deletedUserID = userID
	return f.deleteErr
}

func (f *fakeTokenStore) Create(ctx context.Context, userID, tokenHash, expiresAt string) error {
	f.createdUserID = userID
	f.createHash = tokenHash
	f.createExpiry = expiresAt
	return f.createErr
}

type fakeResendMailer struct {
	requestID string
	to        string
	token     string
	language  string
	err       error
}

func (f *fakeResendMailer) SendVerificationEmail(requestID, to, token, language string) error {
	f.requestID = requestID
	f.to = to
	f.token = token
	f.language = language
	return f.err
}

type fakeAuditRepo struct {
	event *model.AuditEvent
}

func (f *fakeAuditRepo) Log(ctx context.Context, event *model.AuditEvent) error {
	f.event = event
	return nil
}

func newAdminResendRequest(requestID, targetID string, actor *model.User) *http.Request {
	req := httptest.NewRequest("POST", "/api/users/"+targetID+"/resend-verification", nil)
	req.Header.Set("X-Request-ID", requestID)
	req.RemoteAddr = "127.0.0.1:1234"
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", targetID)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, middleware.UserKey, actor)
	return req.WithContext(ctx)
}
