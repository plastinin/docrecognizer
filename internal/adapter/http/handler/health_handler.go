package handler

import (
	"encoding/json"
	"net/http"
)

// HealthHandler обработчик health check запросов
type HealthHandler struct{}

// NewHealthHandler создаёт новый HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthResponse ответ health check
type HealthResponse struct {
	Status string `json:"status"`
}

// Check проверяет состояние сервиса
// GET /health
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
}