package api

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/download"
	"github.com/mtodorov95/yomarr/internal/metadata"
	"github.com/mtodorov95/yomarr/internal/models"
	"github.com/mtodorov95/yomarr/internal/sync"
	"github.com/mtodorov95/yomarr/internal/utils"
)

type SeriesHandler struct {
	Store      db.SeriesStore
	Metadata   metadata.Provider
	SyncEngine *sync.MangaDexSyncEngine
	Scanner    *sync.LibraryScanner
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
		idStr := r.URL.Query().Get("id")
		action := r.URL.Query().Get("action")

		if idStr != "" && action == "refresh" {
			h.refresh(w, r, idStr)
			return
		}

		h.create(w, r)
	case http.MethodPut:
		h.update(w, r)
	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		coverFile := r.URL.Query().Get("cover")

		if idStr != "" && coverFile != "" {
			h.deleteCover(w, r, idStr, coverFile)
			return
		}

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
		AnilistID     string              `json:"anilist_id"`
		MangadexId    string              `json:"mangadex_id"`
		Title         string              `json:"title"`
		AltTitles     map[string][]string `json:"alt_titles"`
		Status        string              `json:"status"`
		Path          string              `json:"path"`
		TotalChapters int                 `json:"total_chapters"`
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

	libraryRoot := filepath.Dir(req.Path)

	finalPath, err := utils.EnsureSeriesDirectory(libraryRoot, &s)
	if err != nil {
		log.Printf("[API Error] Filesystem allocation aborted: %v", err)
		http.Error(w, "Failed creating local storage directory", http.StatusInternalServerError)
		return
	}
	s.Path = finalPath

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

func (h *SeriesHandler) refresh(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid series ID format", http.StatusBadRequest)
		return
	}

	updatedSeries, err := h.Scanner.RefreshSeriesMetadata(id)
	if err != nil {
		log.Printf("[API Error] Metadata refresh task failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedSeries)
}

func (h *SeriesHandler) update(w http.ResponseWriter, r *http.Request) {
	var s models.Series
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if s.ID == 0 {
		http.Error(w, "Missing structural identifier 'id' target field required for updates", http.StatusBadRequest)
		return
	}

	if err := h.Store.Update(&s); err != nil {
		log.Printf("[API Error] Failed committing series entity update: %v", err)
		http.Error(w, "Internal server data persistence failure", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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

func (h *SeriesHandler) deleteCover(w http.ResponseWriter, r *http.Request, idStr string, coverFile string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid series ID format", http.StatusBadRequest)
		return
	}

	s, err := h.Store.GetById(id)
	if err != nil || s == nil {
		http.Error(w, "Series target record not found", http.StatusNotFound)
		return
	}

	_ = utils.DeleteSeriesFile(s.Path, coverFile)

	var updatedCovers []models.VolumeCover
	for _, c := range s.HistoricalCovers {
		if c.URL != coverFile {
			updatedCovers = append(updatedCovers, c)
		}
	}
	s.HistoricalCovers = updatedCovers

	if err := h.Store.Update(s); err != nil {
		log.Printf("[API Error] Failed saving post-delete cover manifest: %v", err)
		http.Error(w, "Internal server data sync failure", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}
