package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/user/kareelio/backend/internal/config"
	"github.com/user/kareelio/backend/internal/mailer"
	"github.com/user/kareelio/backend/internal/middleware"
	"github.com/user/kareelio/backend/internal/model"
	"github.com/user/kareelio/backend/internal/repository"
	"github.com/user/kareelio/backend/internal/validation"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	auditRepo   *repository.AuditRepository
	evRepo      *repository.EmailVerificationRepository
	mailer      *mailer.Mailer
	cfg         *config.Config
}

func NewAuthHandler(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository, auditRepo *repository.AuditRepository, evRepo *repository.EmailVerificationRepository, mailer *mailer.Mailer, cfg *config.Config) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, sessionRepo: sessionRepo, auditRepo: auditRepo, evRepo: evRepo, mailer: mailer, cfg: cfg}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		h.logLoginFailure(r, req.Email)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		h.logLoginFailure(r, req.Email)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if user.EmailVerifiedAt == nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "account not verified", "code": "email_not_verified"})
		return
	}

	session, err := h.sessionRepo.Create(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to create session"})
		return
	}

	sameSite := http.SameSiteLaxMode
	switch strings.ToLower(h.cfg.SessionCookieSameSite) {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: sameSite,
		MaxAge:   h.cfg.SessionDurationHours * 3600,
	})

	_ = h.auditRepo.Log(r.Context(), &model.AuditEvent{
		ActorUserID: &user.ID,
		ActorEmail:  user.Email,
		ActorRole:   string(user.Role),
		ActorIP:     middleware.ClientIP(r),
		Action:      model.AuditActionLoginSuccess,
		TargetType:  "user",
		TargetID:    user.ID,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"user": user.ToResponse(),
	})
}

