package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/plastinin/docrecognizer/internal/domain"
	"go.uber.org/zap"
)

// TaskUseCase бизнес-логика работы с задачами
type TaskUseCase struct {
	taskRepo    TaskRepository
	fileStorage FileStorage
	taskQueue   TaskQueue
	logger      *zap.Logger
}

// NewTaskUseCase создаёт новый экземпляр TaskUseCase
func NewTaskUseCase(
	taskRepo TaskRepository,
	fileStorage FileStorage,
	taskQueue TaskQueue,
	logger *zap.Logger,
) *TaskUseCase {
	return &TaskUseCase{
		taskRepo:    taskRepo,
		fileStorage: fileStorage,
		taskQueue:   taskQueue,
		logger:      logger,
	}
}

// Create создаёт новую задачу на распознавание
func (uc *TaskUseCase) Create(ctx context.Context, input CreateTaskInput) (*domain.Task, error) {
	// Валидируем тип файла
	if err := domain.ValidateContentType(input.ContentType); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Загружаем файл в S3
	fileKey, err := uc.fileStorage.Upload(ctx, input.FileName, input.ContentType, input.FileReader, input.FileSize)
	if err != nil {
		uc.logger.Error("Failed to upload file to storage",
			zap.String("file_name", input.FileName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	uc.logger.Debug("File uploaded to storage",
		zap.String("file_key", fileKey),
		zap.String("file_name", input.FileName),
	)

	// Создаём задачу
	task, err := domain.NewTask(fileKey, input.FileName, input.ContentType, input.Schema)
	if err != nil {
		// Удаляем загруженный файл при ошибке
		_ = uc.fileStorage.Delete(ctx, fileKey)
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Сохраняем задачу в БД
	if err := uc.taskRepo.Create(ctx, task); err != nil {
		// Удаляем загруженный файл при ошибке
		_ = uc.fileStorage.Delete(ctx, fileKey)
		uc.logger.Error("Failed to save task to database",
			zap.String("task_id", task.ID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	// Добавляем задачу в очередь
	if err := uc.taskQueue.Enqueue(ctx, task.ID); err != nil {
		uc.logger.Error("Failed to enqueue task",
			zap.String("task_id", task.ID.String()),
			zap.Error(err),
		)
		// Не возвращаем ошибку — задача создана, можно retry позже
	}

	uc.logger.Info("Task created successfully",
		zap.String("task_id", task.ID.String()),
		zap.String("file_name", input.FileName),
		zap.Strings("schema", input.Schema),
	)

	return task, nil
}

// GetByID возвращает задачу по ID
func (uc *TaskUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	task, err := uc.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// List возвращает список задач
func (uc *TaskUseCase) List(ctx context.Context, filter domain.TaskFilter, pagination domain.Pagination) (*domain.TaskListResult, error) {
	return uc.taskRepo.List(ctx, filter, pagination)
}

// Delete удаляет задачу и связанный файл
func (uc *TaskUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	// Получаем задачу
	task, err := uc.taskRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Удаляем файл из S3
	if err := uc.fileStorage.Delete(ctx, task.FileKey); err != nil {
		uc.logger.Warn("Failed to delete file from storage",
			zap.String("task_id", id.String()),
			zap.String("file_key", task.FileKey),
			zap.Error(err),
		)
		// Продолжаем удаление задачи
	}

	// Удаляем задачу из БД
	if err := uc.taskRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	uc.logger.Info("Task deleted successfully",
		zap.String("task_id", id.String()),
	)

	return nil
}