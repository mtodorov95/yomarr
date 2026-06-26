package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/models"
)

type AnilistProvider struct {
	Client *http.Client
}

const AnilistURL string = "https://graphql.anilist.co"

const detailsQuery = `
	query ($id: Int) {
		Media (id: $id, type: MANGA) {
			id
			title {romaji english}
			status
			chapters
		}
	}
`

const searchQuery = `
query ($search: String) {
  Page (perPage: 10) {
    media (search: $search, type: MANGA) {
      id
      title { romaji english }
      status
    }
  }
}`

func mapAnilistStatus(apiStatus string) models.SeriesStatus {
	switch apiStatus {
	case "FINISHED":
		return models.SeriesCompleted
	case "RELEASING", "HIATUS":
		return models.SeriesOngoing
	case "NOT_YET_RELEASED", "CANCELLED":
		return models.SeriesCancelled
	default:
		return models.SeriesUnknown
	}
}

func (p *AnilistProvider) Search(queryStr string) ([]models.Series, error) {
	payload := map[string]any{
		"query": searchQuery,
		"variables": map[string]string{
			"search": queryStr,
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", AnilistURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	res, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data struct {
		Data struct {
			Page struct {
				Media []struct {
					ID    int64 `json:"id"`
					Title struct {
						Romaji  string `json:"romaji"`
						English string `json:"english"`
					} `json:"title"`
					Status string `json:"status"`
				} `json:"media"`
			} `json:"Page"`
		} `json:"data"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []models.Series
	for _, m := range data.Data.Page.Media {
		title := m.Title.English
		if title == "" {
			title = m.Title.Romaji
		}
		results = append(results, models.Series{
			AnilistID: db.ToPtr(fmt.Sprintf("%d", m.ID)),
			Title:     title,
			Status:    mapAnilistStatus(m.Status),
		})
	}

	return results, nil
}

func (p *AnilistProvider) GetDetails(id string) (*models.Series, error) {
	payload := map[string]any{
		"query": detailsQuery,
		"variables": map[string]string{
			"id": id,
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", AnilistURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	res, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data struct {
		Data struct {
			Media struct {
				ID    int64 `json:"id"`
				Title struct {
					Romaji  string `json:"romaji"`
					English string `json:"english"`
				} `json:"title"`
				Status string `json:"status"`
				Chapters *int `json:"chapters"`
			} `json:"Media"`
		} `json:"data"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	title := data.Data.Media.Title.English
	if title == "" {
		title = data.Data.Media.Title.Romaji
	}

	var rawCount int
	if data.Data.Media.Chapters != nil {
		rawCount = *data.Data.Media.Chapters
	}

	return &models.Series{
		AnilistID: db.ToPtr(fmt.Sprintf("%d", data.Data.Media.ID)),
		Title:     title,
		Status:    mapAnilistStatus(data.Data.Media.Status),
		TotalChapters: rawCount,
	}, nil
}
