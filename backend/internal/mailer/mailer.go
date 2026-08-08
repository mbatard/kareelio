package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/user/kareelio/backend/internal/config"
)

type Mailer struct {
	cfg           *config.Config
	sendFunc      func(to, subject, body string) error
	dialFunc      func(ctx context.Context, network, addr string) (net.Conn, error)
	newClientFunc func(net.Conn, string) (smtpClient, error)
}

func New(cfg *config.Config) *Mailer {
	m := &Mailer{cfg: cfg}
	m.sendFunc = m.send
	m.dialFunc = defaultDial
	m.newClientFunc = defaultNewClient
	return m
}

type smtpClient interface {
	Extension(string) (bool, string)
	StartTLS(*tls.Config) error
	Auth(smtp.Auth) error
	Mail(string) error
	Rcpt(string) error
	Data() (io.WriteCloser, error)
	Quit() error
	Close() error
}

func (m *Mailer) IsConfigured() bool {
	return m.cfg.SMTPHost != "" && m.cfg.SMTPPort != "" && m.cfg.SMTPFrom != ""
}

func LogSMTPConfigSummary(cfg *config.Config) {
	log.Print(SMTPConfigSummary(cfg))
}

func SMTPConfigSummary(cfg *config.Config) string {
	hostPresent := cfg.SMTPHost != ""
	portPresent := cfg.SMTPPort != ""
	fromPresent := cfg.SMTPFrom != ""
	authEnabled := cfg.SMTPUsername != "" && cfg.SMTPPassword != ""
	timeout := cfg.SMTPTimeoutSeconds
	configured := hostPresent && portPresent && fromPresent
	transport := "disabled"
	tlsMode := "none"
	if configured {
		if cfg.SMTPPort == "465" {
			transport = "implicit_tls"
			tlsMode = "implicit"
		} else if cfg.SMTPPort == "25" {
			transport = "smtp_plain"
		} else {
			transport = "smtp_client"
		}
	}

	return fmt.Sprintf(
		"event=smtp.config_summary host_present=%t port=%s from_present=%t auth_enabled=%t timeout_seconds=%d tls_mode=%s transport=%s configured=%t",
		hostPresent,
		cfg.SMTPPort,
		fromPresent,
		authEnabled,
		timeout,
		tlsMode,
		transport,
		configured,
	)
}

func (m *Mailer) SendVerificationEmail(requestID, to, token, language string) error {
	language = normalizeLanguage(language)
	subject, body := m.verificationEmailContent(token, language)

	if !m.IsConfigured() {
		logMailEvent("mail.verification_send_start", requestID, "recipient="+to, "language="+language, "configured=false", "transport=log_only")
		logMailEvent("mail.verification_send_success", requestID, "recipient="+to, "language="+language, "configured=false", "transport=log_only")
		return nil
	}

	logMailEvent("mail.verification_send_start", requestID, "recipient="+to, "language="+language, "configured=true", "transport="+m.transportMode())

	send := m.sendFunc
	if send == nil {
		send = m.send
	}
	if err := send(to, subject, body); err != nil {
		logMailEvent("mail.verification_send_error", requestID, "recipient="+to, "language="+language, "error="+err.Error())
		return err
	}

	logMailEvent("mail.verification_send_success", requestID, "recipient="+to, "language="+language, "configured=true", "transport="+m.transportMode())
	return nil
}

func (m *Mailer) SendAdminNewRegistrationEmail(requestID, adminEmail, registeredUserEmail, registeredDisplayName, language string) error {
	language = normalizeLanguage(language)
	subject, body := m.adminNewRegistrationEmailContent(registeredUserEmail, registeredDisplayName, language)

	if !m.IsConfigured() {
		logMailEvent("mail.admin_registration_send_start", requestID, "recipient="+adminEmail, "language="+language, "configured=false", "transport=log_only")
		logMailEvent("mail.admin_registration_send_success", requestID, "recipient="+adminEmail, "language="+language, "configured=false", "transport=log_only")
		return nil
	}

	logMailEvent("mail.admin_registration_send_start", requestID, "recipient="+adminEmail, "language="+language, "configured=true", "transport="+m.transportMode())

	send := m.sendFunc
	if send == nil {
		send = m.send
	}
	if err := send(adminEmail, subject, body); err != nil {
		logMailEvent("mail.admin_registration_send_error", requestID, "recipient="+adminEmail, "language="+language, "error="+err.Error())
		return err
	}

	logMailEvent("mail.admin_registration_send_success", requestID, "recipient="+adminEmail, "language="+language, "configured=true", "transport="+m.transportMode())
	return nil
}

