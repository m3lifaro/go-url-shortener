package model

type ShortenRecord struct {
	ID         string `json:"uuid"`
	ShortenURL string `json:"short_url"`
	URL        string `json:"original_url"`
	UserID     string `json:"user_id"`
}
