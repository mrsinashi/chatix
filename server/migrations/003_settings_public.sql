-- 003_settings_public.sql
-- Расширение настроек: добавляем поле публичности для системных настроек,
-- которые можно отдавать клиенту без авторизации (например, название организации).

ALTER TABLE settings
    ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_settings_public ON settings(is_public) WHERE is_public = TRUE;

-- Начальные системные настройки
INSERT INTO settings (key, value, user_id, is_public, updated_at)
VALUES
    ('org_name', '"Название организации"', NULL, TRUE, now()),
    ('default_theme', '"light"', NULL, TRUE, now()),
    ('password_min_length', '8', NULL, FALSE, now())
ON CONFLICT (key, user_id) DO UPDATE SET is_public = EXCLUDED.is_public, updated_at = now();