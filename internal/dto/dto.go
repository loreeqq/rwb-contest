package dto

/*
Пакет dto содержит структуры,
используемые для передачи данных между слоями приложения,
а также для работы с HTTP API и RabbitMQ сообщениями.
*/

// SearchEvent — входящее событие поискового запроса,
// получаемое из RabbitMQ.
type SearchEvent struct {
	Query string `json:"query"`
}

// TopItem — один элемент топа запросов.
type TopItem struct {
	Query string `json:"query"`
	Count int    `json:"count"`
}

// TopResponse — ответ HTTP API
// с Top-N запросами за последние 5 минут.
type TopResponse struct {
	WindowSeconds int       `json:"window_seconds"`
	Items         []TopItem `json:"items"`
}

// StopWordRequest — тело запроса
// для добавления стоп-слова.
type StopWordRequest struct {
	Word string `json:"word"`
}

// ErrorResponse — стандартный формат ошибки API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthResponse — ответ healthcheck endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}
