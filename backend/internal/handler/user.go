package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/user/kareelio/backend/internal/middleware"
	"github.com/user/kareelio/backend/internal/model"
	"github.com/user/kareelio/backend/internal/repository"
	"github.com/user/kareelio/backend/internal/validation"
)

type UserHandler struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	auditRepo   *repository.AuditRepository
}

func NewUserHandler(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository, auditRepo *repository.AuditRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo, sessionRepo: sessionRepo, auditRepo: auditRepo}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to list users"})
		return
	}

	responses := make([]model.UserResponse, len(users))
	for i, u := range users {
		responses[i] = u.ToResponse()
	}

	writeJSON(w, http.StatusOK, responses)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, user.ToResponse())
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	req.Email = validation.NormalizeEmail(req.Email)

	if !validation.IsValidEmail(req.Email) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid email address"})
		return
	}

	if req.DisplayName == "" {
		req.DisplayName = req.Email
	}

	user, err := h.userRepo.Create(r.Context(), req, model.RoleUser)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "user already exists or invalid data"})
		return
	}

	ad := middleware.GetAuditData(r.Context())
	if ad != nil {
		ad.TargetType = "user"
		ad.TargetID = user.ID
		ad.Metadata["target_email"] = user.Email
	}
	middleware.LogAudit(r.Context(), h.auditRepo, model.AuditActionUserCreated)

	writeJSON(w, http.StatusCreated, user.ToResponse())
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	existing, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	if existing.Role == model.RoleAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "cannot modify admin user"})
		return
	}

	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email != nil {
		normalized := validation.NormalizeEmail(*req.Email)
		req.Email = &normalized
		if !validation.IsValidEmail(normalized) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid email address"})
			return
		}
	}

	user, err := h.userRepo.Update(r.Context(), id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	ad := middleware.GetAuditData(r.Context())
	action := model.AuditActionUserUpdated
	if req.IsActive != nil && !*req.IsActive {
		action = model.AuditActionUserDeactivated
	}
	if req.IsActive != nil && *req.IsActive {
		action = model.AuditActionUserActivated
	}
	if ad != nil {
		ad.TargetType = "user"
		ad.TargetID = user.ID
		ad.Metadata["target_email"] = user.Email
	}
	middleware.LogAudit(r.Context(), h.auditRepo, action)

	writeJSON(w, http.StatusOK, user.ToResponse())
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	existing, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	if existing.Role == model.RoleAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "cannot modify admin user"})
		return
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "new password is required"})
		return
	}

	if len(req.NewPassword) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	if err := h.userRepo.UpdatePassword(r.Context(), id, req.NewPassword); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to update password"})
		return
	}

	_ = h.sessionRepo.DeleteByUserID(r.Context(), id)

	ad := middleware.GetAuditData(r.Context())
	if ad != nil {
		ad.TargetType = "user"
		ad.TargetID = id
		ad.Metadata["target_email"] = existing.Email
	}
	middleware.LogAudit(r.Context(), h.auditRepo, model.AuditActionAdminPasswordReset)

	writeJSON(w, http.StatusOK, map[string]string{"message": "password updated"})
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	existing, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	if existing.Role == model.RoleAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "cannot delete admin user"})
		return
	}

	if err := h.userRepo.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "cannot delete user"})
		return
	}

	ad := middleware.GetAuditData(r.Context())
	if ad != nil {
		ad.TargetType = "user"
		ad.TargetID = id
		ad.Metadata["target_email"] = existing.Email
	}
	middleware.LogAudit(r.Context(), h.auditRepo, model.AuditActionUserDeleted)

	writeJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}
