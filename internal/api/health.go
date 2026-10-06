// internal/api/health.go
package api

import (
	"context"
	"net/http"
	"time"

	"chatix/internal/storage"
)

func HealthHandler(db *storage.Postgres, cache *storage.Valkey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.Pool.Ping(ctx); err != nil {
			http.Error(w, "db error", http.StatusServiceUnavailable)
			return
		}

		// Проверка Valkey (простой ping)
		if err := cache.Client.Do(ctx, cache.Client.B().Ping().Build()).Error(); err != nil {
			http.Error(w, "cache error", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}