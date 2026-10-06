-- 005_user_contacts.sql
-- Контакты пользователей (номера телефонов, видимость)

CREATE TABLE IF NOT EXISTS user_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(16) NOT NULL CHECK (type IN ('ext', 'work_phone', 'personal_phone')),
    value TEXT NOT NULL,
    visibility VARCHAR(16) NOT NULL DEFAULT 'unit' CHECK (visibility IN ('all', 'unit', 'custom', 'nobody')),
    sort INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_contacts_user ON user_contacts(user_id);

-- Таблица для видимости `custom`
CREATE TABLE IF NOT EXISTS user_contact_acl (
    contact_id UUID NOT NULL REFERENCES user_contacts(id) ON DELETE CASCADE,
    subject_type VARCHAR(16) NOT NULL CHECK (subject_type IN ('user', 'group', 'node')),
    subject_id UUID NOT NULL,
    PRIMARY KEY (contact_id, subject_type, subject_id)
);