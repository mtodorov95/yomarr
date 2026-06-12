package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/models"
)

type SystemHandler struct {
	SeriesStore  db.SeriesStore
	ChapterStore db.ChapterStore
}

type SystemStatsResponse struct {
	TotalSeries       int64 `json:"total_series"`
	DownloadedChapters int64 `json:"downloaded_chapters"`
	MissingChapters    int64 `json:"missing_chapters"`
	SizeOnDiskBytes   int64 `json:"size_on_disk_bytes"`
}

func NewSystemHandler(ss db.SeriesStore, cs db.ChapterStore) *SystemHandler {
	return &SystemHandler{
		SeriesStore:  ss,
		ChapterStore: cs,
	}
}

func (h *SystemHandler) HandleSystemStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	seriesCount, err := h.SeriesStore.Count()
	if err != nil {
		log.Printf("[System API] Error counting series: %v", err)
		seriesCount = 0
	}

	completedCount, _ := h.ChapterStore.CountByStatus(string(models.ChapterDownloaded))
	missingCount, _ := h.ChapterStore.CountByStatus(string(models.ChapterMissing))

	libraryRoot := os.Getenv("MANGA_LIBRARY_ROOT")
	if libraryRoot == "" {
		libraryRoot = "/Manga"
	}

	var sizeOnDisk int64
	err = filepath.Walk(libraryRoot, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			sizeOnDisk += info.Size()
		}
		return nil
	})
	if err != nil {
		log.Printf("[System API Warning] Error calculating disk storage profile footprint: %v", err)
	}

	resp := SystemStatsResponse{
		TotalSeries:        seriesCount,
		DownloadedChapters: completedCount,
		MissingChapters:    missingCount,
		SizeOnDiskBytes:    sizeOnDisk,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
