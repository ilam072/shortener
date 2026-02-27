package dto

type Click struct {
	Alias     string `json:"alias"`
	UserAgent string `json:"user_agent"`
	Client    string `json:"client"`
	Device    string `json:"device"`
	IP        string `json:"ip"`
}

type GetClicks struct {
	Alias       string              `json:"alias"`
	ByDay       []ClicksByDay       `json:"by_day"`
	ByMonth     []ClicksByMonth     `json:"by_month"`
	ByUserAgent []ClicksByUserAgent `json:"by_user_agent"`
}

type ClicksByDay struct {
	Date   string `json:"date"`
	Clicks int    `json:"clicks"`
}

type ClicksByMonth struct {
	Month  string `json:"month"`
	Clicks int    `json:"clicks"`
}

type ClicksByUserAgent struct {
	UserAgent string `json:"user_agent"`
	Clicks    int    `json:"clicks"`
}
