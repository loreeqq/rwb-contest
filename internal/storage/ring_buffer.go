package storage

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"rwb-contest/internal/config"
	"rwb-contest/internal/dto"
)

var ErrWordAlreadyExist = errors.New("stop-word already exists")

/*
Пакет storage реализует хранилище статистики поисковых запросов внутри памяти.

В основе лежит кольцевой буфер из 300 бакетов:
каждый бакет хранит агрегированные счётчики запросов за одну секунду.
Это позволяет эффективно считать Top-N запросов за последние 5 минут
без хранения всех входящих событий.

Хранилище используется одновременно несколькими goroutine:
- RabbitMQ workers добавляют новые запросы;
- HTTP handlers читают текущий топ и стоп-слова.

Для потокобезопасности используется sync.RWMutex:
- Lock при записи;
- RLock при чтении.

Стоп-слова хранятся в map[string]struct{} как множество строк.
*/

type Bucket struct {
	Timestamp int64
	Counts    map[string]int
}

type RingBufferStorage struct {
	mu        sync.RWMutex
	buckets   [config.WindowSeconds]Bucket
	stopWords map[string]struct{}
}

// NewRingBufferStorage создаёт кольцевой буфер (0-299)
// и инициализирует бакеты и мапу стоп-слов.
func NewRingBufferStorage() *RingBufferStorage {
	s := &RingBufferStorage{
		stopWords: make(map[string]struct{}),
	}

	for i := range s.buckets {
		s.buckets[i].Counts = make(map[string]int)
	}

	return s
}

// Add добавляет поисковый запрос
// в бакет текущей секунды.
func (s *RingBufferStorage) Add(query string) {
	query = normalize(query)
	if query == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// СТОП СЛОВО СТОП СЛОВО!!!
	if s.isStopWordLocked(query) {
		return
	}

	now := time.Now().Unix()

	// Определяем индекс бакета
	// для текущей секунды.
	idx := now % config.WindowSeconds

	bucket := &s.buckets[idx]

	// Если бакет хранит старые данные —
	// очищаем его и используем заново.
	if bucket.Timestamp != now {
		bucket.Timestamp = now
		bucket.Counts = make(map[string]int)
	}

	if bucket.Counts[query] >= config.MaxQueryPerSecond {
		return
	}

	bucket.Counts[query]++
}

// Top возвращает Top-N запросов
// за последние 5 минут.
func (s *RingBufferStorage) Top(n int) []dto.TopItem {
	if n <= 0 {
		n = 10
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now().Unix()

	// aggregated хранит итоговые счётчики
	// по всем актуальным бакетам.
	aggregated := make(map[string]int)

	for i := range s.buckets {
		bucket := s.buckets[i]

		if bucket.Timestamp == 0 {
			continue
		}

		// Игнорируем бакеты старше 5 минут.
		if now-bucket.Timestamp >= config.WindowSeconds {
			continue
		}

		// Собираем общую статистику.
		for query, count := range bucket.Counts {
			aggregated[query] += count
		}
	}

	// Преобразуем мапу в слайс для сортировки.
	items := make([]dto.TopItem, 0, len(aggregated))

	for query, count := range aggregated {
		items = append(items, dto.TopItem{
			Query: query,
			Count: count,
		})
	}

	// Сортировка по убыванию популярности.
	sort.Slice(items, func(i, j int) bool {
		return items[i].Count > items[j].Count
	})

	if n > len(items) {
		n = len(items)
	}

	return items[:n]
}

// AddStopWord добавляет слово
// в множество стоп-слов.
func (s *RingBufferStorage) AddStopWord(word string) error {
	word = normalize(word)
	if word == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.stopWords[word]; exists {
		return ErrWordAlreadyExist
	}

	s.stopWords[word] = struct{}{}

	return nil
}

// RemoveStopWord удаляет стоп-слово.
func (s *RingBufferStorage) RemoveStopWord(word string) {
	word = normalize(word)

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.stopWords, word)
}

// ListStopWords возвращает список стоп-слов.
func (s *RingBufferStorage) ListStopWords() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	words := make([]string, 0, len(s.stopWords))

	for word := range s.stopWords {
		words = append(words, word)
	}

	sort.Strings(words)

	return words
}

// isStopWordLocked проверяет,
// содержит ли запрос стоп-слово.
//
// Предполагается, что mutex уже захвачен.
func (s *RingBufferStorage) isStopWordLocked(query string) bool {
	for word := range s.stopWords {
		if strings.Contains(query, word) {
			return true
		}
	}

	return false
}

// normalize приводит строку
// к единому виду:
// удаляет пробелы и переводит в lowercase.
func normalize(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)
	return value
}
