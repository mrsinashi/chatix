package api

import (
	"log/slog"
	"net/http"
	"time"

	"chatix/internal/api/handlers"
	"chatix/internal/repository"
	"chatix/internal/service"
	"chatix/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(db *storage.Postgres, cache *storage.Valkey, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("request", "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})
	r.Use(middleware.Recoverer)

	// Репозитории
	userRepo := repository.NewUserRepository(db.Pool)
	sessionRepo := repository.NewSessionRepository(db.Pool)
	roleRepo := repository.NewRoleRepository(db.Pool)
	auditRepo := repository.NewAuditRepository(db.Pool)

	settingService := service.NewSettingServiceV2(db.Pool)
	settingHandlers := handlers.NewSettingHandlers(settingService)

	// Сервисы
	authService := service.NewAuthService(
		userRepo, sessionRepo, roleRepo, auditRepo,
		cache.Client,
		5,                // maxAttempts
		5*time.Minute,    // lockDuration
		24*time.Hour,     // sessionTTL
	)
	userService := service.NewUserService(userRepo, roleRepo, auditRepo)

	// Обработчики
	authHandlers := handlers.NewAuthHandlers(authService, userService)

	// Технические эндпоинты
	r.Get("/healthz", HealthHandler(db, cache))
	r.Get("/readyz", HealthHandler(db, cache))

	// API
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/settings/system", settingHandlers.GetPublicSettings)
		r.Post("/auth/login", authHandlers.Login)
		r.Post("/auth/logout", authHandlers.Logout)
		r.Get("/auth/me", authHandlers.Me)
	})

	return r
}