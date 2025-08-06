package model

type ShortenRequest struct {
	ShortingURL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}