func (h *AuthHandler) logLoginFailure(r *http.Request, email string) {
	_ = h.auditRepo.Log(r.Context(), &model.AuditEvent{
		ActorEmail: email,
		ActorRole:  "anonymous",
		ActorIP:    middleware.ClientIP(r),
		Action:     model.AuditActionLoginFailed,
		TargetType: "user",
		TargetID:   email,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value(middleware.SessionKey).(string)
	if ok {
		_ = h.sessionRepo.Delete(r.Context(), sessionID)
	}

	user := middleware.GetUserFromContext(r.Context())
	if user != nil {
		_ = h.auditRepo.Log(r.Context(), &model.AuditEvent{
			ActorUserID: &user.ID,
			ActorEmail:  user.Email,
			ActorRole:   string(user.Role),
			ActorIP:     middleware.ClientIP(r),
			Action:      model.AuditActionLogout,
			TargetType:  "user",
			TargetID:    user.ID,
		})
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	writeJSON(w, http.StatusOK, user.ToResponse())
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.RegistrationEnabled {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "registration is disabled"})
		return
	}

	var req struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logAuthEvent(r, "register.error", "stage=decode", "error=invalid request body")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = validation.NormalizeEmail(req.Email)

	if req.Email == "" || req.Password == "" {
		logAuthEvent(r, "register.error", "stage=validation", "reason=missing_email_or_password")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	if !validation.IsValidEmail(req.Email) {
		logAuthEvent(r, "register.error", "stage=validation", "email="+req.Email, "reason=invalid_email")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid email address"})
		return
	}

	if err := validation.ValidatePassword(req.Password); err != nil {
		logAuthEvent(r, "register.error", "stage=validation", "email="+req.Email, "error="+err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if req.DisplayName == "" {
		req.DisplayName = req.Email
	}

	requestID := r.Header.Get("X-Request-ID")
	logAuthEvent(r, "register.start", "email="+req.Email)

	createReq := model.CreateUserRequest{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}

	user, err := h.userRepo.Create(r.Context(), createReq, model.RoleUser)
	if err != nil {
		logAuthEvent(r, "register.error", "stage=user_create", "email="+req.Email, "error="+err.Error())
		logAuthEvent(r, "register.completed", "email="+req.Email, "status=generic_success", "mail_sent=false")
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email can be registered, a verification email has been sent"})
		return
	}
	logAuthEvent(r, "register.user_created", "user_id="+user.ID, "email="+user.Email)

	token, tokenHash, err := generateVerificationToken()
	if err != nil {
		logAuthEvent(r, "register.error", "stage=token_generate", "user_id="+user.ID, "email="+user.Email, "error="+err.Error())
		logAuthEvent(r, "register.completed", "user_id="+user.ID, "email="+user.Email, "status=generic_success", "mail_sent=false")
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email can be registered, a verification email has been sent"})
		return
	}

	expiresAt := time.Now().Add(time.Duration(h.cfg.VerificationTokenTTLHours) * time.Hour)
	if err := h.evRepo.Create(r.Context(), user.ID, tokenHash, expiresAt.UTC().Format(time.RFC3339)); err != nil {
		logAuthEvent(r, "register.error", "stage=token_create", "user_id="+user.ID, "email="+user.Email, "error="+err.Error())
		logAuthEvent(r, "register.completed", "user_id="+user.ID, "email="+user.Email, "status=generic_success", "mail_sent=false")
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email can be registered, a verification email has been sent"})
		return
	}
	logAuthEvent(r, "register.token_created", "user_id="+user.ID, "email="+user.Email, "expires_at="+expiresAt.UTC().Format(time.RFC3339))

	mailErr := h.mailer.SendVerificationEmail(requestID, user.Email, token)
	mailSent := mailErr == nil
	if mailErr != nil {
		logAuthEvent(r, "register.error", "stage=mail_send", "user_id="+user.ID, "email="+user.Email, "error="+mailErr.Error())
	}
	logAuthEvent(r, "register.completed", "user_id="+user.ID, "email="+user.Email, fmt.Sprintf("mail_sent=%t", mailSent))

	_ = h.auditRepo.Log(r.Context(), &model.AuditEvent{
		ActorUserID: &user.ID,
		ActorEmail:  user.Email,
		ActorRole:   string(model.RoleUser),
		ActorIP:     middleware.ClientIP(r),
		Action:      model.AuditActionUserRegistered,
		TargetType:  "user",
		TargetID:    user.ID,
	})

	writeJSON(w, http.StatusCreated, map[string]string{"message": "account created, please verify your email"})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token is required"})
		return
	}

	tokenHash := hashToken(req.Token)

	userID, err := h.evRepo.GetValid(r.Context(), tokenHash)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired token"})
		return
	}

	if err := h.evRepo.MarkUsed(r.Context(), tokenHash); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to verify email"})
		return
	}

	if err := h.userRepo.SetEmailVerified(r.Context(), userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to activate account"})
		return
	}

	user, _ := h.userRepo.GetByID(r.Context(), userID)
	if user != nil {
		_ = h.auditRepo.Log(r.Context(), &model.AuditEvent{
			ActorUserID: &user.ID,
			ActorEmail:  user.Email,
			ActorRole:   string(user.Role),
			ActorIP:     middleware.ClientIP(r),
			Action:      model.AuditActionEmailVerified,
			TargetType:  "user",
			TargetID:    user.ID,
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "email verified successfully"})
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logAuthEvent(r, "resend_verification.error", "stage=decode", "error=invalid request body")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = validation.NormalizeEmail(req.Email)
	logAuthEvent(r, "resend_verification.start", "email="+req.Email)
	if req.Email == "" {
		logAuthEvent(r, "resend_verification.error", "stage=validation", "reason=empty_email")
		logAuthEvent(r, "resend_verification.completed", "email=", "status=generic_success", "mail_sent=false")
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email is registered and unverified, a verification email has been sent"})
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil || user.EmailVerifiedAt != nil {
		logAuthEvent(r, "resend_verification.error", "stage=user_lookup", "email="+req.Email, "reason=not_found_or_verified")
		logAuthEvent(r, "resend_verification.completed", "email="+req.Email, "status=generic_success", "mail_sent=false")
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email is registered and unverified, a verification email has been sent"})
		return
	}
	logAuthEvent(r, "resend_verification.user_loaded", "user_id="+user.ID, "email="+user.Email)

	if err := h.evRepo.DeleteForUser(r.Context(), user.ID); err != nil {
		logAuthEvent(r, "resend_verification.error", "stage=token_delete", "user_id="+user.ID, "email="+user.Email, "error="+err.Error())
	}

	token, tokenHash, err := generateVerificationToken()
	if err != nil {
		logAuthEvent(r, "resend_verification.error", "stage=token_generate", "user_id="+user.ID, "email="+user.Email, "error="+err.Error())
		logAuthEvent(r, "resend_verification.completed", "user_id="+user.ID, "email="+user.Email, "status=generic_success", "mail_sent=false")
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email is registered and unverified, a verification email has been sent"})
		return
	}

	expiresAt := time.Now().Add(time.Duration(h.cfg.VerificationTokenTTLHours) * time.Hour)
	if err := h.evRepo.Create(r.Context(), user.ID, tokenHash, expiresAt.UTC().Format(time.RFC3339)); err != nil {
		logAuthEvent(r, "resend_verification.error", "stage=token_create", "user_id="+user.ID, "email="+user.Email, "error="+err.Error())
		logAuthEvent(r, "resend_verification.completed", "user_id="+user.ID, "email="+user.Email, "status=generic_success", "mail_sent=false")
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email is registered and unverified, a verification email has been sent"})
		return
	}
	logAuthEvent(r, "resend_verification.token_created", "user_id="+user.ID, "email="+user.Email, "expires_at="+expiresAt.UTC().Format(time.RFC3339))

	mailErr := h.mailer.SendVerificationEmail(r.Header.Get("X-Request-ID"), user.Email, token)
	mailSent := mailErr == nil
	if mailErr != nil {
		logAuthEvent(r, "resend_verification.error", "stage=mail_send", "user_id="+user.ID, "email="+user.Email, "error="+mailErr.Error())
	}
	logAuthEvent(r, "resend_verification.completed", "user_id="+user.ID, "email="+user.Email, fmt.Sprintf("mail_sent=%t", mailSent))

	_ = h.auditRepo.Log(r.Context(), &model.AuditEvent{
		ActorUserID: &user.ID,
		ActorEmail:  user.Email,
		ActorRole:   string(user.Role),
		ActorIP:     middleware.ClientIP(r),
		Action:      model.AuditActionEmailVerificationResent,
		TargetType:  "user",
		TargetID:    user.ID,
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "if the email is registered and unverified, a verification email has been sent"})
}

func (h *AuthHandler) AdminResendVerification(w http.ResponseWriter, r *http.Request) {
	status, payload := adminResendVerification(r, h.userRepo, h.evRepo, h.mailer, h.auditRepo, h.cfg.VerificationTokenTTLHours)
	writeJSON(w, status, payload)
}

func generateVerificationToken() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("unable to generate token: %w", err)
	}
	token := hex.EncodeToString(b)
	h := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(h[:])
	return token, tokenHash, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

type userLookup interface {
	GetByID(context.Context, string) (*model.User, error)
}

type verificationTokenStore interface {
	DeleteForUser(context.Context, string) error
	Create(context.Context, string, string, string) error
}

type verificationMailer interface {
	SendVerificationEmail(requestID, to, token string) error
}

type auditLogger interface {
	Log(context.Context, *model.AuditEvent) error
}

func adminResendVerification(r *http.Request, userRepo userLookup, evRepo verificationTokenStore, mailer verificationMailer, auditRepo auditLogger, ttlHours int) (int, map[string]string) {
	requestID := r.Header.Get("X-Request-ID")
	targetID := chi.URLParam(r, "id")
	actor := middleware.GetUserFromContext(r.Context())
	clientIP := middleware.ClientIP(r)

	logAdminResendEvent(requestID, "admin_resend_verification.start", "target_user_id="+targetID)

	target, err := userRepo.GetByID(r.Context(), targetID)
	if err != nil {
		logAdminResendEvent(requestID, "admin_resend_verification.error", "target_user_id="+targetID, "stage=user_lookup", "reason=not_found")
		return http.StatusNotFound, map[string]string{"error": "user not found"}
	}
	if target.Role == model.RoleAdmin {
		logAdminResendEvent(requestID, "admin_resend_verification.error", "target_user_id="+target.ID, "target_email="+target.Email, "stage=user_lookup", "reason=admin_user")
		return http.StatusForbidden, map[string]string{"error": "cannot resend verification for admin user"}
	}
	if target.EmailVerifiedAt != nil {
		logAdminResendEvent(requestID, "admin_resend_verification.error", "target_user_id="+target.ID, "target_email="+target.Email, "stage=user_lookup", "reason=already_verified")
		return http.StatusConflict, map[string]string{"error": "user already verified"}
	}

	if err := evRepo.DeleteForUser(r.Context(), target.ID); err != nil {
		logAdminResendEvent(requestID, "admin_resend_verification.error", "target_user_id="+target.ID, "target_email="+target.Email, "stage=token_delete", "error="+err.Error())
		return http.StatusInternalServerError, map[string]string{"error": "unable to reset verification token"}
	}

	token, tokenHash, err := generateVerificationToken()
	if err != nil {
		logAdminResendEvent(requestID, "admin_resend_verification.error", "target_user_id="+target.ID, "target_email="+target.Email, "stage=token_generate", "error="+err.Error())
		return http.StatusInternalServerError, map[string]string{"error": "unable to generate verification token"}
	}

	expiresAt := time.Now().Add(time.Duration(ttlHours) * time.Hour)
	if err := evRepo.Create(r.Context(), target.ID, tokenHash, expiresAt.UTC().Format(time.RFC3339)); err != nil {
		logAdminResendEvent(requestID, "admin_resend_verification.error", "target_user_id="+target.ID, "target_email="+target.Email, "stage=token_create", "error="+err.Error())
		return http.StatusInternalServerError, map[string]string{"error": "unable to store verification token"}
	}

	logAdminResendEvent(requestID, "admin_resend_verification.token_created", "target_user_id="+target.ID, "target_email="+target.Email, "expires_at="+expiresAt.UTC().Format(time.RFC3339))

	if err := mailer.SendVerificationEmail(requestID, target.Email, token); err != nil {
		logAdminResendEvent(requestID, "admin_resend_verification.error", "target_user_id="+target.ID, "target_email="+target.Email, "stage=mail_send", "error="+err.Error())
		return http.StatusBadGateway, map[string]string{"error": "unable to send verification email"}
	}

	if err := auditRepo.Log(r.Context(), &model.AuditEvent{
		ActorUserID: actorID(actor),
		ActorEmail:  actorEmail(actor),
		ActorRole:   actorRole(actor),
		ActorIP:     clientIP,
		Action:      model.AuditActionEmailVerificationResent,
		TargetType:  "user",
		TargetID:    target.ID,
		Metadata:    repository.MetadataJSON(map[string]any{"target_email": target.Email}),
	}); err != nil {
		logAdminResendEvent(requestID, "admin_resend_verification.error", "target_user_id="+target.ID, "target_email="+target.Email, "stage=audit", "error="+err.Error())
	}

	logAdminResendEvent(requestID, "admin_resend_verification.completed", "target_user_id="+target.ID, "target_email="+target.Email, "mail_sent=true")
	return http.StatusOK, map[string]string{"message": "verification email sent"}
}

func actorID(user *model.User) *string {
	if user == nil {
		return nil
	}
	return &user.ID
}

func actorEmail(user *model.User) string {
	if user == nil {
		return "system"
	}
	return user.Email
}

func actorRole(user *model.User) string {
	if user == nil {
		return "system"
	}
	return string(user.Role)
}

func logAdminResendEvent(requestID, event string, fields ...string) {
	parts := []string{"event=" + event}
	if requestID != "" {
		parts = append(parts, "request_id="+requestID)
	}
	parts = append(parts, fields...)
	log.Print(strings.Join(parts, " "))
}

func logAuthEvent(r *http.Request, event string, fields ...string) {
	parts := []string{"event=" + event}
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		parts = append(parts, "request_id="+requestID)
	}
	parts = append(parts, fields...)
	log.Print(strings.Join(parts, " "))
}
