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

// GetPublicSettings возвращает все публичные системные настройки.
// Можно отдавать без авторизации.
func (s *SettingService) GetPublicSettings(ctx context.Context) (map[string]json.RawMessage, error) {
	query := `
		SELECT key, value
		FROM settings
		WHERE is_public = TRUE AND user_id IS NULL
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("получение публичных настроек: %w", err)
	}
	defer rows.Close()

	result := make(map[string]json.RawMessage)
	for rows.Next() {
		var key string
		var value json.RawMessage
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("чтение настройки: %w", err)
		}
		result[key] = value
	}

	return result, nil
}

// GetUserSettings возвращает все настройки для пользователя с резолвом:
// пользовательские переопределяют системные.
func (s *SettingService) GetUserSettings(ctx context.Context, userID string) (map[string]json.RawMessage, error) {
	query := `
		WITH all_settings AS (
			-- Системные настройки
			SELECT key, value, 1 AS priority
			FROM settings
			WHERE user_id IS NULL
			UNION ALL
			-- Пользовательские настройки
			SELECT key, value, 2 AS priority
			FROM settings
			WHERE user_id = $1
		)
		SELECT DISTINCT ON (key) key, value
		FROM all_settings
		ORDER BY key, priority DESC
	`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("получение настроек пользователя: %w", err)
	}
	defer rows.Close()

	result := make(map[string]json.RawMessage)
	for rows.Next() {
		var key string
		var value json.RawMessage
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("чтение настройки: %w", err)
		}
		result[key] = value
	}

	return result, nil
}

// SetGlobalSetting устанавливает глобальную (системную) настройку.
func (s *SettingService) SetGlobalSetting(ctx context.Context, key string, value json.RawMessage, isPublic bool) error {
	query := `
		INSERT INTO settings (key, value, user_id, is_public, updated_at)
		VALUES ($1, $2, NULL, $3, now())
		ON CONFLICT (key, user_id) DO UPDATE SET value = $2, is_public = $3, updated_at = now()
	`

	_, err := s.pool.Exec(ctx, query, key, value, isPublic)
	if err != nil {
		return fmt.Errorf("установка глобальной настройки: %w", err)
	}

	return nil
}

// SetUserSetting устанавливает пользовательскую настройку.
func (s *SettingService) SetUserSetting(ctx context.Context, userID, key string, value json.RawMessage) error {
	query := `
		INSERT INTO settings (key, value, user_id, is_public, updated_at)
		VALUES ($1, $2, $3, FALSE, now())
		ON CONFLICT (key, user_id) DO UPDATE SET value = $2, updated_at = now()
	`

	_, err := s.pool.Exec(ctx, query, key, value, userID)
	if err != nil {
		return fmt.Errorf("установка пользовательской настройки: %w", err)
	}

	return nil
}