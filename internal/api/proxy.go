package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func ProxyCoverHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "Missing url parameter", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(targetURL, "https://uploads.mangadex.org/") {
		http.Error(w, "Forbidden target domain", http.StatusForbidden)
		return
	}

	proxyReq, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		http.Error(w, "Failed to initialize request", http.StatusInternalServerError)
		return
	}

	proxyReq.Header.Set("User-Agent", "Yomarr/1.0")

	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch image: %s", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Upstream image provider error", resp.StatusCode)
		return
	}

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Cache-Control", "public, max-age=86400")

	_, _ = io.Copy(w, resp.Body)
}
