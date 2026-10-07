-- +goose Up
CREATE TABLE IF NOT EXISTS user_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('ext', 'work_phone', 'personal_phone')),
    value TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'all' CHECK (visibility IN ('all', 'unit', 'custom', 'nobody')),
    sort INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, type, value)
);

CREATE INDEX idx_user_contacts_user_id ON user_contacts(user_id);
CREATE INDEX idx_user_contacts_type ON user_contacts(type);

-- Таблица для кастомной видимости контактов
CREATE TABLE IF NOT EXISTS user_contact_acl (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contact_id UUID NOT NULL REFERENCES user_contacts(id) ON DELETE CASCADE,
    subject_type TEXT NOT NULL CHECK (subject_type IN ('user', 'group', 'node')),
    subject_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(contact_id, subject_type, subject_id)
);

CREATE INDEX idx_user_contact_acl_contact_id ON user_contact_acl(contact_id);
CREATE INDEX idx_user_contact_acl_subject ON user_contact_acl(subject_type, subject_id);

-- +goose Down
DROP TABLE IF EXISTS user_contact_acl;
DROP TABLE IF EXISTS user_contacts;