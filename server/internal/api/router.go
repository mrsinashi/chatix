// internal/api/router.go
package api

import (
	"log/slog"
	"net/http"

	"chatix/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(db *storage.Postgres, cache *storage.Valkey, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	// Простой логгер на базе slog
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("request", "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})
	r.Use(middleware.Recoverer)

	r.Get("/healthz", HealthHandler(db, cache))
	r.Get("/readyz", HealthHandler(db, cache))

	// Здесь позже добавим /api/v1/...
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("pong"))
		})
	})

	return r
}