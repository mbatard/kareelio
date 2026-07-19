package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/user/kareelio/backend/internal/config"
	"github.com/user/kareelio/backend/internal/middleware"
	"github.com/user/kareelio/backend/internal/model"
	"github.com/user/kareelio/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	auditRepo   *repository.AuditRepository
	cfg         *config.Config
}

func NewAuthHandler(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository, auditRepo *repository.AuditRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, sessionRepo: sessionRepo, auditRepo: auditRepo, cfg: cfg}
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
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "account disabled"})
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
		Secure:   h.cfg.SessionCookieSecure,
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
