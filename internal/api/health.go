package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type HealthHandler struct {
	Version string
}

func NewHealthHandler(version string) *HealthHandler {
	cleanVersion := version
	if strings.Contains(version, "-") {
		cleanVersion = strings.Split(version, "-")[0]
	}
	return &HealthHandler{Version: cleanVersion}
}

func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{Status: "ok", Version: h.Version})
}
