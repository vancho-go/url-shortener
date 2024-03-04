// Модуль models содержит в себе типовые структуры Request и Response для различных handler'ов.
package models

// APIShortenRequest содержит поля, необходимые для запроса на эндпоинт, который генерирует один сокращенный URL.
type APIShortenRequest struct {
	URL string `json:"url"`
}

// APIShortenResponse содержит сокращенный URL.
type APIShortenResponse struct {
	Result string `json:"result"`
}

// APIShortenRequest содержит поля, необходимые для запроса на эндпоинт,
// который генерирует сразу несколько сокращенных URL (batch).
type APIBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
	ShortenURL    string `json:"shorten_url"`
}

// APIBatchResponse содержит batch из сокращенный URL.
type APIBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortenURL    string `json:"short_url"`
}

// APIUserURLResponse содержит соотношения "оригинальный URL - сокращенный URL" для конкретного пользователя.
type APIUserURLResponse struct {
	ShortenURL  string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// DeleteURLRequest содержит поля, необходимые для запроса на эндпоинт,
// который удаляет сокращенный URL конкретного пользователя.
type DeleteURLRequest struct {
	UserID     string
	ShortenURL string
}

// APIStatsResponse содержит статистику по количеству сокращенных URL и пользователей.
type APIStatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}
