package domain

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Pagination параметры пагинации
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// NewPagination создаёт параметры пагинации с валидацией
func NewPagination(page, pageSize int) Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// Offset возвращает смещение для SQL запроса
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit возвращает лимит для SQL запроса
func (p Pagination) Limit() int {
	return p.PageSize
}

// TaskFilter фильтры для списка задач
type TaskFilter struct {
	Status *TaskStatus `json:"status,omitempty"`
}

// TaskListResult результат запроса списка задач
type TaskListResult struct {
	Tasks      []*Task    `json:"tasks"`
	Total      int        `json:"total"`
	Pagination Pagination `json:"pagination"`
}