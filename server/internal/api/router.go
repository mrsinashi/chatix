package api

import (
	"log/slog"
	"net/http"
	"time"

	"chatix/internal/api/handlers"
	"chatix/internal/api/middleware"
	"chatix/internal/repository"
	"chatix/internal/service"
	"chatix/internal/storage"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(db *storage.Postgres, cache *storage.Valkey, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("request", "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})
	r.Use(chimiddleware.Recoverer)

	// Репозитории
	userRepo := repository.NewUserRepository(db.Pool)
	sessionRepo := repository.NewSessionRepository(db.Pool)
	roleRepo := repository.NewRoleRepository(db.Pool)
	auditRepo := repository.NewAuditRepository(db.Pool)

	settingService := service.NewSettingServiceV2(db.Pool)
	settingHandlers := handlers.NewSettingHandlers(settingService)
	settingsAdminHandlers := handlers.NewSettingsAdminHandlers(settingService)
	userSettingsHandlers := handlers.NewUserSettingsHandlers(settingService)

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

	// Пользовательские настройки (требуют авторизации)
	authMiddleware := middleware.Auth(authService)
	r.Route("/api/v1/settings", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/", userSettingsHandlers.GetUserSettings)
		r.Get("/{key}", userSettingsHandlers.GetUserSetting)
		r.Put("/{key}", userSettingsHandlers.SetUserSetting)
		r.Delete("/{key}", userSettingsHandlers.DeleteUserSetting)
	})

	// Админские эндпоинты (требуют авторизации и прав)
	r.Route("/api/v1/admin", func(r chi.Router) {
		r.Use(authMiddleware)

		// Просмотр настроек требует права `settings.view`
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(userService, "settings.view"))
			r.Get("/settings", settingsAdminHandlers.ListSettings)
			r.Get("/settings/{key}", settingsAdminHandlers.GetSetting)
		})

		// Управление настройками требует права `settings.manage`
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(userService, "settings.manage"))
			r.Put("/settings/{key}", settingsAdminHandlers.SetSetting)
			r.Delete("/settings/{key}", settingsAdminHandlers.DeleteSetting)
		})
	})

	return r
}