package handler

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/user/kareelio/backend/internal/model"
	"github.com/user/kareelio/backend/internal/validation"
)

type adminNotificationAdminLookup interface {
	GetActiveAdmin(context.Context) (*model.User, error)
}

type adminRegistrationMailer interface {
	SendAdminNewRegistrationEmail(requestID, adminEmail, registeredUserEmail, registeredDisplayName, language string) error
}

func sendAdminNewRegistrationEmailAsync(requestID string, adminRepo adminNotificationAdminLookup, mailer adminRegistrationMailer, registeredUser *model.User) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	go func() {
		defer cancel()
		sendAdminNewRegistrationEmail(ctx, requestID, adminRepo, mailer, registeredUser)
	}()
}

func sendAdminNewRegistrationEmail(ctx context.Context, requestID string, adminRepo adminNotificationAdminLookup, mailer adminRegistrationMailer, registeredUser *model.User) {
	admin, err := adminRepo.GetActiveAdmin(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logAuthEventByRequestID(requestID, "register.admin_notification.skipped", "user_id="+registeredUser.ID, "reason=no_active_admin")
			return
		}
		logAuthEventByRequestID(requestID, "register.admin_notification.error", "user_id="+registeredUser.ID, "stage=admin_lookup", "error="+err.Error())
		return
	}

	adminEmail := validation.NormalizeEmail(admin.Email)
	if !validation.IsValidEmail(adminEmail) {
		logAuthEventByRequestID(requestID, "register.admin_notification.skipped", "user_id="+registeredUser.ID, "admin_user_id="+admin.ID, "reason=invalid_admin_email")
		return
	}
	if isLocalAdminNotificationEmail(adminEmail) {
		logAuthEventByRequestID(requestID, "register.admin_notification.skipped", "user_id="+registeredUser.ID, "admin_user_id="+admin.ID, "reason=local_admin_email")
		return
	}

	registeredDisplayName := registeredUser.DisplayName
	if registeredDisplayName == "" {
		registeredDisplayName = registeredUser.Email
	}
	language := resolveMailLanguage(admin.Language)
	logAuthEventByRequestID(requestID, "register.admin_notification.mail_send_queued", "user_id="+registeredUser.ID, "admin_user_id="+admin.ID, "language="+language)

	if err := mailer.SendAdminNewRegistrationEmail(requestID, adminEmail, registeredUser.Email, registeredDisplayName, language); err != nil {
		logAuthEventByRequestID(requestID, "register.admin_notification.error", "user_id="+registeredUser.ID, "admin_user_id="+admin.ID, "stage=mail_send", "error="+err.Error())
		return
	}

	logAuthEventByRequestID(requestID, "register.admin_notification.mail_send_completed", "user_id="+registeredUser.ID, "admin_user_id="+admin.ID, "mail_sent=true")
}

func isLocalAdminNotificationEmail(email string) bool {
	_, domain, ok := strings.Cut(email, "@")
	if !ok {
		return false
	}
	return domain == "kareelio.local" || strings.HasSuffix(domain, ".local")
}
