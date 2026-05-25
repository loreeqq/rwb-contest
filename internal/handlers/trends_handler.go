package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"rwb-contest/internal/dto"
	"rwb-contest/internal/service"
)

type TrendsHandler struct {
	service *service.TrendsService
}

func NewTrendsHandler(service *service.TrendsService) *TrendsHandler {
	return &TrendsHandler{
		service: service,
	}
}

// GetTop обрабатывает GET /top?n=10
func (h *TrendsHandler) GetTop(c *gin.Context) {

	// Значение по умолчанию.
	n := 10

	// Получаем query parameter.
	nParam := c.Query("n")

	// Если параметр передан —
	// преобразуем string -> int.
	if nParam != "" {
		parsed, err := strconv.Atoi(nParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "invalid request",
			})

			return
		}

		n = parsed
	}

	response := h.service.GetTop(n)

	c.JSON(http.StatusOK, response)
}
