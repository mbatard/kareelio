package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/user/kareelio/backend/internal/config"
	"github.com/user/kareelio/backend/internal/handler"
	"github.com/user/kareelio/backend/internal/mailer"
	"github.com/user/kareelio/backend/internal/middleware"
	"github.com/user/kareelio/backend/internal/model"
	"github.com/user/kareelio/backend/internal/repository"
)

func New(db *pgxpool.Pool, cfg *config.Config) *chi.Mux {
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db, cfg.SessionDurationHours)
	jaRepo := repository.NewJobApplicationRepository(db)
	adminDashRepo := repository.NewAdminDashboardRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	evRepo := repository.NewEmailVerificationRepository(db)

	m := mailer.New(cfg)

	authHandler := handler.NewAuthHandler(userRepo, sessionRepo, auditRepo, evRepo, m, cfg)
	userHandler := handler.NewUserHandler(userRepo, sessionRepo, auditRepo)
	profileHandler := handler.NewProfileHandler(userRepo, auditRepo)
	jaHandler := handler.NewJobApplicationHandler(jaRepo, auditRepo)
	csvHandler := handler.NewCSVHandler(jaRepo, auditRepo)
	adminDashHandler := handler.NewAdminDashboardHandler(adminDashRepo)
	auditHandler := handler.NewAuditHandler(auditRepo)
	aboutHandler := handler.NewAboutHandler()
	healthHandler := handler.NewHealthHandler(db)

	r := chi.NewRouter()

	r.Use(middleware.Panic)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(middleware.CORS(cfg.CorsOrigin))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.NormalizePath)
	r.Use(middleware.Timeout(30))

	rateLimiter := middleware.NewRateLimiter()

	r.Get("/api/healthz", healthHandler.Healthz)
	r.Get("/api/readyz", healthHandler.Readyz)

	r.With(rateLimiter.Limit(10, 1*time.Minute)).Post("/api/auth/login", authHandler.Login)
	r.With(rateLimiter.Limit(5, 1*time.Minute)).Post("/api/auth/register", authHandler.Register)
	r.With(rateLimiter.Limit(10, 1*time.Minute)).Post("/api/auth/verify-email", authHandler.VerifyEmail)
	r.With(rateLimiter.Limit(5, 1*time.Minute)).Post("/api/auth/resend-verification", authHandler.ResendVerification)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(sessionRepo, userRepo))
		r.Use(middleware.AuditCapture)
		r.Use(middleware.CSRFProtection(cfg.CorsOrigin))

		r.Post("/api/auth/logout", authHandler.Logout)
		r.Get("/api/auth/me", authHandler.Me)

		r.Get("/api/profile", profileHandler.Get)
		r.Put("/api/profile", profileHandler.Update)
		r.Put("/api/profile/password", profileHandler.ChangePassword)

		r.Get("/api/about", aboutHandler.Get)

		r.Route("/api/job-applications", func(r chi.Router) {
			r.Use(middleware.RequireRole(model.RoleUser))
			r.Get("/", jaHandler.List)
			r.Post("/", jaHandler.Create)
			r.Get("/export", csvHandler.Export)
			r.Post("/import", csvHandler.Import)
			r.Get("/{id}", jaHandler.Get)
			r.Put("/{id}", jaHandler.Update)
			r.Delete("/{id}", jaHandler.Delete)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(model.RoleAdmin))

			r.Get("/api/admin/dashboard", adminDashHandler.Get)
			r.Get("/api/admin/audit", auditHandler.List)

			r.Route("/api/users", func(r chi.Router) {
				r.Get("/", userHandler.List)
				r.Post("/", userHandler.Create)
				r.Get("/{id}", userHandler.Get)
			r.Put("/{id}", userHandler.Update)
			r.Delete("/{id}", userHandler.Delete)
			r.Put("/{id}/password", userHandler.ChangePassword)
			})
		})
	})

	return r
}
