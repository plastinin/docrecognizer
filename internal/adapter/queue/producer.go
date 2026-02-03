package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/plastinin/docrecognizer/internal/config"
)

// Типы задач
const (
	TypeDocumentRecognition = "document:recognition"
)

// DocumentRecognitionPayload данные задачи на распознавание
type DocumentRecognitionPayload struct {
	TaskID string `json:"task_id"`
}

// TaskProducer отправляет задачи в очередь
type TaskProducer struct {
	client *asynq.Client
}

// NewTaskProducer создаёт новый экземпляр TaskProducer
func NewTaskProducer(cfg config.RedisConfig) *TaskProducer {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &TaskProducer{client: client}
}

// Enqueue добавляет задачу в очередь
func (p *TaskProducer) Enqueue(ctx context.Context, taskID uuid.UUID) error {
	payload, err := json.Marshal(DocumentRecognitionPayload{
		TaskID: taskID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeDocumentRecognition, payload,
		asynq.MaxRetry(3),              // Максимум 3 попытки
		asynq.Queue("recognition"),     // Очередь для распознавания
	)

	_, err = p.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// Close закрывает соединение
func (p *TaskProducer) Close() error {
	return p.client.Close()
}