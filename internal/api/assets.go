package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func FileAssetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "Missing path parameter", http.StatusBadRequest)
		return
	}

	cleanedPath := filepath.Clean(filePath)

	libraryRoot := os.Getenv("MANGA_LIBRARY_ROOT")
	if libraryRoot == "" {
		libraryRoot = "/Manga"
	}
	if !strings.HasPrefix(cleanedPath, libraryRoot) {
	    http.Error(w, "Access denied", http.StatusForbidden)
	    return
	}

	info, err := os.Stat(cleanedPath)
	if os.IsNotExist(err) || info.IsDir() {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, cleanedPath)
}
