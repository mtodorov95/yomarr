package metadata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mtodorov95/yomarr/internal/models"
)

type MangaDexProvider struct {
	Client *http.Client
}

type mdLocalizedString map[string]string

type mdMangaAttributes struct {
	Title     mdLocalizedString   `json:"title"`
	AltTitles []mdLocalizedString `json:"altTitles"`
	Status    string              `json:"status"`
	Links     map[string]string   `json:"links"`
}

type mdMangaData struct {
	ID         string            `json:"id"`
	Attributes mdMangaAttributes `json:"attributes"`
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
	Data []mdChapterData `json:"data"`
}

type ExtChapter struct {
	Number   float64
	Volume   *string
	Title    string
	Language string
}

func getMDTitle(titleMap mdLocalizedString) string {
	if t, ok := titleMap["ja"]; ok && t != "" {
		return t
	}
	if t, ok := titleMap["en"]; ok && t != "" {
		return t
	}
	for _, t := range titleMap {
		if t != "" {
			return t
		}
	}
	return "Unknown Title"
}

func mapMDStatus(status string) string {
	switch status {
	case "ongoing":
		return "Monitored"
	case "completed":
		return "Ended"
	default:
		return "Monitored"
	}
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
			if enTitle, ok := alt["en"]; ok && enTitle != "" {
				fallbacks = append(fallbacks, enTitle)
			}
		}

		results = append(results, models.Series{
			MangadexID: &mdID,
			Title:      getMDTitle(item.Attributes.Title),
			AltTitles:  fallbacks,
			Status:     mapMDStatus(item.Attributes.Status),
			AnilistID:  alIDPtr,
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
		if enTitle, ok := alt["en"]; ok && enTitle != "" {
			fallbacks = append(fallbacks, enTitle)
		}
	}

	return &models.Series{
		MangadexID: &mdID,
		AnilistID:  alIDPtr,
		AltTitles:  fallbacks,
		Title:      getMDTitle(res.Data.Attributes.Title),
		Status:     mapMDStatus(res.Data.Attributes.Status),
	}, nil
}

func (p *MangaDexProvider) GetChapterFeed(mangadexID string) ([]ExtChapter, error) {
	apiURL := fmt.Sprintf("https://api.mangadex.org/manga/%s/feed", url.PathEscape(mangadexID))
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("translatedLanguage[]", "en")
	q.Add("limit", "500")
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

	var list []ExtChapter
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

	return list, nil
}
