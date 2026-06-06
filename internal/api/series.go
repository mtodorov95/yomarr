package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/download"
	"github.com/mtodorov95/yomarr/internal/metadata"
	"github.com/mtodorov95/yomarr/internal/models"
	"github.com/mtodorov95/yomarr/internal/sync"
)

type SeriesHandler struct {
	Store      db.SeriesStore
	Metadata   metadata.Provider
	SyncEngine *sync.MangaDexSyncEngine
}

func (h *SeriesHandler) HandleSeries(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Search
		if query := r.URL.Query().Get("search"); query != "" {
			h.searchMetadata(w, query)
			return
		}
		// By id
		idStr := r.URL.Query().Get("id")
		if idStr != "" {
			h.getById(w, idStr)
			return
		}
		// All
		h.getAll(w)
	case http.MethodPost:
		h.create(w, r)
	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		if idStr != "" {
			h.delete(w, idStr)
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *SeriesHandler) getAll(w http.ResponseWriter) {
	list, err := h.Store.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(list)
}

func (h *SeriesHandler) getById(w http.ResponseWriter, idStr string) {
	id, _ := strconv.ParseInt(idStr, 10, 64)
	s, err := h.Store.GetById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(s)
}

func (h *SeriesHandler) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AnilistID     string   `json:"anilist_id"`
		MangadexId    string   `json:"mangadex_id"`
		Title         string   `json:"title"`
		AltTitles     []string `json:"alt_titles"`
		Status        string   `json:"status"`
		Path          string   `json:"path"`
		TotalChapters int      `json:"total_chapters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var s models.Series
	if req.MangadexId != "" {
		extSeries, err := h.Metadata.GetDetails(req.MangadexId)
		if err != nil {
			http.Error(w, "Metadata fetch fail: "+err.Error(), http.StatusInternalServerError)
			return
		}
		s = *extSeries
		s.Path = req.Path

		if (s.AnilistID == nil || *s.AnilistID == "") && req.AnilistID != "" {
			s.AnilistID = db.ToPtr(req.AnilistID)
			s.TotalChapters = req.TotalChapters
		}
	} else {
		s = models.Series{
			Title:         req.Title,
			AltTitles:     req.AltTitles,
			Status:        models.SeriesStatus(req.Status),
			Path:          req.Path,
			MangadexID:    db.ToPtr(req.MangadexId),
			AnilistID:     db.ToPtr(req.AnilistID),
			TotalChapters: req.TotalChapters,
		}
	}

	if s.Path != "" && (s.Thumbnail != "" || len(s.HistoricalCovers) > 0) {
		log.Printf("[API] Downloading remote imagery locally into: %s/Covers", s.Path)
		localThumb, localHists, err := download.DownloadSeriesCovers(
			http.DefaultClient,
			s.Path,
			s.Thumbnail,
			s.HistoricalCovers,
		)
		if err != nil {
			log.Printf("[API Warning] Cover sync incomplete: %v", err)
		} else {
			s.Thumbnail = localThumb
			s.HistoricalCovers = localHists
		}
	}

	if err := h.Store.Insert(&s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.MangadexID != nil && *s.MangadexID != "" {
		if err := h.SyncEngine.SyncChapters(s.ID, req.MangadexId); err != nil {
			log.Printf("Non-fatal chapter ingestion sync error: %v", err)
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

func (h *SeriesHandler) searchMetadata(w http.ResponseWriter, query string) {
	results, err := h.Metadata.Search(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(results)
}

func (h *SeriesHandler) delete(w http.ResponseWriter, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid series ID", http.StatusBadRequest)
		return
	}

	if err := h.Store.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
