package dto

type Click struct {
	Alias     string `json:"alias"`
	UserAgent string `json:"user_agent"`
	Client    string `json:"client"`
	Device    string `json:"device"`
	IP        string `json:"ip"` // валидировать
}
