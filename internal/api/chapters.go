package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/models"
)

type ChapterHandler struct {
	Store db.ChapterStore
}

type LanguageVariant struct {
	ID       int64   `json:"id"`
	Status   string  `json:"status"`
	Language string  `json:"language"`
	FilePath *string `json:"file_path,omitempty"`
}

type GroupedChapter struct {
	Number   float64           `json:"number"`
	Volume   *int              `json:"volume"`
	Variants []LanguageVariant `json:"variants"`
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

	groupsMap := make(map[float64]*GroupedChapter)
	var orderedKeys []float64

	for i := range list {
		ch := list[i]
		lang := strings.ToLower(ch.Language)
		if lang == "" {
			lang = "en"
		}

		variant := LanguageVariant{
			ID:       ch.ID,
			Status:   string(ch.Status),
			Language: lang,
			FilePath: ch.FilePath,
		}

		if group, exists := groupsMap[ch.Number]; exists {
			group.Variants = append(group.Variants, variant)
		} else {
			groupsMap[ch.Number] = &GroupedChapter{
				Number:   ch.Number,
				Volume:   ch.Volume,
				Variants: []LanguageVariant{variant},
			}
			orderedKeys = append(orderedKeys, ch.Number)
		}
	}

	response := make([]GroupedChapter, 0, len(orderedKeys))
	for _, key := range orderedKeys {
		response = append(response, *groupsMap[key])
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
