package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mtodorov95/yomarr/internal/indexer"
	"github.com/mtodorov95/yomarr/internal/sync"
)

type TorrentVariant struct {
	Title     string `json:"title"`
	Link      string `json:"link"`
	Seeders   int    `json:"seeders"`
	Leechers  int    `json:"leechers"`
	Size      string `json:"size"`
	InfoHash  string `json:"info_hash"`
}

type ChapterGroup struct {
	ChapterNumber float64          `json:"chapter_number"`
	Volume        *int             `json:"volume,omitempty"`
	English       []TorrentVariant `json:"english"`
	Raws          []TorrentVariant `json:"raws"`
}

type SearchHandler struct {
	NyaaEngine *sync.NyaaSyncEngine
}

func NewSearchHandler(e *sync.NyaaSyncEngine) *SearchHandler {
	return &SearchHandler{NyaaEngine: e}
}

func (h *SearchHandler) SearchMissing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("series_id")
	if idStr == "" {
		http.Error(w, "missing series_id", http.StatusBadRequest)
		return
	}

	seriesID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid series_id", http.StatusBadRequest)
		return
	}

	go func() {
		_ = h.NyaaEngine.FindMissingChapters(seriesID)
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"searching"}`))
}

func (h *SearchHandler) SearchChaptersInteractive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("series_id")
	seriesID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid or missing series_id", http.StatusBadRequest)
		return
	}

	series, err := h.NyaaEngine.SeriesStore.GetById(seriesID)
	if err != nil {
		http.Error(w, "Series not found", http.StatusNotFound)
		return
	}

	searchQueries := append([]string{series.Title}, series.AltTitles...)
	var rawResults []indexer.NyaaResult

	for _, queryTitle := range searchQueries {
		rawResults, err = h.NyaaEngine.Indexer.Search(queryTitle)
		if err == nil && len(rawResults) > 0 {
			break
		}
	}

	groupMap := make(map[string]*ChapterGroup)

	for _, res := range rawResults {
		parsed, ok := indexer.ParseTorrentTitle(res.Title)
		if !ok {
			continue
		}

		chKey := fmt.Sprintf("%.2f", parsed.StartNum)

		group, exists := groupMap[chKey]
		if !exists {
			group = &ChapterGroup{
				ChapterNumber: parsed.StartNum,
				English:       []TorrentVariant{},
				Raws:          []TorrentVariant{},
			}
			groupMap[chKey] = group
		}

		variant := TorrentVariant{
			Title:    res.Title,
			Link:     res.Link,
			Seeders:  res.Seeders,
			Leechers: res.Leechers,
			Size:     res.Size,
			InfoHash: res.InfoHash,
		}

		if res.Language == "raw" {
			group.Raws = append(group.Raws, variant)
		} else {
			group.English = append(group.English, variant)
		}
	}

	var finalResponse []*ChapterGroup
	for _, g := range groupMap {
		if len(g.English) > 0 || len(g.Raws) > 0 {
			finalResponse = append(finalResponse, g)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(finalResponse)
}
