-- 004_core_align_tz.sql
-- Приведение схемы данных в соответствие с ТЗ

-- Обновляем значения статуса в существующих данных
UPDATE users SET status = 'archived' WHERE status = 'inactive';
UPDATE users SET status = 'blocked' WHERE status = 'locked';

-- Изменяем тип статуса
ALTER TABLE users ALTER COLUMN status TYPE TEXT;
DROP TYPE user_status;
CREATE TYPE user_status AS ENUM ('active', 'blocked', 'archived');
ALTER TABLE users ALTER COLUMN status TYPE user_status USING status::user_status;

-- Добавляем недостающие поля в `users`
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS kind VARCHAR(16) NOT NULL DEFAULT 'person' CHECK (kind IN ('person', 'room', 'role')),
    ADD COLUMN IF NOT EXISTS room_text TEXT,
    ADD COLUMN IF NOT EXISTS role_title TEXT,
    ADD COLUMN IF NOT EXISTS avatar_file_id UUID,
    ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ;

-- Добавляем недостающие поля в `sessions`
ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS device_info TEXT,
    ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ;

-- Добавляем поле `scope_node_id` в `user_roles`
ALTER TABLE user_roles
    ADD COLUMN IF NOT EXISTS scope_node_id UUID;

-- Пересчитываем нормализованные логины по новым правилам
-- (без учёта регистра, «ё» → «е», без лишних пробелов)
UPDATE users
SET username_normalized = lower(replace(trim(username), 'ё', 'е'));