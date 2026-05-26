package metrics

/*
Пакет metrics содержит Prometheus-метрики приложения.

Метрики используются для:
- мониторинга количества обработанных событий;
- отслеживания HTTP-запросов;
- наблюдения за состоянием stop-list.

Все метрики экспортируются через endpoint /metrics.
*/

import "github.com/prometheus/client_golang/prometheus"

var (

	// Общее количество обработанных search events.
	SearchEventsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "search_events_total",
			Help: "Total number of processed search events",
		},
	)

	// Количество невалидных search events.
	InvalidSearchEventsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "search_events_invalid_total",
			Help: "Total number of invalid search events",
		},
	)

	// Количество запросов к endpoint /top.
	TopRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "top_requests_total",
			Help: "Total number of /top requests",
		},
	)

	// Текущее количество stop-words.
	StopWordsCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "stop_words_count",
			Help: "Current number of stop words",
		},
	)
)

// Init регистрирует все Prometheus-метрики приложения.
func Init() {
	prometheus.MustRegister(
		SearchEventsTotal,
		InvalidSearchEventsTotal,
		TopRequestsTotal,
		StopWordsCount,
	)
}
