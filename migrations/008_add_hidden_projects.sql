-- Скрытые проекты и список пользователей с доступом к ним
ALTER TABLE projects ADD COLUMN IF NOT EXISTS is_hidden BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS project_accesses (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_project_accesses_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_project_accesses_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT uc_project_accesses_project_user
        UNIQUE (project_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_projects_is_hidden ON projects(is_hidden);
CREATE INDEX IF NOT EXISTS idx_project_accesses_deleted_at ON project_accesses(deleted_at);
CREATE INDEX IF NOT EXISTS idx_project_accesses_project_id ON project_accesses(project_id);
CREATE INDEX IF NOT EXISTS idx_project_accesses_user_id ON project_accesses(user_id);
