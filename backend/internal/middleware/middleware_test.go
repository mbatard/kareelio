package middleware

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/kareelio/backend/internal/model"
)

func TestLoggingSkipsHealthChecks(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(originalOutput) })

	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, path := range []string{"/api/healthz", "/api/readyz"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	if got := buf.String(); got != "" {
		t.Fatalf("health check logs = %q, want empty", got)
	}
}

func TestLoggingLogsNonHealthRequests(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(originalOutput) })

	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	got := buf.String()
	if !strings.Contains(got, "POST /api/auth/register 201") {
		t.Fatalf("request log = %q, want register request", got)
	}
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{name: "unauthorized", user: nil, wantStatus: http.StatusUnauthorized},
		{name: "forbidden", user: &model.User{ID: "user-1", Role: model.RoleUser}, wantStatus: http.StatusForbidden},
		{name: "allowed", user: &model.User{ID: "admin-1", Role: model.RoleAdmin}, wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequireRole(model.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "/api/admin/notifications", nil)
			if tt.user != nil {
				req = req.WithContext(context.WithValue(req.Context(), UserKey, tt.user))
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}
