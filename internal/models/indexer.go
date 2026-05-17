package models

type Indexer struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	APIKey   string `json:"api_key"`
	Priority int    `json:"priority"`
}
