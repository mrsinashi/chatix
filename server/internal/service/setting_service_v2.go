package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingServiceV2 struct {
	pool *pgxpool.Pool
}

func NewSettingServiceV2(pool *pgxpool.Pool) *SettingServiceV2 {
	return &SettingServiceV2{pool: pool}
}

// GetSystemSettings возвращает все системные настройки (scope_type = 'system').
// Используется для публичного эндпоинта.
func (s *SettingServiceV2) GetSystemSettings(ctx context.Context) (map[string]json.RawMessage, error) {
	query := `
		SELECT sd.key, COALESCE(sv.value, sd.default_value)
		FROM setting_defs sd
		LEFT JOIN setting_values sv ON sv.key = sd.key AND sv.scope_type = 'system'
		WHERE 'system' = ANY(sd.scopes)
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("получение системных настроек: %w", err)
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

// GetUserSettings возвращает настройки для пользователя с резолвом:
// пользовательские переопределяют системные.
func (s *SettingServiceV2) GetUserSettings(ctx context.Context, userID string) (map[string]json.RawMessage, error) {
	query := `
		WITH user_settings AS (
			SELECT sv.key, sv.value
			FROM setting_values sv
			JOIN setting_defs sd ON sd.key = sv.key
			WHERE sv.scope_type = 'user' AND sv.scope_id = $1
		)
		SELECT sd.key, COALESCE(us.value, sd.default_value)
		FROM setting_defs sd
		LEFT JOIN user_settings us ON us.key = sd.key
		WHERE 'user' = ANY(sd.scopes)
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

// SetSystemSetting устанавливает системную настройку.
func (s *SettingServiceV2) SetSystemSetting(ctx context.Context, key string, value json.RawMessage, changedBy *string) error {
	query := `
		INSERT INTO setting_values (key, scope_type, scope_id, value, changed_by, changed_at)
		VALUES ($1, 'system', NULL, $2, $3, now())
		ON CONFLICT (key, scope_type, scope_id) DO UPDATE 
		SET value = $2, changed_by = $3, changed_at = now()
	`

	_, err := s.pool.Exec(ctx, query, key, value, changedBy)
	if err != nil {
		return fmt.Errorf("установка системной настройки: %w", err)
	}

	return nil
}

// SetUserSetting устанавливает пользовательскую настройку.
func (s *SettingServiceV2) SetUserSetting(ctx context.Context, userID, key string, value json.RawMessage) error {
	query := `
		INSERT INTO setting_values (key, scope_type, scope_id, value, changed_by, changed_at)
		VALUES ($1, 'user', $2, $3, $2, now())
		ON CONFLICT (key, scope_type, scope_id) DO UPDATE 
		SET value = $3, changed_by = $2, changed_at = now()
	`

	_, err := s.pool.Exec(ctx, query, key, userID, value)
	if err != nil {
		return fmt.Errorf("установка пользовательской настройки: %w", err)
	}

	return nil
}

// DeleteUserSetting удаляет пользовательскую настройку (возврат к дефолту).
func (s *SettingServiceV2) DeleteUserSetting(ctx context.Context, userID, key string) error {
	query := `
		DELETE FROM setting_values
		WHERE key = $1 AND scope_type = 'user' AND scope_id = $2
	`

	_, err := s.pool.Exec(ctx, query, key, userID)
	if err != nil {
		return fmt.Errorf("удаление пользовательской настройки: %w", err)
	}

	return nil
}