package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/metadata"
	"github.com/mtodorov95/yomarr/internal/models"
)

type SeriesHandler struct {
	Store db.SeriesStore
	Metadata metadata.Provider
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
	var s models.Series
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.Store.Insert(&s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
