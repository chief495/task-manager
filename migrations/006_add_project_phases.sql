-- Этапы работ внутри проекта. Не заменяют глобальные stages проекта.
CREATE TABLE IF NOT EXISTS project_phases (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    position INTEGER NOT NULL DEFAULT 0,
    responsible_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_project_phases_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_project_phases_responsible
        FOREIGN KEY (responsible_id)
        REFERENCES users(id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_project_phases_deleted_at ON project_phases(deleted_at);
CREATE INDEX IF NOT EXISTS idx_project_phases_project_id ON project_phases(project_id);
CREATE INDEX IF NOT EXISTS idx_project_phases_responsible_id ON project_phases(responsible_id);
CREATE INDEX IF NOT EXISTS idx_project_phases_project_position ON project_phases(project_id, position);

-- Команда этапа проекта
CREATE TABLE IF NOT EXISTS project_phase_members (
    id SERIAL PRIMARY KEY,
    project_phase_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_project_phase_members_phase
        FOREIGN KEY (project_phase_id)
        REFERENCES project_phases(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_project_phase_members_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT uc_project_phase_members_phase_user
        UNIQUE (project_phase_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_project_phase_members_deleted_at ON project_phase_members(deleted_at);
CREATE INDEX IF NOT EXISTS idx_project_phase_members_phase_id ON project_phase_members(project_phase_id);
CREATE INDEX IF NOT EXISTS idx_project_phase_members_user_id ON project_phase_members(user_id);

-- Привязка задач к этапу. NULL сохраняет обратную совместимость для старых задач.
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS project_phase_id INTEGER;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_tasks_project_phase'
    ) THEN
        ALTER TABLE tasks
            ADD CONSTRAINT fk_tasks_project_phase
            FOREIGN KEY (project_phase_id)
            REFERENCES project_phases(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_tasks_project_phase_id ON tasks(project_phase_id);
