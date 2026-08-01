package handler

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"
	"time"
)

func TestSendVerificationEmailAsyncReturnsBeforeMailerCompletes(t *testing.T) {
	mailer := &blockingVerificationMailer{
		started: make(chan mailCall, 1),
		release: make(chan struct{}),
		done:    make(chan struct{}),
		err:     errors.New("smtp client init: 421 no system resources"),
	}

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

	returned := make(chan struct{})
	go func() {
		sendVerificationEmailAsync("req-123", "register", "user-1", "user@example.com", "secret-token", mailer)
		close(returned)
	}()

	select {
	case <-returned:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("sendVerificationEmailAsync blocked on mail send")
	}

	var call mailCall
	select {
	case call = <-mailer.started:
	case <-time.After(time.Second):
		t.Fatal("mailer was not called")
	}
	if call.requestID != "req-123" || call.to != "user@example.com" || call.token != "secret-token" {
		t.Fatalf("unexpected mail call: %+v", call)
	}

	close(mailer.release)
	select {
	case <-mailer.done:
	case <-time.After(time.Second):
		t.Fatal("mailer did not complete")
	}

	logs := buf.String()
	if !strings.Contains(logs, "event=register.mail_send_queued request_id=req-123 user_id=user-1 email=user@example.com") {
		t.Fatalf("missing queued log: %s", logs)
	}
	if !strings.Contains(logs, "event=register.error request_id=req-123 stage=mail_send user_id=user-1 email=user@example.com error=smtp client init: 421 no system resources") {
		t.Fatalf("missing async error log: %s", logs)
	}
	if !strings.Contains(logs, "event=register.mail_send_completed request_id=req-123 user_id=user-1 email=user@example.com mail_sent=false") {
		t.Fatalf("missing completion log: %s", logs)
	}
	if strings.Contains(logs, "secret-token") {
		t.Fatalf("log output leaked token: %s", logs)
	}
}

func TestSendVerificationEmailAsyncLogsSuccess(t *testing.T) {
	mailer := &blockingVerificationMailer{
		started: make(chan mailCall, 1),
		release: make(chan struct{}),
		done:    make(chan struct{}),
	}

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

	sendVerificationEmailAsync("req-456", "resend_verification", "user-2", "resend@example.com", "secret-token", mailer)

	select {
	case <-mailer.started:
	case <-time.After(time.Second):
		t.Fatal("mailer was not called")
	}
	close(mailer.release)
	select {
	case <-mailer.done:
	case <-time.After(time.Second):
		t.Fatal("mailer did not complete")
	}

	logs := buf.String()
	if !strings.Contains(logs, "event=resend_verification.mail_send_completed request_id=req-456 user_id=user-2 email=resend@example.com mail_sent=true") {
		t.Fatalf("missing success log: %s", logs)
	}
	if strings.Contains(logs, "secret-token") {
		t.Fatalf("log output leaked token: %s", logs)
	}
}

type blockingVerificationMailer struct {
	started chan mailCall
	release chan struct{}
	done    chan struct{}
	err     error
}

type mailCall struct {
	requestID string
	to        string
	token     string
}

func (m *blockingVerificationMailer) SendVerificationEmail(requestID, to, token string) error {
	m.started <- mailCall{requestID: requestID, to: to, token: token}
	<-m.release
	defer close(m.done)
	return m.err
}
