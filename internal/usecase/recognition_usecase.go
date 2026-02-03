package usecase

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/plastinin/docrecognizer/internal/domain"
	"go.uber.org/zap"
)

// RecognitionUseCase бизнес-логика распознавания документов
type RecognitionUseCase struct {
	taskRepo     TaskRepository
	fileStorage  FileStorage
	llmClient    LLMClient
	pdfConverter PDFConverter
	logger       *zap.Logger
}

// NewRecognitionUseCase создаёт новый экземпляр RecognitionUseCase
func NewRecognitionUseCase(
	taskRepo TaskRepository,
	fileStorage FileStorage,
	llmClient LLMClient,
	pdfConverter PDFConverter,
	logger *zap.Logger,
) *RecognitionUseCase {
	return &RecognitionUseCase{
		taskRepo:     taskRepo,
		fileStorage:  fileStorage,
		llmClient:    llmClient,
		pdfConverter: pdfConverter,
		logger:       logger,
	}
}

// ProcessTask обрабатывает задачу распознавания
func (uc *RecognitionUseCase) ProcessTask(ctx context.Context, taskID uuid.UUID) error {
	uc.logger.Info("Starting task processing",
		zap.String("task_id", taskID.String()),
	)

	// Получаем задачу
	task, err := uc.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Проверяем статус
	if task.Status.IsFinal() {
		uc.logger.Warn("Task already in final status, skipping",
			zap.String("task_id", taskID.String()),
			zap.String("status", task.Status.String()),
		)
		return nil
	}

	// Переводим в статус "processing"
	if err := task.MarkProcessing(); err != nil {
		return fmt.Errorf("failed to mark task as processing: %w", err)
	}
	if err := uc.taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Скачиваем файл из S3
	fileReader, err := uc.fileStorage.Download(ctx, task.FileKey)
	if err != nil {
		uc.markTaskFailed(ctx, task, fmt.Sprintf("failed to download file: %v", err))
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer fileReader.Close()

	// Читаем содержимое файла
	fileData, err := io.ReadAll(fileReader)
	if err != nil {
		uc.markTaskFailed(ctx, task, fmt.Sprintf("failed to read file: %v", err))
		return fmt.Errorf("failed to read file: %w", err)
	}

	uc.logger.Debug("File downloaded from storage",
		zap.String("task_id", taskID.String()),
		zap.Int("file_size", len(fileData)),
		zap.String("content_type", task.ContentType),
	)

	// Подготавливаем изображение для LLM
	imageData, err := uc.prepareImageData(fileData, task.ContentType)
	if err != nil {
		uc.markTaskFailed(ctx, task, fmt.Sprintf("failed to prepare image: %v", err))
		return fmt.Errorf("failed to prepare image: %w", err)
	}

	// Отправляем на распознавание в LLM
	result, err := uc.llmClient.RecognizeDocument(ctx, imageData, "image/png", task.Schema)
	if err != nil {
		uc.markTaskFailed(ctx, task, fmt.Sprintf("LLM recognition failed: %v", err))
		return fmt.Errorf("LLM recognition failed: %w", err)
	}

	// Успешно завершаем задачу
	if err := task.MarkCompleted(result); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}
	if err := uc.taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	uc.logger.Info("Task completed successfully",
		zap.String("task_id", taskID.String()),
		zap.Any("result", result),
	)

	return nil
}

// prepareImageData подготавливает изображение для отправки в LLM
func (uc *RecognitionUseCase) prepareImageData(fileData []byte, contentType string) ([]byte, error) {
	// Если это PDF — конвертируем первую страницу в изображение
	if strings.Contains(strings.ToLower(contentType), "pdf") {
		if uc.pdfConverter == nil {
			return nil, fmt.Errorf("PDF converter not available")
		}
		
		imageData, err := uc.pdfConverter.ConvertFirstPage(fileData)
		if err != nil {
			return nil, fmt.Errorf("failed to convert PDF: %w", err)
		}
		return imageData, nil
	}

	// Для изображений возвращаем как есть
	return fileData, nil
}

// markTaskFailed помечает задачу как неудачную
func (uc *RecognitionUseCase) markTaskFailed(ctx context.Context, task *domain.Task, errMsg string) {
	uc.logger.Error("Task processing failed",
		zap.String("task_id", task.ID.String()),
		zap.String("error", errMsg),
	)

	if err := task.MarkFailed(errMsg); err != nil {
		uc.logger.Error("Failed to mark task as failed",
			zap.String("task_id", task.ID.String()),
			zap.Error(err),
		)
		return
	}

	if err := uc.taskRepo.Update(ctx, task); err != nil {
		uc.logger.Error("Failed to update failed task",
			zap.String("task_id", task.ID.String()),
			zap.Error(err),
		)
	}
}