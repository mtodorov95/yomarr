package models

type Indexer struct {
	ID                   int64  `json:"id"`
	Name                 string `json:"name"`
	URL                  string `json:"url"`
	APIKey               string `json:"api_key"`
	Priority             int    `json:"priority"`
	EnableRSS            bool   `json:"enable_rss"`
	EnableSearch         bool   `json:"enable_search"`
	AdditionalParameters string `json:"additional_parameters"`
	MinimumSeeders       int    `json:"minimum_seeders"`
	SeedTime             int    `json:"seed_time"`
}

type DownloadClient struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	UseSSL   bool   `json:"use_ssl"`
	Username string `json:"username"`
	Password string `json:"password"`
	Category string `json:"category"`
}
