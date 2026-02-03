package dto

import (
	"time"

	"github.com/plastinin/docrecognizer/internal/domain"
)

// CreateTaskRequest запрос на создание задачи
// Файл передаётся через multipart/form-data
type CreateTaskRequest struct {
	Schema []string `json:"schema"` // Поля для извлечения
}

// TaskResponse ответ с информацией о задаче
type TaskResponse struct {
	ID          string         `json:"id"`
	Status      string         `json:"status"`
	FileName    string         `json:"file_name"`
	ContentType string         `json:"content_type"`
	Schema      []string       `json:"schema"`
	Result      map[string]any `json:"result,omitempty"`
	Error       string         `json:"error,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

// TaskFromDomain конвертирует доменную модель в DTO
func TaskFromDomain(task *domain.Task) *TaskResponse {
	return &TaskResponse{
		ID:          task.ID.String(),
		Status:      task.Status.String(),
		FileName:    task.FileName,
		ContentType: task.ContentType,
		Schema:      task.Schema,
		Result:      task.Result,
		Error:       task.Error,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		CompletedAt: task.CompletedAt,
	}
}

// TaskListResponse ответ со списком задач
type TaskListResponse struct {
	Tasks      []*TaskResponse `json:"tasks"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// TaskListFromDomain конвертирует результат списка в DTO
func TaskListFromDomain(result *domain.TaskListResult) *TaskListResponse {
	tasks := make([]*TaskResponse, len(result.Tasks))
	for i, task := range result.Tasks {
		tasks[i] = TaskFromDomain(task)
	}

	totalPages := result.Total / result.Pagination.PageSize
	if result.Total%result.Pagination.PageSize > 0 {
		totalPages++
	}

	return &TaskListResponse{
		Tasks:      tasks,
		Total:      result.Total,
		Page:       result.Pagination.Page,
		PageSize:   result.Pagination.PageSize,
		TotalPages: totalPages,
	}
}