package usecase

import (
	"io"
)

// CreateTaskInput входные данные для создания задачи
type CreateTaskInput struct {
	FileName    string    // Имя файла
	ContentType string    // MIME тип
	FileSize    int64     // Размер файла
	FileReader  io.Reader // Содержимое файла
	Schema      []string  // Поля для извлечения
}

// ProcessTaskInput входные данные для обработки задачи воркером
type ProcessTaskInput struct {
	TaskID string
}