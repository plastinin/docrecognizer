package domain

// TaskStatus представляет статус задачи распознавания
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"     // Задача создана, ожидает обработки
	TaskStatusProcessing TaskStatus = "processing"  // Задача в обработке
	TaskStatusCompleted  TaskStatus = "completed"   // Задача успешно завершена
	TaskStatusFailed     TaskStatus = "failed"      // Задача завершилась с ошибкой
)

// IsValid проверяет валидность статуса
func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusPending, TaskStatusProcessing, TaskStatusCompleted, TaskStatusFailed:
		return true
	}
	return false
}

// IsFinal проверяет, является ли статус финальным
func (s TaskStatus) IsFinal() bool {
	return s == TaskStatusCompleted || s == TaskStatusFailed
}

func (s TaskStatus) String() string {
	return string(s)
}