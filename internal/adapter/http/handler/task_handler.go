package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/plastinin/docrecognizer/internal/adapter/http/dto"
	"github.com/plastinin/docrecognizer/internal/domain"
	"github.com/plastinin/docrecognizer/internal/usecase"
	"go.uber.org/zap"
)

const (
	maxUploadSize = 32 << 20 // 32 MB
)

// TaskHandler обработчик HTTP запросов для задач
type TaskHandler struct {
	taskUC *usecase.TaskUseCase
	logger *zap.Logger
}

// NewTaskHandler создаёт новый TaskHandler
func NewTaskHandler(taskUC *usecase.TaskUseCase, logger *zap.Logger) *TaskHandler {
	return &TaskHandler{
		taskUC: taskUC,
		logger: logger,
	}
}

// Create создаёт новую задачу
// POST /api/v1/tasks
// Content-Type: multipart/form-data
// - file: файл документа
// - schema: JSON массив полей для извлечения
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Ограничиваем размер загрузки
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Парсим multipart форму
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		h.logger.Warn("Failed to parse multipart form", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse form data")
		return
	}

	// Получаем файл
	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Warn("Failed to get file from form", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, "file_required", "File is required")
		return
	}
	defer file.Close()

	// Получаем schema
	schemaJSON := r.FormValue("schema")
	if schemaJSON == "" {
		h.respondError(w, http.StatusBadRequest, "schema_required", "Schema is required")
		return
	}

	var schema []string
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_schema", "Schema must be a JSON array of strings")
		return
	}

	if len(schema) == 0 {
		h.respondError(w, http.StatusBadRequest, "empty_schema", "Schema cannot be empty")
		return
	}

	// Определяем content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Пытаемся определить по расширению
		ct, err := domain.ContentTypeFromFileName(header.Filename)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid_file_type", "Unsupported file type")
			return
		}
		contentType = ct
	}

	// Валидируем тип файла
	if err := domain.ValidateContentType(contentType); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_file_type", "Unsupported file type. Supported: PNG, JPEG, WEBP, TIFF, PDF")
		return
	}

	// Создаём задачу
	input := usecase.CreateTaskInput{
		FileName:    header.Filename,
		ContentType: contentType,
		FileSize:    header.Size,
		FileReader:  file,
		Schema:      schema,
	}

	task, err := h.taskUC.Create(r.Context(), input)
	if err != nil {
		h.logger.Error("Failed to create task", zap.Error(err))
		
		if errors.Is(err, domain.ErrUnsupportedFileType) {
			h.respondError(w, http.StatusBadRequest, "invalid_file_type", err.Error())
			return
		}
		
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create task")
		return
	}

	h.respondJSON(w, http.StatusCreated, dto.TaskFromDomain(task))
}

// GetByID возвращает задачу по ID
// GET /api/v1/tasks/{id}
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid task ID format")
		return
	}

	task, err := h.taskUC.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			h.respondError(w, http.StatusNotFound, "not_found", "Task not found")
			return
		}
		h.logger.Error("Failed to get task", zap.String("task_id", idStr), zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to get task")
		return
	}

	h.respondJSON(w, http.StatusOK, dto.TaskFromDomain(task))
}

// List возвращает список задач
// GET /api/v1/tasks?page=1&page_size=20&status=pending
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	// Парсим параметры пагинации
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	pagination := domain.NewPagination(page, pageSize)

	// Парсим фильтры
	filter := domain.TaskFilter{}
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := domain.TaskStatus(statusStr)
		if status.IsValid() {
			filter.Status = &status
		}
	}

	result, err := h.taskUC.List(r.Context(), filter, pagination)
	if err != nil {
		h.logger.Error("Failed to list tasks", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to list tasks")
		return
	}

	h.respondJSON(w, http.StatusOK, dto.TaskListFromDomain(result))
}

// Delete удаляет задачу
// DELETE /api/v1/tasks/{id}
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid task ID format")
		return
	}

	err = h.taskUC.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			h.respondError(w, http.StatusNotFound, "not_found", "Task not found")
			return
		}
		h.logger.Error("Failed to delete task", zap.String("task_id", idStr), zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to delete task")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// respondJSON отправляет JSON ответ
func (h *TaskHandler) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// respondError отправляет ответ с ошибкой
func (h *TaskHandler) respondError(w http.ResponseWriter, status int, errCode string, message string) {
	h.respondJSON(w, status, dto.NewErrorResponse(errCode, message))
}