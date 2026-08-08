package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/user/kareelio/backend/internal/middleware"
	"github.com/user/kareelio/backend/internal/repository"
)

type AdminNotificationHandler struct {
	notificationRepo adminNotificationStore
}

type adminNotificationStore interface {
	CountNewUserRegistrations(context.Context, string) (int, error)
	AcknowledgeUserRegistrations(context.Context, string) error
}

func NewAdminNotificationHandler(notificationRepo *repository.AdminNotificationRepository) *AdminNotificationHandler {
	return &AdminNotificationHandler{notificationRepo: notificationRepo}
}

func (h *AdminNotificationHandler) GetUserRegistrations(w http.ResponseWriter, r *http.Request) {
	admin := middleware.GetUserFromContext(r.Context())
	if admin == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	count, err := h.notificationRepo.CountNewUserRegistrations(r.Context(), admin.ID)
	if err != nil {
		logAdminNotificationEvent(r.Header.Get("X-Request-ID"), "admin_notifications.count.error", "admin_user_id="+admin.ID, "error="+err.Error())
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load notification count"})
		return
	}

	logAdminNotificationEvent(r.Header.Get("X-Request-ID"), "admin_notifications.count", "admin_user_id="+admin.ID, "new_user_registrations="+itoa(count))
	writeJSON(w, http.StatusOK, map[string]int{"new_user_registrations": count})
}

func (h *AdminNotificationHandler) AcknowledgeUserRegistrations(w http.ResponseWriter, r *http.Request) {
	admin := middleware.GetUserFromContext(r.Context())
	if admin == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := h.notificationRepo.AcknowledgeUserRegistrations(r.Context(), admin.ID); err != nil {
		logAdminNotificationEvent(r.Header.Get("X-Request-ID"), "admin_notifications.ack.error", "admin_user_id="+admin.ID, "error="+err.Error())
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to acknowledge notifications"})
		return
	}

	logAdminNotificationEvent(r.Header.Get("X-Request-ID"), "admin_notifications.ack", "admin_user_id="+admin.ID)
	w.WriteHeader(http.StatusNoContent)
}

func logAdminNotificationEvent(requestID, event string, fields ...string) {
	parts := []string{"event=" + event}
	if requestID != "" {
		parts = append(parts, "request_id="+requestID)
	}
	parts = append(parts, fields...)
	log.Print(strings.Join(parts, " "))
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
