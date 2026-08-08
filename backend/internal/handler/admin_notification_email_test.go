package handler

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/user/kareelio/backend/internal/model"
)

func TestSendAdminNewRegistrationEmailSendsToActiveAdmin(t *testing.T) {
	repo := &fakeAdminLookup{admin: &model.User{ID: "admin-1", Email: "admin@example.com", Language: "fr"}}
	mailer := &fakeAdminMailer{}

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

	sendAdminNewRegistrationEmail(context.Background(), "req-123", repo, mailer, &model.User{ID: "user-1", Email: "user@example.com", DisplayName: "Jane Doe"})

	if !repo.called {
		t.Fatal("expected admin lookup")
	}
	if mailer.requestID != "req-123" || mailer.adminEmail != "admin@example.com" || mailer.registeredEmail != "user@example.com" || mailer.registeredDisplayName != "Jane Doe" || mailer.language != "fr" {
		t.Fatalf("unexpected mailer call: %+v", mailer)
	}

	logs := buf.String()
	for _, want := range []string{
		"event=register.admin_notification.mail_send_queued request_id=req-123 user_id=user-1 admin_user_id=admin-1 language=fr",
		"event=register.admin_notification.mail_send_completed request_id=req-123 user_id=user-1 admin_user_id=admin-1 mail_sent=true",
	} {
		if !strings.Contains(logs, want) {
			t.Fatalf("missing log %q: %s", want, logs)
		}
	}
	for _, leak := range []string{"token=", "password="} {
		if strings.Contains(logs, leak) {
			t.Fatalf("log output leaked %q: %s", leak, logs)
		}
	}
}

func TestSendAdminNewRegistrationEmailSkipsLocalAdminEmail(t *testing.T) {
	repo := &fakeAdminLookup{admin: &model.User{ID: "admin-1", Email: "admin@kareelio.local", Language: "en"}}
	mailer := &fakeAdminMailer{}

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

	sendAdminNewRegistrationEmail(context.Background(), "req-456", repo, mailer, &model.User{ID: "user-2", Email: "user@example.com", DisplayName: "Jane Doe"})

	if mailer.called {
		t.Fatal("expected mailer to be skipped")
	}
	if !strings.Contains(buf.String(), "reason=local_admin_email") {
		t.Fatalf("expected skip log, got %s", buf.String())
	}
}

func TestSendAdminNewRegistrationEmailSkipsInvalidAdminEmail(t *testing.T) {
	repo := &fakeAdminLookup{admin: &model.User{ID: "admin-1", Email: "not-an-email", Language: "en"}}
	mailer := &fakeAdminMailer{}

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

	sendAdminNewRegistrationEmail(context.Background(), "req-789", repo, mailer, &model.User{ID: "user-3", Email: "user@example.com", DisplayName: "Jane Doe"})

	if mailer.called {
		t.Fatal("expected mailer to be skipped")
	}
	if !strings.Contains(buf.String(), "reason=invalid_admin_email") {
		t.Fatalf("expected invalid-email skip log, got %s", buf.String())
	}
}

func TestSendAdminNewRegistrationEmailLogsMailerError(t *testing.T) {
	repo := &fakeAdminLookup{admin: &model.User{ID: "admin-1", Email: "admin@example.com", Language: "en"}}
	mailer := &fakeAdminMailer{err: errors.New("smtp broken")}

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

	sendAdminNewRegistrationEmail(context.Background(), "req-789", repo, mailer, &model.User{ID: "user-3", Email: "user@example.com", DisplayName: "Jane Doe"})

	logs := buf.String()
	if !strings.Contains(logs, "event=register.admin_notification.error request_id=req-789 user_id=user-3 admin_user_id=admin-1 stage=mail_send error=smtp broken") {
		t.Fatalf("missing error log: %s", logs)
	}
}

func TestSendAdminNewRegistrationEmailAsyncUsesTimeoutContext(t *testing.T) {
	repo := &fakeAdminLookupDeadline{ready: make(chan struct{}, 1)}
	mailer := &blockingAdminMailer{started: make(chan struct{}, 1), release: make(chan struct{}), done: make(chan struct{})}
	registered := &model.User{ID: "user-4", Email: "user4@example.com", DisplayName: "Jane Doe"}

	sendAdminNewRegistrationEmailAsync("req-123", repo, mailer, registered)

	select {
	case <-repo.ready:
	case <-time.After(time.Second):
		t.Fatal("admin lookup was not invoked")
	}

	if repo.ctx == nil {
		t.Fatal("expected context to be captured")
	}
	if _, ok := repo.ctx.Deadline(); !ok {
		t.Fatal("expected async admin lookup context to have a deadline")
	}

	select {
	case <-mailer.started:
	case <-time.After(time.Second):
		t.Fatal("admin mailer was not invoked")
	}

	close(mailer.release)
	select {
	case <-mailer.done:
	case <-time.After(time.Second):
		t.Fatal("async admin notification did not finish")
	}

	if err := repo.ctx.Err(); err != context.Canceled {
		t.Fatalf("expected context to be canceled, got %v", err)
	}
}

type fakeAdminLookup struct {
	admin  *model.User
	err    error
	called bool
}

func (f *fakeAdminLookup) GetActiveAdmin(ctx context.Context) (*model.User, error) {
	f.called = true
	return f.admin, f.err
}

type fakeAdminLookupDeadline struct {
	ctx   context.Context
	ready chan struct{}
}

func (f *fakeAdminLookupDeadline) GetActiveAdmin(ctx context.Context) (*model.User, error) {
	f.ctx = ctx
	select {
	case f.ready <- struct{}{}:
	default:
	}
	return &model.User{ID: "admin-1", Email: "admin@example.com", Language: "en"}, nil
}

type blockingAdminMailer struct {
	started chan struct{}
	release chan struct{}
	done    chan struct{}
}

func (m *blockingAdminMailer) SendAdminNewRegistrationEmail(requestID, adminEmail, registeredUserEmail, registeredDisplayName, language string) error {
	m.started <- struct{}{}
	<-m.release
	defer close(m.done)
	return nil
}

type fakeAdminMailer struct {
	requestID             string
	adminEmail            string
	registeredEmail       string
	registeredDisplayName string
	language              string
	err                   error
	called                bool
}

func (f *fakeAdminMailer) SendAdminNewRegistrationEmail(requestID, adminEmail, registeredUserEmail, registeredDisplayName, language string) error {
	f.called = true
	f.requestID = requestID
	f.adminEmail = adminEmail
	f.registeredEmail = registeredUserEmail
	f.registeredDisplayName = registeredDisplayName
	f.language = language
	return f.err
}
