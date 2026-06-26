package models

import (
	"encoding/json"
	"errors"
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
	Monitored        bool                `json:"monitored"`
    Downloading      bool                `json:"downloading"`
	TotalChapters    int                 `json:"total_chapters"`
	Thumbnail        string              `json:"thumbnail"`
	HistoricalCovers []VolumeCover       `json:"historical_covers"`
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

type VolumeCover struct {
	Volume float64 `json:"volume"`
	URL    string  `json:"url"`
}

func (vc *VolumeCover) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("empty json data")
	}

	if data[0] == '"' {
		var legacyURL string
		if err := json.Unmarshal(data, &legacyURL); err != nil {
			return err
		}
		vc.URL = legacyURL
		vc.Volume = -1.0
		return nil
	}

	type Alias VolumeCover
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(vc),
	}
	
	return json.Unmarshal(data, &aux)
}
