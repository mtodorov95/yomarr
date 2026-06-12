package models

import (
	"time"
)

type Series struct {
	ID               int64               `json:"id"`
	AnilistID        *string             `json:"anilist_id"`
	MangadexID       *string             `json:"mangadex_id"`
	Title            string              `json:"title"`
	AltTitles        map[string][]string `json:"alt_titles"`
	Path             string              `json:"path"`
	Status           SeriesStatus        `json:"status"`
	TotalChapters    int                 `json:"total_chapters"`
	Thumbnail        string              `json:"thumbnail"`
	HistoricalCovers []string            `json:"historical_covers"`
	Author           *string             `json:"author"`
	Genres           []string            `json:"genres"`
	Description      *string             `json:"description"`
	Artist           *string             `json:"artist"`
	Year             *int                `json:"year"`
	LastChapter      *string             `json:"last_chapter"`
	LastVolume       *string             `json:"last_volume"`
	DownloadedCount  int                 `json:"downloaded_count"`
}

type Chapters struct {
	ID          int64         `json:"id"`
	SeriesID    int64         `json:"series_id"`
	Number      float64       `json:"number"`
	Volume      *int          `json:"volume"`
	FilePath    *string       `json:"file_path"`
	Status      ChapterStatus `json:"status"`
	ReleaseDate *time.Time    `json:"release_date"`
	Language    string        `json:"language"`
}
