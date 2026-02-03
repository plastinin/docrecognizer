package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/plastinin/docrecognizer/internal/adapter/llm"
	"github.com/plastinin/docrecognizer/internal/adapter/queue"
	"github.com/plastinin/docrecognizer/internal/adapter/repository"
	"github.com/plastinin/docrecognizer/internal/adapter/storage"
	"github.com/plastinin/docrecognizer/internal/config"
	"github.com/plastinin/docrecognizer/internal/usecase"
	"github.com/plastinin/docrecognizer/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// Инициализируем логгер
	log := logger.Must(cfg.Log.Level, cfg.Log.Format)
	defer log.Sync()

	log.Info("Starting docrecognizer worker",
		zap.String("ollama_host", cfg.Ollama.Host),
		zap.String("ollama_model", cfg.Ollama.Model),
	)

	// Контекст для инициализации
	ctx := context.Background()

	// Инициализируем PostgreSQL
	dbPool, err := repository.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbPool.Close()
	log.Info("Connected to PostgreSQL")

	// Инициализируем S3 Storage
	s3Storage, err := storage.NewS3Storage(ctx, cfg.S3)
	if err != nil {
		log.Fatal("Failed to connect to S3", zap.Error(err))
	}
	log.Info("Connected to S3",
		zap.String("endpoint", cfg.S3.Endpoint),
		zap.String("bucket", cfg.S3.Bucket),
	)

	// Инициализируем Ollama клиент
	ollamaClient := llm.NewOllamaClient(cfg.Ollama, log)

	// Проверяем доступность Ollama
	if err := ollamaClient.CheckHealth(ctx); err != nil {
		log.Warn("Ollama health check failed", zap.Error(err))
		log.Warn("Make sure Ollama is running: ollama serve")
	} else {
		log.Info("Ollama is healthy")

		// Проверяем наличие модели
		if err := ollamaClient.CheckModel(ctx); err != nil {
			log.Warn("Model check failed", zap.Error(err))
		}
	}

	// Инициализируем PDF конвертер
	pdfConverter := llm.NewPDFConverter()

	// Инициализируем репозитории
	taskRepo := repository.NewTaskRepository(dbPool)

	// Инициализируем use cases
	recognitionUC := usecase.NewRecognitionUseCase(taskRepo, s3Storage, ollamaClient, pdfConverter, log)

	// Инициализируем consumer
	consumer := queue.NewTaskConsumer(cfg.Redis, recognitionUC, log)

	// Запускаем consumer в горутине
	go func() {
		if err := consumer.Start(); err != nil {
			log.Fatal("Failed to start consumer", zap.Error(err))
		}
	}()

	log.Info("Worker started, waiting for tasks...")

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down worker...")

	// Останавливаем consumer
	consumer.Stop()

	log.Info("Worker stopped")
}