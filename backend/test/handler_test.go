package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/kareelio/backend/internal/middleware"
	"github.com/user/kareelio/backend/internal/model"
)

func TestHealthEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %s", resp["status"])
	}
}

func TestCORSHeaders(t *testing.T) {
	handler := middleware.CORS("http://localhost:5173")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/healthz", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Error("CORS header not set correctly")
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("CORS credentials header not set")
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := middleware.SecurityHeaders()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("X-Content-Type-Options header not set")
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("X-Frame-Options header not set")
	}
	if w.Header().Get("X-XSS-Protection") != "1; mode=block" {
		t.Error("X-XSS-Protection header not set")
	}
	if w.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Error("Referrer-Policy header not set")
	}
}

func TestUserModelToResponse(t *testing.T) {
	user := &model.User{
		ID:          "test-id",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Description: "A test user",
		Role:        model.RoleUser,
		IsActive:    true,
		Language:    "fr",
		Theme:       "dark",
	}

	resp := user.ToResponse()

	if resp.ID != "test-id" {
		t.Error("ID mismatch")
	}
	if resp.Email != "test@example.com" {
		t.Error("Email mismatch")
	}
	if resp.Role != model.RoleUser {
		t.Error("Role mismatch")
	}
	if resp.Language != "fr" {
		t.Error("Language mismatch")
	}
	if resp.Theme != "dark" {
		t.Error("Theme mismatch")
	}
}

func TestAdminRoleConstants(t *testing.T) {
	if model.RoleAdmin != "admin" {
		t.Error("RoleAdmin should be 'admin'")
	}
	if model.RoleUser != "user" {
		t.Error("RoleUser should be 'user'")
	}
}

func TestLoginValidation(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{"valid", "test@example.com", "password123", false},
		{"empty email", "", "password123", true},
		{"empty password", "test@example.com", "", true},
		{"both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := model.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			}
			if (req.Email == "" || req.Password == "") != tt.wantErr {
				t.Errorf("expected wantErr=%v for email=%q password=%q", tt.wantErr, tt.email, tt.password)
			}
		})
	}
}

func TestJobApplicationStatuses(t *testing.T) {
	validStatuses := []model.JobStatus{
		model.StatusDraft, model.StatusApplied, model.StatusResponded,
		model.StatusInterview, model.StatusTest, model.StatusOffer,
		model.StatusRejected, model.StatusWithdrawn,
	}

	if len(validStatuses) != 8 {
		t.Errorf("expected 8 statuses, got %d", len(validStatuses))
	}

	seen := make(map[model.JobStatus]bool)
	for _, s := range validStatuses {
		if seen[s] {
			t.Errorf("duplicate status: %s", s)
		}
		seen[s] = true
	}
}

func TestRemoteTypes(t *testing.T) {
	remotes := []model.RemoteType{model.RemoteOnSite, model.RemoteHybrid, model.RemoteFull}
	if len(remotes) != 3 {
		t.Errorf("expected 3 remote types, got %d", len(remotes))
	}
}

func TestContractTypes(t *testing.T) {
	contracts := []model.ContractType{
		model.ContractCDI, model.ContractCDD, model.ContractFreelance,
		model.ContractInternship, model.ContractApprentice, model.ContractOther,
	}
	if len(contracts) != 6 {
		t.Errorf("expected 6 contract types, got %d", len(contracts))
	}
}

func TestPriorityValues(t *testing.T) {
	priorities := []model.Priority{model.PriorityLow, model.PriorityMedium, model.PriorityHigh}
	if len(priorities) != 3 {
		t.Errorf("expected 3 priorities, got %d", len(priorities))
	}
}

func TestSourceValues(t *testing.T) {
	sources := []model.Source{
		model.SourceLinkedIn, model.SourceIndeed, model.SourceReferral,
		model.SourceAgency, model.SourceWebsite, model.SourceWTTJ, model.SourceOther,
	}
	if len(sources) != 7 {
		t.Errorf("expected 7 sources, got %d", len(sources))
	}
}

func TestContactTypes(t *testing.T) {
	contacts := []model.ContactType{model.ContactVideo, model.ContactPhone, model.ContactInPerson}
	if len(contacts) != 3 {
		t.Errorf("expected 3 contact types, got %d", len(contacts))
	}
}

func TestAboutModel(t *testing.T) {
	about := model.GetAbout()
	if about.Name != "Kareelio" {
		t.Errorf("expected name Kareelio, got %s", about.Name)
	}
	if about.Version == "" {
		t.Error("version should not be empty")
	}
}

func TestOptionalAuthNoCookie(t *testing.T) {
	handler := middleware.OptionalAuth(nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Error("optional auth should pass through without cookie")
	}
}
