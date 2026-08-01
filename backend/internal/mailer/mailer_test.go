package mailer

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/user/kareelio/backend/internal/config"
)

func TestSMTPConfigSummaryUnauthenticatedRelay(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "smtp.internal",
		SMTPPort:     "25",
		SMTPUsername: "smtp-user",
		SMTPPassword: "super-secret",
		SMTPFrom:     "noreply@kareelio.test",
	}

	got := SMTPConfigSummary(cfg)
	for _, want := range []string{
		"event=smtp.config_summary",
		"host_present=true",
		"port=25",
		"from_present=true",
		"auth_enabled=true",
		"tls_mode=none",
		"transport=smtp_plain",
		"configured=true",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("summary missing %q: %s", want, got)
		}
	}

	for _, secret := range []string{"smtp.internal", "smtp-user", "super-secret"} {
		if strings.Contains(got, secret) {
			t.Fatalf("summary leaked %q: %s", secret, got)
		}
	}
}

func TestSendReturnsPhaseSpecificErrors(t *testing.T) {
	baseCfg := &config.Config{
		SMTPHost:                  "smtp.internal",
		SMTPPort:                  "25",
		SMTPFrom:                  "noreply@kareelio.test",
		SMTPTimeoutSeconds:        2,
		VerificationTokenTTLHours: 24,
		AppPublicURL:              "https://app.example",
	}

	t.Run("connect", func(t *testing.T) {
		m := &Mailer{cfg: baseCfg}
		m.dialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return nil, errors.New("dial boom")
		}
		err := m.send("user@example.com", "subject", "body")
		if err == nil || !strings.Contains(err.Error(), "smtp connect") {
			t.Fatalf("expected connect error, got %v", err)
		}
	})

	t.Run("sender", func(t *testing.T) {
		m := &Mailer{cfg: baseCfg}
		m.dialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			c1, c2 := net.Pipe()
			_ = c2.Close()
			return c1, nil
		}
		m.newClientFunc = func(conn net.Conn, host string) (smtpClient, error) {
			return &fakeSMTPClient{mailErr: errors.New("mail boom")}, nil
		}
		err := m.send("user@example.com", "subject", "body")
		if err == nil || !strings.Contains(err.Error(), "smtp sender") {
			t.Fatalf("expected sender error, got %v", err)
		}
	})

	t.Run("recipient", func(t *testing.T) {
		m := &Mailer{cfg: baseCfg}
		m.dialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			c1, c2 := net.Pipe()
			_ = c2.Close()
			return c1, nil
		}
		m.newClientFunc = func(conn net.Conn, host string) (smtpClient, error) {
			return &fakeSMTPClient{rcptErr: errors.New("rcpt boom")}, nil
		}
		err := m.send("user@example.com", "subject", "body")
		if err == nil || !strings.Contains(err.Error(), "smtp recipient") {
			t.Fatalf("expected recipient error, got %v", err)
		}
	})

	t.Run("data", func(t *testing.T) {
		m := &Mailer{cfg: baseCfg}
		m.dialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			c1, c2 := net.Pipe()
			_ = c2.Close()
			return c1, nil
		}
		m.newClientFunc = func(conn net.Conn, host string) (smtpClient, error) {
			return &fakeSMTPClient{dataErr: errors.New("data boom")}, nil
		}
		err := m.send("user@example.com", "subject", "body")
		if err == nil || !strings.Contains(err.Error(), "smtp data") {
			t.Fatalf("expected data error, got %v", err)
		}
	})
}

func TestSendVerificationEmailLogsStartAndSuccessWithoutTokenLeak(t *testing.T) {
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

	m := &Mailer{
		cfg: &config.Config{
			SMTPHost:                  "smtp.internal",
			SMTPPort:                  "25",
			SMTPFrom:                  "noreply@kareelio.test",
			VerificationTokenTTLHours: 24,
			AppPublicURL:              "https://app.example",
		},
	}
	m.sendFunc = func(to, subject, body string) error {
		if to != "user@example.com" {
			t.Fatalf("unexpected recipient: %s", to)
		}
		if subject == "" || body == "" {
			t.Fatal("expected subject and body")
		}
		return nil
	}

	token := "token-secret-123"
	if err := m.SendVerificationEmail("req-123", "user@example.com", token); err != nil {
		t.Fatalf("SendVerificationEmail returned error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"event=mail.verification_send_start",
		"request_id=req-123",
		"recipient=user@example.com",
		"event=mail.verification_send_success",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("log output missing %q: %s", want, out)
		}
	}
	for _, leak := range []string{token, "verify-email?token=" + token} {
		if strings.Contains(out, leak) {
			t.Fatalf("log output leaked %q: %s", leak, out)
		}
	}
}

