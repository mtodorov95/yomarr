package metadata

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mtodorov95/yomarr/internal/models"
)

type MangaDexProvider struct {
	Client *http.Client
}

type mdLocalizedString map[string]string

type mdRelationship struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

type mdMangaAttributes struct {
	Title     mdLocalizedString   `json:"title"`
	AltTitles []mdLocalizedString `json:"altTitles"`
	Status    string              `json:"status"`
	Links     map[string]string   `json:"links"`
}

type mdMangaData struct {
	ID            string            `json:"id"`
	Attributes    mdMangaAttributes `json:"attributes"`
	Relationships []mdRelationship  `json:"relationships"`
}

type mdSearchResponse struct {
	Data []mdMangaData `json:"data"`
}

type mdDetailsResponse struct {
	Data mdMangaData `json:"data"`
}

type mdChapterAttributes struct {
	Volume             *string `json:"volume"`
	Chapter            string  `json:"chapter"`
	Title              string  `json:"title"`
	TranslatedLanguage string  `json:"translatedLanguage"`
	PublishAt          string  `json:"publishAt"`
}

type mdChapterData struct {
	ID         string              `json:"id"`
	Attributes mdChapterAttributes `json:"attributes"`
}

type mdFeedResponse struct {
	Data   []mdChapterData `json:"data"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
	Total  int             `json:"total"`
}

type ExtChapter struct {
	Number   float64
	Volume   *string
	Title    string
	Language string
}

func getMDTitle(titleMap mdLocalizedString, altTitles []mdLocalizedString) string {
	if t, ok := titleMap["ja"]; ok && t != "" {
		return t
	}
	if t, ok := titleMap["ja-ro"]; ok && t != "" {
		return t
	}

	if t, ok := titleMap["en"]; ok && t != "" {
		return t
	}
	for _, alt := range altTitles {
		if t, ok := alt["en"]; ok && t != "" {
			return t
		}
	}

	for _, t := range titleMap {
		if t != "" {
			return t
		}
	}
	for _, alt := range altTitles {
		for _, t := range alt {
			if t != "" {
				return t
			}
		}
	}
	return "Unknown Title"
}

func mapMDStatus(status string) models.SeriesStatus {
	switch status {
	case "ongoing", "hiatus":
		return models.SeriesOngoing
	case "completed":
		return models.SeriesCompleted
	case "cancelled":
		return models.SeriesUnmonitored
	default:
		return models.SeriesOngoing
	}
}

func (p *MangaDexProvider) fetchAllCovers(mangaID string) (string, []string, error) {
	apiURL := fmt.Sprintf("https://api.mangadex.org/cover?manga[]=%s&limit=100", url.QueryEscape(mangaID))
	resp, err := p.Client.Get(apiURL)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("failed fetching covers, status: %d", resp.StatusCode)
	}

	var coverRes struct {
		Data []struct {
			Attributes struct {
				FileName string `json:"fileName"`
				Volume   string `json:"volume"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&coverRes); err != nil {
		return "", nil, err
	}

	var primaryCover string
	var historicalCovers []string

	for _, c := range coverRes.Data {
		coverURL := fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", mangaID, c.Attributes.FileName)

		if c.Attributes.Volume == "1" || c.Attributes.Volume == "" {
			primaryCover = coverURL
		}
		historicalCovers = append(historicalCovers, coverURL)
	}

	if primaryCover == "" && len(historicalCovers) > 0 {
		primaryCover = historicalCovers[0]
	}

	return primaryCover, historicalCovers, nil
}

