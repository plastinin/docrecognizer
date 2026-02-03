package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/plastinin/docrecognizer/internal/config"
	"github.com/plastinin/docrecognizer/internal/usecase"
	"go.uber.org/zap"
)

// TaskConsumer обрабатывает задачи из очереди
type TaskConsumer struct {
	server            *asynq.Server
	mux               *asynq.ServeMux
	recognitionUC     *usecase.RecognitionUseCase
	logger            *zap.Logger
}

// NewTaskConsumer создаёт новый экземпляр TaskConsumer
func NewTaskConsumer(
	cfg config.RedisConfig,
	recognitionUC *usecase.RecognitionUseCase,
	logger *zap.Logger,
) *TaskConsumer {
	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Addr(),
			Password: cfg.Password,
			DB:       cfg.DB,
		},
		asynq.Config{
			Concurrency: 2, // Количество одновременных воркеров (для CPU режима лучше меньше)
			Queues: map[string]int{
				"recognition": 10, // Приоритет очереди
				"default":     1,
			},
			Logger: newAsynqLogger(logger),
		},
	)

	consumer := &TaskConsumer{
		server:        server,
		mux:           asynq.NewServeMux(),
		recognitionUC: recognitionUC,
		logger:        logger,
	}

	// Регистрируем обработчики
	consumer.mux.HandleFunc(TypeDocumentRecognition, consumer.handleDocumentRecognition)

	return consumer
}

// Start запускает обработку задач
func (c *TaskConsumer) Start() error {
	c.logger.Info("Starting task consumer")
	return c.server.Start(c.mux)
}

// Stop останавливает обработку задач
func (c *TaskConsumer) Stop() {
	c.logger.Info("Stopping task consumer")
	c.server.Stop()
	c.server.Shutdown()
}

// handleDocumentRecognition обрабатывает задачу распознавания документа
func (c *TaskConsumer) handleDocumentRecognition(ctx context.Context, t *asynq.Task) error {
	var payload DocumentRecognitionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		c.logger.Error("Failed to unmarshal payload",
			zap.Error(err),
			zap.ByteString("payload", t.Payload()),
		)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	taskID, err := uuid.Parse(payload.TaskID)
	if err != nil {
		c.logger.Error("Invalid task ID",
			zap.String("task_id", payload.TaskID),
			zap.Error(err),
		)
		return fmt.Errorf("invalid task ID: %w", err)
	}

	c.logger.Info("Processing document recognition task",
		zap.String("task_id", taskID.String()),
	)

	if err := c.recognitionUC.ProcessTask(ctx, taskID); err != nil {
		c.logger.Error("Failed to process task",
			zap.String("task_id", taskID.String()),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// asynqLogger адаптер логгера для asynq
type asynqLogger struct {
	logger *zap.Logger
}

func newAsynqLogger(logger *zap.Logger) *asynqLogger {
	return &asynqLogger{logger: logger.Named("asynq")}
}

func (l *asynqLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *asynqLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *asynqLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *asynqLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}