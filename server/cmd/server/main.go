package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chatix/internal/api"
	"chatix/internal/config"
	"chatix/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("не удалось загрузить конфиг", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	db, err := storage.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		slog.Error("ошибка подключения к PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	cache, err := storage.NewValkey(cfg.ValkeyAddr)
	if err != nil {
		slog.Error("ошибка подключения к Valkey", "error", err)
		os.Exit(1)
	}
	defer cache.Close()

	router := api.NewRouter(db, cache, logger)

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second, // Увеличено для загрузки файлов
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("сервер запущен", "addr", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("ошибка сервера", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("получен сигнал остановки, завершаем работу...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("ошибка при остановке сервера", "error", err)
	}
	slog.Info("сервер остановлен")
}