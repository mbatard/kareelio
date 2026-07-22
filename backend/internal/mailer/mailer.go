package mailer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"github.com/user/kareelio/backend/internal/config"
)

type Mailer struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) IsConfigured() bool {
	return m.cfg.SMTPHost != "" && m.cfg.SMTPPort != "" && m.cfg.SMTPFrom != ""
}

func (m *Mailer) SendVerificationEmail(to, token string) error {
	if !m.IsConfigured() {
		log.Printf("[MAILER] SMTP not configured, verification link: %s/verify-email?token=%s", m.cfg.AppPublicURL, token)
		return nil
	}

	link := fmt.Sprintf("%s/verify-email?token=%s", m.cfg.AppPublicURL, token)

	subject := "Kareelio - Vérifiez votre adresse e-mail"
	if strings.Contains(m.cfg.SMTPFrom, "noreply") || strings.Contains(m.cfg.SMTPFrom, "no-reply") {
		subject = "Kareelio - Verify your email address"
	}

	body := fmt.Sprintf(`Bonjour,

Merci pour votre inscription sur Kareelio.

Cliquez sur le lien ci-dessous pour valider votre adresse e-mail :

%s

Ce lien expire dans %d heures.

Si vous n'avez pas créé de compte, vous pouvez ignorer cet e-mail.

---
Kareelio - Suivi de candidatures

Hello,

Thank you for signing up on Kareelio.

Click the link below to verify your email address:

%s

This link expires in %d hours.

If you did not create an account, you can ignore this email.

---
Kareelio - Job Application Tracker`, link, m.cfg.VerificationTokenTTLHours, link, m.cfg.VerificationTokenTTLHours)

	return m.send(to, subject, body)
}

func (m *Mailer) send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", m.cfg.SMTPHost, m.cfg.SMTPPort)

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
		return m.sendTLS(addr, auth, to, msg)
	}

	return smtp.SendMail(addr, auth, m.cfg.SMTPFrom, []string{to}, []byte(msg))
}

func (m *Mailer) sendTLS(addr string, auth smtp.Auth, to, msg string) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: false,
	})
	if err != nil {
		return fmt.Errorf("unable to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("unable to create SMTP client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("unable to authenticate: %w", err)
		}
	}

	if err := client.Mail(m.cfg.SMTPFrom); err != nil {
		return fmt.Errorf("unable to set sender: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("unable to set recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("unable to send data: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("unable to write message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("unable to close data writer: %w", err)
	}

	return client.Quit()
}
