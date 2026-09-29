-- -----------------------------------------------------------------
-- МИГРАЦИЯ: Создание базовой схемы БД (001_init.sql)
-- -----------------------------------------------------------------

-- 1. Таблица users (пользователи)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,  -- хешированный пароль
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'manager', 'user')),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  -- для soft delete
);

-- Индекс по email для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- 2. Таблица stages (стадии проекта)
CREATE TABLE IF NOT EXISTS stages (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    position INT NOT NULL DEFAULT 0,  -- порядок в workflow
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Таблица projects (проекты)
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    start_date TIMESTAMPTZ NOT NULL,
    deadline_date TIMESTAMPTZ NOT NULL,
    responsible_id INT NOT NULL,      -- ответственный (связь с users)
    stage_id INT NOT NULL,           -- текущая стадия (связь с stages)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,          -- для soft delete

    -- Внешние ключи
    CONSTRAINT fk_projects_responsible
        FOREIGN KEY (responsible_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_projects_stage
        FOREIGN KEY (stage_id)
        REFERENCES stages(id)
        ON DELETE CASCADE
);

-- Индексы для проектов
CREATE INDEX IF NOT EXISTS idx_projects_responsible_id ON projects(responsible_id);
CREATE INDEX IF NOT EXISTS idx_projects_stage_id ON projects(stage_id);

-- 4. Таблица statuses (статусы задач)
CREATE TABLE IF NOT EXISTS statuses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    is_final BOOLEAN NOT NULL DEFAULT FALSE,  -- финальный статус?
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Таблица tasks (задачи)
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    parent_id INT,                    -- родительская задача (NULL для корневых)
    status_id INT NOT NULL,          -- текущий статус (связь с statuses)
    executor_id INT NOT NULL,        -- исполнитель (связь с users)
    project_id INT NOT NULL,         -- проект (связь с projects)
    start_time TIMESTAMPTZ,          -- плановое начало
    end_time TIMESTAMPTZ,            -- плановое завершение
    priority INT DEFAULT 0,          -- приоритет: 0=низкий, 1=средний, 2=высокий
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,          -- для soft delete

    -- Внешние ключи
    CONSTRAINT fk_tasks_parent
        FOREIGN KEY (parent_id)
        REFERENCES tasks(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_tasks_status
        FOREIGN KEY (status_id)
        REFERENCES statuses(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_tasks_executor
        FOREIGN KEY (executor_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_tasks_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE
);

-- Индексы для задач
CREATE INDEX IF NOT EXISTS idx_tasks_status_id ON tasks(status_id);
CREATE INDEX IF NOT EXISTS idx_tasks_executor_id ON tasks(executor_id);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id);

-- 6. Таблица stage_transitions (разрешённые переходы между стадиями)
CREATE TABLE IF NOT EXISTS stage_transitions (
    id SERIAL PRIMARY KEY,
    from_stage_id INT NOT NULL,
    to_stage_id INT NOT NULL,
    condition VARCHAR(200),  -- условие перехода (например, "all_tasks_completed")
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Уникальность комбинации from→to
    CONSTRAINT uc_stage_transition
        UNIQUE (from_stage_id, to_stage_id),

    -- Внешние ключи
    CONSTRAINT fk_stage_transitions_from
        FOREIGN KEY (from_stage_id)
        REFERENCES stages(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_stage_transitions_to
        FOREIGN KEY (to_stage_id)
        REFERENCES stages(id)
        ON DELETE CASCADE
);

-- 7. Таблица status_transitions (разрешённые переходы между статусами)
CREATE TABLE IF NOT EXISTS status_transitions (
    id SERIAL PRIMARY KEY,
    from_status_id INT NOT NULL,
    to_status_id INT NOT NULL,
    condition VARCHAR(200),  -- условие перехода (например, "approved_by_manager")
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Уникальность комбинации from→to
    CONSTRAINT uc_status_transition
        UNIQUE (from_status_id, to_status_id),

    -- Внешние ключи
    CONSTRAINT fk_status_transitions_from
        FOREIGN KEY (from_status_id)
        REFERENCES statuses(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_status_transitions_to
        FOREIGN KEY (to_status_id)
        REFERENCES statuses(id)
        ON DELETE CASCADE
);

-- 8. Таблица trips (командировки/выезды)
CREATE TABLE IF NOT EXISTS trips (
    id SERIAL PRIMARY KEY,
    place VARCHAR(200) NOT NULL,     -- место командировки
    project_id INT NOT NULL,        -- проект (связь с projects)
    person_id INT NOT NULL,          -- сотрудник (связь с users)
    begin_date TIMESTAMPTZ NOT NULL,  -- начало
    end_date TIMESTAMPTZ NOT NULL,    -- окончание
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Внешние ключи
    CONSTRAINT fk_trips_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_trips_person
        FOREIGN KEY (person_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Индексы для командировок
CREATE INDEX IF NOT EXISTS idx_trips_project_id ON trips(project_id);
CREATE INDEX IF NOT EXISTS idx_trips_person_id ON trips(person_id);
CREATE INDEX IF NOT EXISTS idx_trips_begin_date ON trips(begin_date);

-- -----------------------------------------------------------------
-- ИНДЕКСЫ ДЛЯ ПОИСКА И СОРТИРОВКИ
-- -----------------------------------------------------------------

-- Для быстрого поиска активных задач по исполнителю и статусу
CREATE INDEX IF NOT EXISTS idx_active_tasks
    ON tasks(executor_id, status_id)
    WHERE deleted_at IS NULL;

-- Для поиска проектов по сроку выполнения
CREATE INDEX IF NOT EXISTS idx_projects_deadline
    ON projects(deadline_date)
    WHERE deleted_at IS NULL;

-- Для поиска командировок в диапазоне дат
CREATE INDEX IF NOT EXISTS idx_trips_date_range
    ON trips(begin_date, end_date);
-- 9. Таблица reports (отчёты по проектам)
CREATE TABLE IF NOT EXISTS reports (
                                       id SERIAL PRIMARY KEY,
                                       user_id INT NOT NULL,             -- автор отчёта (связь с users)
                                       project_id INT,                   -- необязательная привязка к проекту
                                       text TEXT NOT NULL,               -- содержание отчёта
                                       overwork INTEGER DEFAULT 0,      -- переработка в часах (необязательно для заполнения),
                                       created_at TIMESTAMPTZ DEFAULT NOW(), -- дата создания
                                       updated_at TIMESTAMPTZ DEFAULT NOW(), -- дата обновления
                                       deleted_at TIMESTAMPTZ,           -- для soft delete

    -- Внешние ключи
                                       CONSTRAINT fk_reports_user
                                           FOREIGN KEY (user_id)
                                               REFERENCES users(id)
                                               ON DELETE CASCADE,

                                       CONSTRAINT fk_reports_project
                                           FOREIGN KEY (project_id)
                                               REFERENCES projects(id)
                                               ON DELETE SET NULL
);

-- Индекс для быстрого поиска отчётов по пользователю и проекту
CREATE INDEX IF NOT EXISTS idx_reports_user_id ON reports(user_id);
CREATE INDEX IF NOT EXISTS idx_reports_project_id ON reports(project_id);
CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at);
-- -----------------------------------------------------------------
-- ОГРАНИЧЕНИЯ И ПРОВЕРКИ
-- -----------------------------------------------------------------

-- Проверка: дата окончания командировки не раньше начала
