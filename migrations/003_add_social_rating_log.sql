-- Таблица для хранения истории изменения социального рейтинга
CREATE TABLE IF NOT EXISTS social_rating_logs (
                                                  id SERIAL PRIMARY KEY,
                                                  user_id INTEGER NOT NULL,
                                                  rating FLOAT NOT NULL,
                                                  change_reason TEXT,
                                                  created_at TIMESTAMPTZ DEFAULT NOW(),

    -- Внешний ключ
    CONSTRAINT fk_social_rating_logs_user
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE

    -- Индекс для быстрого поиска по пользователю и дате
    );
CREATE INDEX IF NOT EXISTS idx_social_rating_logs_user_date ON social_rating_logs(user_id, created_at);
-- Добавляем колонку deleted_at в таблицу social_rating_logs для поддержки soft delete
ALTER TABLE social_rating_logs
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Создаём индекс для колонки deleted_at, если нужно по производительности
CREATE INDEX IF NOT EXISTS idx_social_rating_logs_deleted_at ON social_rating_logs(deleted_at);