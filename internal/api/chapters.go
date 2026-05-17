package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/models"
)

type ChapterHandler struct {
	Store db.ChapterStore
}

func (h *ChapterHandler) HandleChapters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
		case http.MethodGet:
			seriesIdStr := r.URL.Query().Get("series_id")
			if seriesIdStr == "" {
				http.Error(w, "Missing series_id parameter", http.StatusBadRequest)
				return
			}
			h.getBySeriesId(w, seriesIdStr)
		case http.MethodPost:
			h.create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ChapterHandler) getBySeriesId(w http.ResponseWriter, idStr string) {
	seriesId, _ := strconv.ParseInt(idStr, 10, 64)
	list, err := h.Store.GetBySeriesId(seriesId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(list)
}

func (h *ChapterHandler) create(w http.ResponseWriter, r *http.Request) {
	var c models.Chapters
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Store.Insert(&c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}
