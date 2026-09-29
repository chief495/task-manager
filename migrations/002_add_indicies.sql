-- Индексы для часто запрашиваемых полей
CREATE INDEX IF NOT EXISTS idx_tasks_status_id ON tasks(status_id);
CREATE INDEX IF NOT EXISTS idx_projects_responsible_id ON projects(responsible_id);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id);
