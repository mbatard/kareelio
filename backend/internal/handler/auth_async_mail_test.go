package handler

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSendVerificationEmailAsyncReturnsBeforeMailerCompletes(t *testing.T) {
	mailer := &blockingVerificationMailer{
		started: make(chan mailCall, 1),
		release: make(chan struct{}),
		err:     errors.New("smtp client init: 421 no system resources"),
	}

	logs := newAsyncMailLogCapture("event=register.mail_send_completed")
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
	log.SetOutput(logs)

	returned := make(chan struct{})
	go func() {
		sendVerificationEmailAsync("req-123", "register", "user-1", "user@example.com", "secret-token", "fr", mailer)
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
	if call.requestID != "req-123" || call.to != "user@example.com" || call.token != "secret-token" || call.language != "fr" {
		t.Fatalf("unexpected mail call: %+v", call)
	}

	close(mailer.release)
	select {
	case <-logs.completed:
	case <-time.After(time.Second):
		t.Fatal("completion log was not written")
	}

	output := logs.String()
	if !strings.Contains(output, "event=register.mail_send_queued request_id=req-123 user_id=user-1 email=user@example.com language=fr") {
		t.Fatalf("missing queued log: %s", output)
	}
	if !strings.Contains(output, "event=register.error request_id=req-123 stage=mail_send user_id=user-1 email=user@example.com language=fr error=smtp client init: 421 no system resources") {
		t.Fatalf("missing async error log: %s", output)
	}
	if !strings.Contains(output, "event=register.mail_send_completed request_id=req-123 user_id=user-1 email=user@example.com language=fr mail_sent=false") {
		t.Fatalf("missing completion log: %s", output)
	}
	if strings.Contains(output, "secret-token") {
		t.Fatalf("log output leaked token: %s", output)
	}
}

func TestSendVerificationEmailAsyncLogsSuccess(t *testing.T) {
	mailer := &blockingVerificationMailer{
		started: make(chan mailCall, 1),
		release: make(chan struct{}),
	}

	logs := newAsyncMailLogCapture("event=resend_verification.mail_send_completed")
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
	log.SetOutput(logs)

	sendVerificationEmailAsync("req-456", "resend_verification", "user-2", "resend@example.com", "secret-token", "system", mailer)

	select {
	case <-mailer.started:
	case <-time.After(time.Second):
		t.Fatal("mailer was not called")
	}
	close(mailer.release)
	select {
	case <-logs.completed:
	case <-time.After(time.Second):
		t.Fatal("completion log was not written")
	}

	output := logs.String()
	if !strings.Contains(output, "event=resend_verification.mail_send_completed request_id=req-456 user_id=user-2 email=resend@example.com language=en mail_sent=true") {
		t.Fatalf("missing success log: %s", output)
	}
	if strings.Contains(output, "secret-token") {
		t.Fatalf("log output leaked token: %s", output)
	}
}

func TestResolvePublicLanguage(t *testing.T) {
	tests := []struct {
		name           string
		requested      string
		acceptLanguage string
		want           string
	}{
		{name: "requested fr", requested: "fr", acceptLanguage: "en-US,en;q=0.9", want: "fr"},
		{name: "requested en", requested: "en", acceptLanguage: "fr-FR,fr;q=0.9", want: "en"},
		{name: "accept fr", requested: "", acceptLanguage: "fr-FR,fr;q=0.9,en;q=0.8", want: "fr"},
		{name: "accept en", requested: "", acceptLanguage: "en-US,en;q=0.9", want: "en"},
		{name: "fallback", requested: "system", acceptLanguage: "de-DE,de;q=0.9", want: "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolvePublicLanguage(tt.requested, tt.acceptLanguage)
			if got != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got)
			}
		})
	}
}

type blockingVerificationMailer struct {
	started chan mailCall
	release chan struct{}
	err     error
}

type asyncMailLogCapture struct {
	mu        sync.Mutex
	buf       bytes.Buffer
	completed chan struct{}
	marker    string
}

func newAsyncMailLogCapture(marker string) *asyncMailLogCapture {
	return &asyncMailLogCapture{
		completed: make(chan struct{}),
		marker:    marker,
	}
}

func (c *asyncMailLogCapture) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	n, err := c.buf.Write(p)
	if strings.Contains(string(p), c.marker) {
		close(c.completed)
	}
	return n, err
}

func (c *asyncMailLogCapture) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}

type mailCall struct {
	requestID string
	to        string
	token     string
	language  string
}

func (m *blockingVerificationMailer) SendVerificationEmail(requestID, to, token, language string) error {
	m.started <- mailCall{requestID: requestID, to: to, token: token, language: language}
	<-m.release
	return m.err
}
