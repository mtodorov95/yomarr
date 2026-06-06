package api

import (
	"net/http"

	"github.com/mtodorov95/yomarr/internal/sync"
)

type LibraryHandler struct {
	Scanner *sync.LibraryScanner	
}

func NewLibraryHandler(s *sync.LibraryScanner) *LibraryHandler {
	return &LibraryHandler{Scanner: s}
}

func (h *LibraryHandler) ScanLibrary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		_ = h.Scanner.ScanLibrary();
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"scanning"}`))
}

