package dto

type Link struct {
	URL   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}
