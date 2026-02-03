package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plastinin/docrecognizer/internal/domain"
)

// TaskRepository реализация репозитория задач для PostgreSQL
type TaskRepository struct {
	pool *pgxpool.Pool
}

// NewTaskRepository создаёт новый экземпляр TaskRepository
func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

// Create создаёт новую задачу в БД
func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (id, status, file_key, file_name, content_type, schema, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		task.ID,
		task.Status,
		task.FileKey,
		task.FileName,
		task.ContentType,
		task.Schema,
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	return nil
}

// GetByID возвращает задачу по ID
func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	query := `
		SELECT id, status, file_key, file_name, content_type, schema, result, error, created_at, updated_at, completed_at
		FROM tasks
		WHERE id = $1
	`

	task := &domain.Task{}
	var errorMsg *string  // Указатель для NULL

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.Status,
		&task.FileKey,
		&task.FileName,
		&task.ContentType,
		&task.Schema,
		&task.Result,
		&errorMsg,  // Сканируем в указатель
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Обрабатываем NULL
	if errorMsg != nil {
		task.Error = *errorMsg
	}

	return task, nil
}

// Update обновляет задачу в БД
func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE tasks
		SET status = $2, result = $3, error = $4, updated_at = $5, completed_at = $6
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		task.ID,
		task.Status,
		task.Result,
		task.Error,
		task.UpdatedAt,
		task.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

// Delete удаляет задачу из БД
func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

// List возвращает список задач с пагинацией и фильтрацией
func (r *TaskRepository) List(ctx context.Context, filter domain.TaskFilter, pagination domain.Pagination) (*domain.TaskListResult, error) {
	// Базовый запрос
	baseQuery := `FROM tasks WHERE 1=1`
	args := []any{}
	argIndex := 1

	// Добавляем фильтр по статусу
	if filter.Status != nil {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}

	// Запрос на подсчёт общего количества
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Запрос на получение данных
	selectQuery := fmt.Sprintf(`
		SELECT id, status, file_key, file_name, content_type, schema, result, error, created_at, updated_at, completed_at
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIndex, argIndex+1)

	args = append(args, pagination.Limit(), pagination.Offset())

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*domain.Task, 0)
	for rows.Next() {
		task := &domain.Task{}
		var errorMsg *string  // Указатель для NULL

		err := rows.Scan(
			&task.ID,
			&task.Status,
			&task.FileKey,
			&task.FileName,
			&task.ContentType,
			&task.Schema,
			&task.Result,
			&errorMsg,  // Сканируем в указатель
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		// Обрабатываем NULL
		if errorMsg != nil {
			task.Error = *errorMsg
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return &domain.TaskListResult{
		Tasks:      tasks,
		Total:      total,
		Pagination: pagination,
	}, nil
}