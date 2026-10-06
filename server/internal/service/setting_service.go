package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingService struct {
	pool *pgxpool.Pool
}

func NewSettingService(pool *pgxpool.Pool) *SettingService {
	return &SettingService{pool: pool}
}

// GetSetting возвращает настройку по ключу (глобальную или пользовательскую)
func (s *SettingService) GetSetting(ctx context.Context, key string, userID *string) (json.RawMessage, error) {
	query := `
		SELECT value FROM settings
		WHERE key = $1 AND ($2::uuid IS NULL OR user_id = $2)
		ORDER BY user_id DESC NULLS LAST
		LIMIT 1
	`

	var value json.RawMessage
	err := s.pool.QueryRow(ctx, query, key, userID).Scan(&value)
	if err != nil {
		return nil, fmt.Errorf("получение настройки: %w", err)
	}

	return value, nil
}

// SetSetting устанавливает настройку (глобальную или пользовательскую)
func (s *SettingService) SetSetting(ctx context.Context, key string, value json.RawMessage, userID *string) error {
	query := `
		INSERT INTO settings (key, value, user_id, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (key, user_id) DO UPDATE SET value = $2, updated_at = now()
	`

	_, err := s.pool.Exec(ctx, query, key, value, userID)
	if err != nil {
		return fmt.Errorf("установка настройки: %w", err)
	}

	return nil
}

// DeleteSetting удаляет настройку
func (s *SettingService) DeleteSetting(ctx context.Context, key string, userID *string) error {
	query := `
		DELETE FROM settings
		WHERE key = $1 AND ($2::uuid IS NULL OR user_id = $2)
	`

	_, err := s.pool.Exec(ctx, query, key, userID)
	if err != nil {
		return fmt.Errorf("удаление настройки: %w", err)
	}

	return nil
}