package dto

// ErrorResponse ответ с ошибкой
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// NewErrorResponse создаёт ответ с ошибкой
func NewErrorResponse(err string, message string) *ErrorResponse {
	return &ErrorResponse{
		Error:   err,
		Message: message,
	}
}