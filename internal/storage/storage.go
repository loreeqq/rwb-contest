package storage

import "rwb-contest/internal/dto"

/*
Storage описывает контракт хранилища статистики запросов.

Интерфейс используется сервисным слоем
и скрывает конкретную реализацию хранения данных.

В текущем проекте используется кольцевой буфер внутри памяти,
но при необходимости реализацию можно заменить,
например на редис или постгрес,
не меняя остальной код приложения.
*/
type Storage interface {

	// Add добавляет новый поисковый запрос в статистику.
	Add(query string)

	// Top возвращает Top-N популярных запросов
	// за последние 5 минут.
	Top(n int) []dto.TopItem

	// AddStopWord добавляет слово в стоп-лист.
	AddStopWord(word string) error

	// RemoveStopWord удаляет слово из стоп-листа.
	RemoveStopWord(word string)

	// ListStopWords возвращает список стоп-слов.
	ListStopWords() []string
}
