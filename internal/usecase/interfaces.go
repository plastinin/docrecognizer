package usecase

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/plastinin/docrecognizer/internal/domain"
)

// TaskRepository интерфейс для работы с хранилищем задач
type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.TaskFilter, pagination domain.Pagination) (*domain.TaskListResult, error)
}

// FileStorage интерфейс для работы с файловым хранилищем (S3)
type FileStorage interface {
	Upload(ctx context.Context, fileName string, contentType string, reader io.Reader, size int64) (fileKey string, err error)
	Download(ctx context.Context, fileKey string) (io.ReadCloser, error)
	Delete(ctx context.Context, fileKey string) error
	GetURL(ctx context.Context, fileKey string) (string, error)
}

// LLMClient интерфейс для работы с LLM (Ollama)
type LLMClient interface {
	RecognizeDocument(ctx context.Context, imageData []byte, contentType string, schema []string) (map[string]any, error)
}

// TaskQueue интерфейс для работы с очередью задач
type TaskQueue interface {
	Enqueue(ctx context.Context, taskID uuid.UUID) error
}

// PDFConverter интерфейс для конвертации PDF в изображения
type PDFConverter interface {
	ConvertFirstPage(pdfData []byte) ([]byte, error)
}