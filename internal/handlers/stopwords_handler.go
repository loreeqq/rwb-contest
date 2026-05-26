package handlers

import (
	"net/http"
	"rwb-contest/internal/dto"
	"rwb-contest/internal/service"

	"github.com/gin-gonic/gin"
)

/*
Пакет handlers содержит HTTP-обработчики приложения.

StopWordsHandler отвечает за работу со стоп-листом:
- получение списка стоп-слов;
- добавление нового стоп-слова;
- удаление стоп-слова.

Обработчик использует сервисный слой
и не работает напрямую с хранилищем данных.
*/

type StopWordsHandler struct {
	service *service.TrendsService
}

// NewStopWordsHandler создаёт экземпляр обработчика стоп-слов.
func NewStopWordsHandler(service *service.TrendsService) *StopWordsHandler {
	return &StopWordsHandler{
		service: service,
	}
}

// GetStopWords возвращает список стоп-слов.
func (h *StopWordsHandler) GetStopWords(c *gin.Context) {
	response := h.service.ListStopWords()

	c.JSON(http.StatusOK, response)
}

// AddStopWord добавляет новое слово в стоп-лист.
func (h *StopWordsHandler) AddStopWord(c *gin.Context) {

	// Входное тело запроса.
	var input dto.StopWordRequest

	// Проверяем корректность JSON.
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "invalid request",
		})

		return
	}

	h.service.AddStopWord(input.Word)

	c.JSON(http.StatusOK, gin.H{
		"message": "stop-word added",
		"word":    input.Word,
	})
}

// RemoveStopWord удаляет слово из стоп-листа.
func (h *StopWordsHandler) RemoveStopWord(c *gin.Context) {

	// Получаем слово из path parameter.
	word := c.Param("word")

	if word == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "invalid request",
		})

		return
	}

	h.service.RemoveStopWord(word)

	c.JSON(http.StatusOK, gin.H{
		"message": "stop-word removed",
		"word":    word,
	})
}
