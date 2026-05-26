package service

import (
	"rwb-contest/internal/config"
	"rwb-contest/internal/dto"
	"rwb-contest/internal/metrics"
	"rwb-contest/internal/storage"
)

/*
Пакет service содержит бизнес-логику приложения.

TrendsService — промежуточный слой между:
- обработчиками HTTP/RabbitMQ;
- слоем хранения данных.

Сервис отвечает за:
- обработку входящих поисковых запросов;
- получение Top-N запросов;
- управление стоп-словами.

Сервис не знает деталей реализации хранилища
и работает только через интерфейс storage.Storage.
*/

type TrendsService struct {
	storage storage.Storage
}

// NewTrendsService создаёт экземпляр сервиса.
func NewTrendsService(storage storage.Storage) *TrendsService {
	return &TrendsService{
		storage: storage,
	}
}

// ProcessEvent обрабатывает входящее событие поиска.
func (s *TrendsService) ProcessEvent(event dto.SearchEvent) {
	s.storage.Add(event.Query)
}

// GetTop возвращает Top-N запросов
// за последние 5 минут.
func (s *TrendsService) GetTop(n int) dto.TopResponse {

	metrics.TopRequestsTotal.Inc()

	return dto.TopResponse{
		WindowSeconds: config.WindowSeconds,
		Items:         s.storage.Top(n),
	}
}

// AddStopWord добавляет слово в стоп-лист.
func (s *TrendsService) AddStopWord(word string) error {
	err := s.storage.AddStopWord(word)
	metrics.StopWordsCount.Set(
		float64(len(s.storage.ListStopWords())),
	)
	return err
}

// RemoveStopWord удаляет слово из стоп-листа.
func (s *TrendsService) RemoveStopWord(word string) {
	metrics.StopWordsCount.Set(
		float64(len(s.storage.ListStopWords())),
	)
	s.storage.RemoveStopWord(word)
}

// ListStopWords возвращает список стоп-слов.
func (s *TrendsService) ListStopWords() []string {
	return s.storage.ListStopWords()
}
