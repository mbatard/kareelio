package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
