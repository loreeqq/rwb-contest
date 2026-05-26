package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rwb-contest/internal/dto"
)

/*
HealthHandler отвечает за healthcheck эндпоинт.

Используется для проверки,
что сервис запущен и доступен.
*/

type HealthHandler struct{}

// NewHealthHandler создаёт экземпляр health handler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health возвращает состояние сервиса.
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{
		Status: "ok",
	})
}
