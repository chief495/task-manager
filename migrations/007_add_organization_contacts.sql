-- Контакты из карточек компаний, не привязанные к конкретному проекту
CREATE TABLE IF NOT EXISTS organization_contacts (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL,
    full_name VARCHAR(200) NOT NULL,
    position VARCHAR(200),
    phone VARCHAR(50),
    email VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_organization_contacts_organization
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_organization_contacts_deleted_at ON organization_contacts(deleted_at);
CREATE INDEX IF NOT EXISTS idx_organization_contacts_organization_id ON organization_contacts(organization_id);
