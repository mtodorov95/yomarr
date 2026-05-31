package models

import (
	"time"
)

type Series struct {
	ID            int64    `json:"id"`
	AnilistID     *string  `json:"anilist_id"`
	MangadexID    *string  `json:"mangadex_id"`
	Title         string   `json:"title"`
	AltTitles     []string `json:"alt_titles"`
	Path          string   `json:"path"`
	Status        string   `json:"status"`
	TotalChapters int      `json:"total_chapters"`
}

type Chapters struct {
	ID          int64      `json:"id"`
	SeriesID    int64      `json:"series_id"`
	Number      float64    `json:"number"`
	Volume      *int       `json:"volume"`
	FilePath    *string    `json:"file_path"`
	Status      string     `json:"status"`
	ReleaseDate *time.Time `json:"release_date"`
	Language    string     `json:"language"`
}
