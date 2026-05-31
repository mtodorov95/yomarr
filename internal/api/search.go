package api

import (
	"net/http"
	"strconv"

	"github.com/mtodorov95/yomarr/internal/sync"
)

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
