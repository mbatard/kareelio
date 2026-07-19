package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/user/kareelio/backend/internal/middleware"
	"github.com/user/kareelio/backend/internal/model"
	"github.com/user/kareelio/backend/internal/repository"
)

type JobApplicationHandler struct {
	jaRepo     *repository.JobApplicationRepository
	auditRepo  *repository.AuditRepository
}

func NewJobApplicationHandler(jaRepo *repository.JobApplicationRepository, auditRepo *repository.AuditRepository) *JobApplicationHandler {
	return &JobApplicationHandler{jaRepo: jaRepo, auditRepo: auditRepo}
}

func (h *JobApplicationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req model.CreateJobApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Company == "" || req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "company and title are required"})
		return
	}

	if req.Status == "" {
		req.Status = model.StatusDraft
	}

	ja, err := h.jaRepo.Create(r.Context(), userID, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to create job application"})
		return
	}

	ad := middleware.GetAuditData(r.Context())
	if ad != nil {
		ad.TargetType = "job_application"
		ad.TargetID = ja.ID
	}
	middleware.LogAudit(r.Context(), h.auditRepo, model.AuditActionJobAppCreated)

	writeJSON(w, http.StatusCreated, ja)
}

func (h *JobApplicationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	applications, err := h.jaRepo.List(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to list job applications"})
		return
	}

	if applications == nil {
		applications = []model.JobApplication{}
	}

	writeJSON(w, http.StatusOK, applications)
}

func (h *JobApplicationHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	ja, err := h.jaRepo.GetByID(r.Context(), userID, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job application not found"})
		return
	}

	writeJSON(w, http.StatusOK, ja)
}

func (h *JobApplicationHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	var req model.UpdateJobApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	ja, err := h.jaRepo.Update(r.Context(), userID, id, req)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	ad := middleware.GetAuditData(r.Context())
	if ad != nil {
		ad.TargetType = "job_application"
		ad.TargetID = ja.ID
	}
	middleware.LogAudit(r.Context(), h.auditRepo, model.AuditActionJobAppUpdated)

	writeJSON(w, http.StatusOK, ja)
}

func (h *JobApplicationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.jaRepo.Delete(r.Context(), userID, id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	ad := middleware.GetAuditData(r.Context())
	if ad != nil {
		ad.TargetType = "job_application"
		ad.TargetID = id
	}
	middleware.LogAudit(r.Context(), h.auditRepo, model.AuditActionJobAppDeleted)

	writeJSON(w, http.StatusOK, map[string]string{"message": "job application deleted"})
}