func (m *Mailer) verificationEmailContent(token, language string) (string, string) {
	if language == "fr" {
		return "Kareelio - Vérifiez votre adresse e-mail", fmt.Sprintf(`Bonjour,

Merci pour votre inscription sur Kareelio.

Cliquez sur le lien ci-dessous pour valider votre adresse e-mail :

%s/verify-email?token=%s

Ce lien expire dans %d heures.

Si vous n'avez pas créé de compte, vous pouvez ignorer cet e-mail.

---
Kareelio - Suivi de candidatures`, m.cfg.AppPublicURL, token, m.cfg.VerificationTokenTTLHours)
	}

	return "Kareelio - Verify your email address", fmt.Sprintf(`Hello,

Thank you for signing up on Kareelio.

Click the link below to verify your email address:

%s/verify-email?token=%s

This link expires in %d hours.

If you did not create an account, you can ignore this email.

---
Kareelio - Job Application Tracker`, m.cfg.AppPublicURL, token, m.cfg.VerificationTokenTTLHours)
}

func (m *Mailer) adminNewRegistrationEmailContent(registeredUserEmail, registeredDisplayName, language string) (string, string) {
	adminUsersURL := strings.TrimRight(m.cfg.AppPublicURL, "/") + "/admin/users"
	if language == "fr" {
		return "Kareelio - Nouvel utilisateur inscrit", fmt.Sprintf(`Bonjour,

Un nouvel utilisateur vient de s'inscrire sur Kareelio.

Nom affiché : %s
E-mail : %s

Consultez la liste des utilisateurs : %s

---
Kareelio - Suivi de candidatures`, registeredDisplayName, registeredUserEmail, adminUsersURL)
	}

	return "Kareelio - New user registration", fmt.Sprintf(`Hello,

A new user just registered on Kareelio.

Display name: %s
Email: %s

View the users list: %s

---
Kareelio - Job Application Tracker`, registeredDisplayName, registeredUserEmail, adminUsersURL)
}

func normalizeLanguage(language string) string {
	if language == "fr" {
		return "fr"
	}
	return "en"
}

func (m *Mailer) send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", m.cfg.SMTPHost, m.cfg.SMTPPort)
	timeout := time.Duration(m.cfg.SMTPTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", m.cfg.SMTPFrom, to, subject, body)

	var auth smtp.Auth
	if m.cfg.SMTPUsername != "" && m.cfg.SMTPPassword != "" {
		auth = smtp.PlainAuth("", m.cfg.SMTPUsername, m.cfg.SMTPPassword, m.cfg.SMTPHost)
	}

	if m.cfg.SMTPPort == "465" {
		return m.sendImplicitTLS(ctx, addr, auth, to, msg)
	}

	dialFunc := m.dialFunc
	if dialFunc == nil {
		dialFunc = defaultDial
	}
	conn, err := dialFunc(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp connect %s: %w", addr, err)
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	newClientFunc := m.newClientFunc
	if newClientFunc == nil {
		newClientFunc = defaultNewClient
	}
	client, err := newClientFunc(conn, m.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("smtp client init: %w", err)
	}
	defer client.Close()

	if m.cfg.SMTPPort != "25" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: m.cfg.SMTPHost}); err != nil {
				return fmt.Errorf("smtp starttls: %w", err)
			}
		}
	}

	return m.deliverSMTP(client, auth, to, msg)
}

func (m *Mailer) transportMode() string {
	if m.cfg.SMTPPort == "465" {
		return "implicit_tls"
	}
	if m.cfg.SMTPPort == "25" {
		return "smtp_plain"
	}
	return "smtp_client"
}

func (m *Mailer) sendImplicitTLS(ctx context.Context, addr string, auth smtp.Auth, to, msg string) error {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp connect %s: %w", addr, err)
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	tlsConn := tls.Client(conn, &tls.Config{ServerName: m.cfg.SMTPHost})
	if err := tlsConn.Handshake(); err != nil {
		return fmt.Errorf("smtp tls handshake: %w", err)
	}

	newClientFunc := m.newClientFunc
	if newClientFunc == nil {
		newClientFunc = defaultNewClient
	}
	client, err := newClientFunc(tlsConn, m.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("smtp client init: %w", err)
	}
	defer client.Close()

	return m.deliverSMTP(client, auth, to, msg)
}

func (m *Mailer) deliverSMTP(client smtpClient, auth smtp.Auth, to, msg string) error {
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(m.cfg.SMTPFrom); err != nil {
		return fmt.Errorf("smtp sender: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		_ = w.Close()
		return fmt.Errorf("smtp data write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp data close: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}
	return nil
}

func logMailEvent(event, requestID string, fields ...string) {
	parts := []string{"event=" + event}
	if requestID != "" {
		parts = append(parts, "request_id="+requestID)
	}
	parts = append(parts, fields...)
	log.Print(strings.Join(parts, " "))
}

func defaultDial(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := net.Dialer{}
	return dialer.DialContext(ctx, network, addr)
}

func defaultNewClient(conn net.Conn, host string) (smtpClient, error) {
	return smtp.NewClient(conn, host)
}
