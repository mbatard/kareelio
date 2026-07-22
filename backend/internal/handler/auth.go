package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

	if !user.IsActive {
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = validation.NormalizeEmail(req.Email)

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	if !validation.IsValidEmail(req.Email) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid email address"})
		return
	}

	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	if req.DisplayName == "" {
		req.DisplayName = req.Email
	}

	createReq := model.CreateUserRequest{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}

	user, err := h.userRepo.Create(r.Context(), createReq, model.RoleUser)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email can be registered, a verification email has been sent"})
		return
	}

	if err := h.userRepo.Deactivate(r.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email can be registered, a verification email has been sent"})
		return
	}

	token, tokenHash, err := generateVerificationToken()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email can be registered, a verification email has been sent"})
		return
	}

	expiresAt := time.Now().Add(time.Duration(h.cfg.VerificationTokenTTLHours) * time.Hour)
	if err := h.evRepo.Create(r.Context(), user.ID, tokenHash, expiresAt.UTC().Format(time.RFC3339)); err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email can be registered, a verification email has been sent"})
		return
	}

	_ = h.mailer.SendVerificationEmail(user.Email, token)

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

	if err := h.userRepo.Activate(r.Context(), userID); err != nil {
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = validation.NormalizeEmail(req.Email)
	if req.Email == "" {
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email is registered and unverified, a verification email has been sent"})
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil || user.IsActive {
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email is registered and unverified, a verification email has been sent"})
		return
	}

	_ = h.evRepo.DeleteForUser(r.Context(), user.ID)

	token, tokenHash, err := generateVerificationToken()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email is registered and unverified, a verification email has been sent"})
		return
	}

	expiresAt := time.Now().Add(time.Duration(h.cfg.VerificationTokenTTLHours) * time.Hour)
	if err := h.evRepo.Create(r.Context(), user.ID, tokenHash, expiresAt.UTC().Format(time.RFC3339)); err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email is registered and unverified, a verification email has been sent"})
		return
	}

	_ = h.mailer.SendVerificationEmail(user.Email, token)

	writeJSON(w, http.StatusOK, map[string]string{"message": "if the email is registered and unverified, a verification email has been sent"})
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
