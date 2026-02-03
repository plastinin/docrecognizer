package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/plastinin/docrecognizer/internal/adapter/http/handler"
	httpmiddleware "github.com/plastinin/docrecognizer/internal/adapter/http/middleware"
	"go.uber.org/zap"
)

// NewRouter создаёт и настраивает HTTP роутер
func NewRouter(
	taskHandler *handler.TaskHandler,
	healthHandler *handler.HealthHandler,
	logger *zap.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httpmiddleware.NewLoggingMiddleware(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// Health check (вне версионирования API)
	r.Get("/health", healthHandler.Check)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Tasks
		r.Route("/tasks", func(r chi.Router) {
			r.Post("/", taskHandler.Create)
			r.Get("/", taskHandler.List)
			r.Get("/{id}", taskHandler.GetByID)
			r.Delete("/{id}", taskHandler.Delete)
		})
	})

	return r
}