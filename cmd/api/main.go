package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/plastinin/docrecognizer/internal/adapter/http/handler"
	"github.com/plastinin/docrecognizer/internal/adapter/queue"
	"github.com/plastinin/docrecognizer/internal/adapter/repository"
	"github.com/plastinin/docrecognizer/internal/adapter/storage"
	"github.com/plastinin/docrecognizer/internal/config"
	"github.com/plastinin/docrecognizer/internal/usecase"
	"github.com/plastinin/docrecognizer/pkg/logger"
	"go.uber.org/zap"

	apphttp "github.com/plastinin/docrecognizer/internal/adapter/http"
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

	log.Info("Starting docrecognizer API",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
	)

	// Контекст с отменой для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// Инициализируем Queue Producer
	queueProducer := queue.NewTaskProducer(cfg.Redis)
	defer queueProducer.Close()
	log.Info("Connected to Redis",
		zap.String("addr", cfg.Redis.Addr()),
	)

	// Инициализируем репозитории
	taskRepo := repository.NewTaskRepository(dbPool)

	// Инициализируем use cases
	taskUC := usecase.NewTaskUseCase(taskRepo, s3Storage, queueProducer, log)

	// Инициализируем handlers
	taskHandler := handler.NewTaskHandler(taskUC, log)
	healthHandler := handler.NewHealthHandler()

	// Создаём роутер
	router := apphttp.NewRouter(taskHandler, healthHandler, log)

	// Создаём HTTP сервер
	server := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Запускаем сервер в горутине
	go func() {
		log.Info("HTTP server starting",
			zap.String("addr", cfg.Server.Addr()),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server stopped")
}