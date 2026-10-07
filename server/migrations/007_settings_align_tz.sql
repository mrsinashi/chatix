-- 007_settings_align_tz.sql
-- Приведение ключей настроек к формату ТЗ и исправление уровней

BEGIN;

-- Временно отключаем внешний ключ
ALTER TABLE setting_values DROP CONSTRAINT setting_values_key_fkey;

-- Обновляем ключи в обеих таблицах
UPDATE setting_values SET key = 'org.name' WHERE key = 'org_name';
UPDATE setting_values SET key = 'client.theme' WHERE key = 'default_theme';
UPDATE setting_values SET key = 'auth.password_min_len' WHERE key = 'password_min_length';

UPDATE setting_defs SET key = 'org.name' WHERE key = 'org_name';
UPDATE setting_defs SET key = 'client.theme' WHERE key = 'default_theme';
UPDATE setting_defs SET key = 'auth.password_min_len' WHERE key = 'password_min_length';

-- Восстанавливаем внешний ключ
ALTER TABLE setting_values ADD CONSTRAINT setting_values_key_fkey
    FOREIGN KEY (key) REFERENCES setting_defs(key) ON DELETE CASCADE;

-- Исправляем допустимые уровни по ТЗ
UPDATE setting_defs SET scopes = '{system,user}' WHERE key = 'client.theme';
UPDATE setting_defs SET scopes = '{system}' WHERE key = 'auth.password_min_len';
UPDATE setting_defs SET scopes = '{system}' WHERE key = 'org.name';

-- Добавляем обязательные настройки по ТЗ раздел 5
INSERT INTO setting_defs (key, type, default_value, scopes, description, category) VALUES
    ('message.edit_window_sec', 'integer', '900', '{system,department,group,user}', 'Окно редактирования сообщений в секундах', 'Сообщения'),
    ('message.delete_window_sec', 'integer', '900', '{system,department,group,user}', 'Окно удаления сообщений в секундах', 'Сообщения'),
    ('message.delete_mode', 'string', '"soft"', '{system,department,group,user}', 'Режим удаления: мягкое с пометкой', 'Сообщения'),
    ('message.max_length', 'integer', '8000', '{system}', 'Максимальная длина сообщения в символах', 'Сообщения'),
    ('presence.idle_after_sec', 'integer', '600', '{system,user}', 'Через сколько секунд бездействия статус меняется на «отошёл»', 'Статусы'),
    ('auth.session_ttl_days', 'integer', '0', '{system}', 'Срок жизни сессии в днях. 0 = бессрочно', 'Безопасность'),
    ('auth.must_change_password', 'boolean', 'false', '{system,user}', 'Требовать смену пароля при первом входе', 'Безопасность'),
    ('auth.login_pattern', 'string', '"^[а-яА-ЯёЁa-zA-Z0-9._-]+$"', '{system}', 'Допустимый шаблон логина', 'Безопасность'),
    ('auth.max_attempts', 'integer', '5', '{system}', 'Максимум неудачных попыток входа до блокировки', 'Безопасность'),
    ('auth.lock_minutes', 'integer', '15', '{system}', 'Длительность блокировки после превышения попыток входа', 'Безопасность'),
    ('contacts.ext_visibility', 'string', '"all"', '{system,user}', 'Видимость внутреннего номера', 'Контакты'),
    ('contacts.phone_visibility', 'string', '"unit"', '{system,user}', 'Видимость личного и рабочего номера', 'Контакты'),
    ('files.max_size_mb', 'integer', '4096', '{system,department,group,user}', 'Максимальный размер файла в МБ', 'Файлы'),
    ('files.server_threshold_mb', 'integer', '500', '{system}', 'Порог «большого» файла в МБ', 'Файлы'),
    ('files.large_max_ttl_days', 'integer', '7', '{system}', 'Максимальный срок хранения большого файла в днях', 'Файлы'),
    ('files.large_grace_hours', 'integer', '24', '{system}', 'Сколько часов хранить большой файл после скачивания всеми', 'Файлы'),
    ('files.compress_enabled', 'boolean', 'true', '{system}', 'Сжимать файлы при ротации', 'Файлы'),
    ('files.blocked_extensions', 'string', '"exe,bat,cmd,scr,js,vbs,ps1,msi"', '{system}', 'Опасные расширения файлов', 'Файлы'),
    ('files.direct_transfer_enabled', 'boolean', 'false', '{system}', 'Прямая передача файлов между клиентами', 'Файлы'),
    ('broadcast.require_ack_default', 'boolean', 'false', '{system}', 'Требовать подтверждение прочтения рассылок по умолчанию', 'Рассылки'),
    ('client.history_per_chat', 'integer', '500', '{user}', 'Сколько сообщений хранить локально для каждого чата', 'Клиент'),
    ('client.autostart', 'boolean', 'true', '{user}', 'Запускать вместе с системой', 'Клиент'),
    ('client.accent', 'string', '"blue"', '{system,user}', 'Акцентный цвет', 'Клиент'),
    ('client.radius', 'string', '"medium"', '{user}', 'Скругление углов: none, small, medium, large', 'Клиент'),
    ('client.scale', 'integer', '100', '{user}', 'Масштаб интерфейса в процентах (80-150)', 'Клиент'),
    ('client.density', 'string', '"comfortable"', '{user}', 'Плотность интерфейса: comfortable, compact', 'Клиент'),
    ('client.reduce_motion', 'string', '"auto"', '{user}', 'Анимации: auto, on, off', 'Клиент'),
    ('client.show_open_chats', 'boolean', 'true', '{user}', 'Показывать полосу открытых переписок', 'Клиент'),
    ('client.user_list_label', 'string', '"name"', '{user}', 'Подпись людей в списке: name, room, role', 'Клиент'),
    ('client.user_list_sort', 'string', '"online_first"', '{user}', 'Сортировка списка людей', 'Клиент'),
    ('client.pin_enabled', 'boolean', 'false', '{user}', 'PIN-блокировка включена', 'Клиент'),
    ('client.pin_auto_lock_min', 'integer', '5', '{user}', 'Автоблокировка по простою в минутах', 'Клиент'),
    ('client.copy_format', 'string', '"with_author_time"', '{user}', 'Формат копирования сообщений', 'Клиент')
ON CONFLICT (key) DO NOTHING;

COMMIT;