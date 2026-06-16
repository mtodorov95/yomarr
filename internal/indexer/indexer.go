package indexer

type SearchResult struct {
	Title    string `json:"title"`
	Link     string `json:"link"`
	Guid     string `json:"guid"`
	Seeders  int    `json:"seeders"`
	Leechers int    `json:"leechers"`
	InfoHash string `json:"infoHash"`
	Size     string `json:"size"`
	Language string `json:"language"`
	SeedTime int    `json:"seed_time"`
}

type Indexer interface {
	Search(query string) ([]SearchResult, error)
	FetchLatestRSS() ([]SearchResult, error)
}
