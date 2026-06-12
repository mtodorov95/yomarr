package api

import (
	"net/http"
	"strconv"

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
