-- 006_setting_defs_values.sql
-- Новая система настроек по ТЗ (`setting_defs`, `setting_values`)

CREATE TABLE IF NOT EXISTS setting_defs (
    key VARCHAR(128) PRIMARY KEY,
    type VARCHAR(32) NOT NULL,
    default_value JSONB NOT NULL,
    scopes TEXT[] NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS setting_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(128) NOT NULL REFERENCES setting_defs(key) ON DELETE CASCADE,
    scope_type VARCHAR(16) NOT NULL CHECK (scope_type IN ('system', 'department', 'group', 'user', 'chat')),
    scope_id UUID,
    value JSONB NOT NULL,
    changed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (key, scope_type, scope_id)
);

CREATE INDEX IF NOT EXISTS idx_setting_values_scope ON setting_values(scope_type, scope_id);

-- Регистрируем настройки, которые уже используются
INSERT INTO setting_defs (key, type, default_value, scopes, description, category)
VALUES
    ('org_name', 'string', '"Название организации"', '{system}', 'Название организации, отображается в заголовке окна', 'Организация'),
    ('default_theme', 'string', '"light"', '{system}', 'Тема по умолчанию', 'Внешний вид'),
    ('password_min_length', 'integer', '8', '{system}', 'Минимальная длина пароля', 'Безопасность')
ON CONFLICT (key) DO NOTHING;

-- Переносим существующие системные настройки из `settings` в `setting_values`
INSERT INTO setting_values (key, scope_type, scope_id, value)
SELECT key, 'system', NULL, value
FROM settings
WHERE user_id IS NULL AND key IN (SELECT key FROM setting_defs)
ON CONFLICT (key, scope_type, scope_id) DO NOTHING;