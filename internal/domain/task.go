package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Ошибки домена
var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrInvalidTaskStatus = errors.New("invalid task status")
	ErrEmptySchema       = errors.New("schema cannot be empty")
	ErrEmptyFileKey      = errors.New("file key cannot be empty")
)

// Task представляет задачу на распознавание документа
type Task struct {
	ID           uuid.UUID         `json:"id"`
	Status       TaskStatus        `json:"status"`
	FileKey      string            `json:"file_key"`       // Ключ файла в S3
	FileName     string            `json:"file_name"`      // Оригинальное имя файла
	ContentType  string            `json:"content_type"`   // MIME тип (image/png, application/pdf)
	Schema       []string          `json:"schema"`         // Поля для извлечения
	Result       map[string]any    `json:"result,omitempty"` // Результат распознавания
	Error        string            `json:"error,omitempty"`  // Текст ошибки (если failed)
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
}

// NewTask создаёт новую задачу
func NewTask(fileKey, fileName, contentType string, schema []string) (*Task, error) {
	if fileKey == "" {
		return nil, ErrEmptyFileKey
	}
	if len(schema) == 0 {
		return nil, ErrEmptySchema
	}

	now := time.Now()

	return &Task{
		ID:          uuid.New(),
		Status:      TaskStatusPending,
		FileKey:     fileKey,
		FileName:    fileName,
		ContentType: contentType,
		Schema:      schema,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// MarkProcessing переводит задачу в статус "в обработке"
func (t *Task) MarkProcessing() error {
	if t.Status != TaskStatusPending {
		return ErrInvalidTaskStatus
	}
	t.Status = TaskStatusProcessing
	t.UpdatedAt = time.Now()
	return nil
}

// MarkCompleted переводит задачу в статус "завершена"
func (t *Task) MarkCompleted(result map[string]any) error {
	if t.Status != TaskStatusProcessing {
		return ErrInvalidTaskStatus
	}
	now := time.Now()
	t.Status = TaskStatusCompleted
	t.Result = result
	t.UpdatedAt = now
	t.CompletedAt = &now
	return nil
}

// MarkFailed переводит задачу в статус "ошибка"
func (t *Task) MarkFailed(errMsg string) error {
	if t.Status != TaskStatusProcessing && t.Status != TaskStatusPending {
		return ErrInvalidTaskStatus
	}
	now := time.Now()
	t.Status = TaskStatusFailed
	t.Error = errMsg
	t.UpdatedAt = now
	t.CompletedAt = &now
	return nil
}

// CanRetry проверяет, можно ли повторить задачу
func (t *Task) CanRetry() bool {
	return t.Status == TaskStatusFailed
}