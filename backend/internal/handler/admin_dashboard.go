package handler

import (
	"net/http"

	"github.com/user/kareelio/backend/internal/repository"
)

type AdminDashboardHandler struct {
	dashRepo *repository.AdminDashboardRepository
}

func NewAdminDashboardHandler(dashRepo *repository.AdminDashboardRepository) *AdminDashboardHandler {
	return &AdminDashboardHandler{dashRepo: dashRepo}
}

func (h *AdminDashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	dash, err := h.dashRepo.GetDashboard(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load dashboard"})
		return
	}

	writeJSON(w, http.StatusOK, dash)
}
