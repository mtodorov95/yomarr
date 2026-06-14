package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/models"
)

type DownloadClientHandler struct {
	Store db.DownloadClientStore
}

func (h *DownloadClientHandler) HandleDownloadClients(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		h.getAll(w)
	case http.MethodPost:
		h.create(w, r)
	case http.MethodPut:
		h.update(w, r)
	case http.MethodDelete:
		h.delete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DownloadClientHandler) getAll(w http.ResponseWriter) {
	list, err := h.Store.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *DownloadClientHandler) create(w http.ResponseWriter, r *http.Request) {
	var dc models.DownloadClient
	if err := json.NewDecoder(r.Body).Decode(&dc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Store.Insert(&dc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dc)
}

func (h *DownloadClientHandler) update(w http.ResponseWriter, r *http.Request) {
	var dc models.DownloadClient
	if err := json.NewDecoder(r.Body).Decode(&dc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Store.Update(&dc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dc)
}

func (h *DownloadClientHandler) delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid 'id' parameter format", http.StatusBadRequest)
		return
	}

	if err := h.Store.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