func (p *MangaDexProvider) Search(query string) ([]models.Series, error) {
	apiURL := "https://api.mangadex.org/manga"
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("title", query)
	q.Add("limit", "10")
	q.Add("includes[]", "cover_art")
	req.URL.RawQuery = q.Encode()

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mangadex search failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mangadex search returned code: %d", resp.StatusCode)
	}

	var res mdSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	var results []models.Series
	for _, item := range res.Data {
		mdID := item.ID
		var alIDPtr *string
		if alID, ok := item.Attributes.Links["al"]; ok && alID != "" {
			alIDPtr = &alID
		}

		var fallbacks []string
		for _, alt := range item.Attributes.AltTitles {
			for _, altTitle := range alt {
				if altTitle != "" {
					fallbacks = append(fallbacks, altTitle)
				}
			}
		}

		var primaryCover string
		for _, rel := range item.Relationships {
			if rel.Type == "cover_art" {
				if filename, ok := rel.Attributes["fileName"].(string); ok {
					primaryCover = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", mdID, filename)
				}
			}
		}

		results = append(results, models.Series{
			MangadexID: &mdID,
			Title:      getMDTitle(item.Attributes.Title, item.Attributes.AltTitles),
			AltTitles:  fallbacks,
			Status:     mapMDStatus(item.Attributes.Status),
			AnilistID:  alIDPtr,
			Thumbnail:  primaryCover,
			HistoricalCovers: make([]string, 0),
		})
	}

	return results, nil
}

func (p *MangaDexProvider) GetDetails(id string) (*models.Series, error) {
	apiURL := fmt.Sprintf("https://api.mangadex.org/manga/%s", url.PathEscape(id))
	resp, err := p.Client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("mangadex get details failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mangadex details returned code: %d", resp.StatusCode)
	}

	var res mdDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	mdID := res.Data.ID
	var alIDPtr *string
	if alID, ok := res.Data.Attributes.Links["al"]; ok && alID != "" {
		alIDPtr = &alID
	}

	var fallbacks []string
	for _, alt := range res.Data.Attributes.AltTitles {
		for _, altTitle := range alt {
			if altTitle != "" {
				fallbacks = append(fallbacks, altTitle)
			}
		}
	}

	thumbnail, historical, err := p.fetchAllCovers(mdID)
	if err != nil {
		log.Printf("[Metadata Warning] Failed retrieving covers for %s: %v", mdID, err)
	}

	return &models.Series{
		MangadexID:       &mdID,
		AnilistID:        alIDPtr,
		AltTitles:        fallbacks,
		Title:            getMDTitle(res.Data.Attributes.Title, res.Data.Attributes.AltTitles),
		Status:           mapMDStatus(res.Data.Attributes.Status),
		Thumbnail:        thumbnail,
		HistoricalCovers: historical,
	}, nil
}

func (p *MangaDexProvider) GetChapterFeed(mangadexID string) ([]ExtChapter, error) {
	var list []ExtChapter
	offset := 0
	limit := 500

	for {
		apiURL := fmt.Sprintf("https://api.mangadex.org/manga/%s/feed", url.PathEscape(mangadexID))
		req, err := http.NewRequest(http.MethodGet, apiURL, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Add("limit", strconv.Itoa(limit))
		q.Add("offset", strconv.Itoa(offset))
		q.Add("order[chapter]", "asc")
		req.URL.RawQuery = q.Encode()

		resp, err := p.Client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("mangadex returned status: %d", resp.StatusCode)
		}

		var feed mdFeedResponse
		if err := json.NewDecoder(resp.Body).Decode(&feed); err != nil {
			return nil, err
		}

		for _, ch := range feed.Data {
			if ch.Attributes.Chapter == "" {
				continue
			}

			var num float64
			if _, err := fmt.Sscanf(ch.Attributes.Chapter, "%f", &num); err != nil {
				continue
			}

			list = append(list, ExtChapter{
				Number:   num,
				Volume:   ch.Attributes.Volume,
				Title:    ch.Attributes.Title,
				Language: ch.Attributes.TranslatedLanguage,
			})
		}

		if offset+len(feed.Data) >= feed.Total || len(feed.Data) == 0 {
			break
		}

		offset += limit
	}

	log.Printf("[Metadata Feed] Successfully aggregated complete chapter skeleton tracking array. Total size: %d", len(list))
	return list, nil
}
