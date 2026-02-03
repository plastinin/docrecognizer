-- Расширение для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enum для статусов задачи
CREATE TYPE task_status AS ENUM ('pending', 'processing', 'completed', 'failed');

-- Таблица задач
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status task_status NOT NULL DEFAULT 'pending',
    file_key VARCHAR(512) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    schema TEXT[] NOT NULL,
    result JSONB,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Индексы
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_created_at ON tasks(created_at DESC);
CREATE INDEX idx_tasks_status_created_at ON tasks(status, created_at DESC);

-- Комментарии
COMMENT ON TABLE tasks IS 'Задачи на распознавание документов';
COMMENT ON COLUMN tasks.id IS 'Уникальный идентификатор задачи';
COMMENT ON COLUMN tasks.status IS 'Статус задачи: pending, processing, completed, failed';
COMMENT ON COLUMN tasks.file_key IS 'Ключ файла в S3 хранилище';
COMMENT ON COLUMN tasks.file_name IS 'Оригинальное имя загруженного файла';
COMMENT ON COLUMN tasks.content_type IS 'MIME тип файла';
COMMENT ON COLUMN tasks.schema IS 'Массив полей для извлечения из документа';
COMMENT ON COLUMN tasks.result IS 'Результат распознавания в формате JSON';
COMMENT ON COLUMN tasks.error IS 'Текст ошибки при неудачной обработке';