func TestSendVerificationEmailWithoutSMTPLogsOnly(t *testing.T) {
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

	m := &Mailer{cfg: &config.Config{AppPublicURL: "https://app.example"}}
	token := "token-secret-456"
	if err := m.SendVerificationEmail("req-456", "user@example.com", token); err != nil {
		t.Fatalf("SendVerificationEmail returned error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"event=mail.verification_send_start",
		"configured=false",
		"transport=log_only",
		"event=mail.verification_send_success",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("log output missing %q: %s", want, out)
		}
	}
	for _, leak := range []string{token, "verify-email?token=" + token} {
		if strings.Contains(out, leak) {
			t.Fatalf("log output leaked %q: %s", leak, out)
		}
	}
}

func TestSendVerificationEmailToPlainRelay(t *testing.T) {
	addr, received := startPlainSMTPServer(t)
	host, port, ok := strings.Cut(addr, ":")
	if !ok {
		t.Fatalf("unexpected listener addr: %s", addr)
	}

	m := &Mailer{
		cfg: &config.Config{
			SMTPHost:                  host,
			SMTPPort:                  port,
			SMTPFrom:                  "noreply@kareelio.test",
			SMTPTimeoutSeconds:        2,
			VerificationTokenTTLHours: 24,
			AppPublicURL:              "https://app.example",
		},
	}

	token := "relay-token-789"
	if err := m.SendVerificationEmail("req-relay", "user@example.com", token); err != nil {
		t.Fatalf("SendVerificationEmail returned error: %v", err)
	}

	select {
	case msg := <-received:
		for _, want := range []string{
			"To: user@example.com",
			"Subject: Kareelio - Verify your email address",
			"https://app.example/verify-email?token=" + token,
		} {
			if !strings.Contains(msg, want) {
				t.Fatalf("captured message missing %q: %s", want, msg)
			}
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for captured SMTP message")
	}
}

type fakeSMTPClient struct {
	authErr  error
	mailErr  error
	rcptErr  error
	dataErr  error
	writeErr error
	closeErr error
	quitErr  error
}

func (f *fakeSMTPClient) Extension(string) (bool, string) { return false, "" }
func (f *fakeSMTPClient) StartTLS(*tls.Config) error      { return nil }
func (f *fakeSMTPClient) Auth(smtp.Auth) error            { return f.authErr }
func (f *fakeSMTPClient) Mail(string) error               { return f.mailErr }
func (f *fakeSMTPClient) Rcpt(string) error               { return f.rcptErr }
func (f *fakeSMTPClient) Data() (io.WriteCloser, error) {
	if f.dataErr != nil {
		return nil, f.dataErr
	}
	return f, nil
}
func (f *fakeSMTPClient) Write(p []byte) (int, error) {
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return len(p), nil
}
func (f *fakeSMTPClient) Close() error { return f.closeErr }
func (f *fakeSMTPClient) Quit() error  { return f.quitErr }

func startPlainSMTPServer(t *testing.T) (string, <-chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	msgCh := make(chan string, 1)

	t.Cleanup(func() {
		_ = ln.Close()
	})

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(conn)
		writeLine := func(line string) {
			_, _ = fmt.Fprintf(writer, "%s\r\n", line)
			_ = writer.Flush()
		}

		writeLine("220 test-smtp ready")
		inData := false
		var msg strings.Builder

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			upper := strings.ToUpper(line)

			if inData {
				if line == "." {
					select {
					case msgCh <- msg.String():
					default:
					}
					writeLine("250 queued")
					inData = false
					continue
				}
				if strings.HasPrefix(line, "..") {
					line = line[1:]
				}
				msg.WriteString(line)
				msg.WriteString("\n")
				continue
			}

			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				writeLine("250-localhost")
				writeLine("250 OK")
			case strings.HasPrefix(upper, "MAIL FROM:"):
				writeLine("250 OK")
			case strings.HasPrefix(upper, "RCPT TO:"):
				writeLine("250 OK")
			case upper == "DATA":
				writeLine("354 End data with <CR><LF>.<CR><LF>")
				inData = true
			case upper == "QUIT":
				writeLine("221 Bye")
				return
			default:
				writeLine("502 Command not implemented")
			}
		}
	}()

	return ln.Addr().String(), msgCh
}
