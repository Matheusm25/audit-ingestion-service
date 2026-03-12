package handler

import (
	"encoding/json"
	"net/http"

	"github.com/matheusm25/audit-ingestion-service/internal/health"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type ServicesHealthResponse struct {
	RabbitMQ   string `json:"rabbitMQStatus"`
	ClickHouse string `json:"clickHouseStatus"`
}

type HealthHandler struct {
	checker *health.Checker
}

func NewHealthHandler(checker *health.Checker) *HealthHandler {
	return &HealthHandler{
		checker: checker,
	}
}

func (h *HealthHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "healthy",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *HealthHandler) HealthCheckServicesHandler(w http.ResponseWriter, r *http.Request) {
	rabbitmqStatus := h.checker.CheckRabbitMQ()
	clickhouseStatus := h.checker.CheckClickHouse()

	response := ServicesHealthResponse{
		RabbitMQ:   rabbitmqStatus,
		ClickHouse: clickhouseStatus,
	}

	statusCode := http.StatusOK
	if rabbitmqStatus == health.StatusUnhealthy || clickhouseStatus == health.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
