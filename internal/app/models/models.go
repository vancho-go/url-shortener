package models

type APIShortenRequest struct {
	URL string `json:"url"`
}

type APIShortenResponse struct {
	Result string `json:"result"`
}

type APIBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type APIBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortenURL    string `json:"short_url"`
}

type APIUserURLResponse struct {
	ShortenURL  string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
