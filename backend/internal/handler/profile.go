package handler

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/user/kareelio/backend/internal/middleware"
	"github.com/user/kareelio/backend/internal/model"
	"github.com/user/kareelio/backend/internal/repository"
	"github.com/user/kareelio/backend/internal/validation"
)

type ProfileHandler struct {
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
}

func NewProfileHandler(userRepo *repository.UserRepository, auditRepo *repository.AuditRepository) *ProfileHandler {
	return &ProfileHandler{userRepo: userRepo, auditRepo: auditRepo}
}

func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	writeJSON(w, http.StatusOK, user.ToResponse())
}

func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if user.Role == model.RoleAdmin {
		req.Email = nil
	}

	if req.Email != nil {
		normalized := validation.NormalizeEmail(*req.Email)
		req.Email = &normalized
		if !validation.IsValidEmail(normalized) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid email address"})
			return
		}
	}

	updated, err := h.userRepo.UpdateProfile(r.Context(), user.ID, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	ad := middleware.GetAuditData(r.Context())
	if ad != nil {
		ad.TargetType = "user"
		ad.TargetID = user.ID
	}
	middleware.LogAudit(r.Context(), h.auditRepo, model.AuditActionProfileUpdated)

	writeJSON(w, http.StatusOK, updated.ToResponse())
}

func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "current and new password are required"})
		return
	}

	if len(req.NewPassword) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	fullUser, err := h.userRepo.GetByID(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to fetch user"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(fullUser.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "incorrect current password"})
		return
	}

	if err := h.userRepo.UpdatePassword(r.Context(), user.ID, req.NewPassword); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to update password"})
		return
	}

	ad := middleware.GetAuditData(r.Context())
	if ad != nil {
		ad.TargetType = "user"
		ad.TargetID = user.ID
	}
	middleware.LogAudit(r.Context(), h.auditRepo, model.AuditActionPasswordChanged)

	writeJSON(w, http.StatusOK, map[string]string{"message": "password updated"})
}
