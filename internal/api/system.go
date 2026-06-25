package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/models"
)

type SystemHandler struct {
	SeriesStore  db.SeriesStore
	ChapterStore db.ChapterStore
	cacheMutex   sync.RWMutex
	cachedStats  *SystemStatsResponse
	lastCached   time.Time
	cacheTTL     time.Duration
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
		cacheTTL:     2 * time.Hour,
	}
}

func (h *SystemHandler) HandleSystemStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()

	h.cacheMutex.RLock()
	useCache := h.cachedStats != nil && time.Since(h.lastCached) < h.cacheTTL
	if useCache {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(h.cachedStats)
		log.Printf("[System API] Cache HIT. Served stats in %v. Next refresh in %v", time.Since(start), h.cacheTTL-time.Since(h.lastCached))
		h.cacheMutex.RUnlock()
		return
	}
	h.cacheMutex.RUnlock()

	log.Printf("[System API] Cache MISS or expired. Computing fresh stats...")

	seriesCount, err := h.SeriesStore.Count()
	if err != nil {
		log.Printf("[System API] Error counting series: %v", err)
		seriesCount = 0
	}

	completedCount, err := h.ChapterStore.CountByStatus(string(models.ChapterDownloaded))
	if err != nil {
		log.Printf("[System API Error] Failed counting downloaded chapters for state '%s': %v", string(models.ChapterDownloaded), err)
	}
	missingCount, err := h.ChapterStore.CountByStatus(string(models.ChapterMissing))
	if err != nil {
		log.Printf("[System API Error] Failed counting missing chapters for state '%s': %v", string(models.ChapterMissing), err)
	}

	libraryRoot := os.Getenv("MANGA_LIBRARY_ROOT")
	if libraryRoot == "" {
		libraryRoot = "/Manga"
	}

	log.Printf("[System API] Traversing storage footprint at path: %s", libraryRoot)
	diskStart := time.Now()
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
	log.Printf("[System API] Storage scan completed in %v", time.Since(diskStart))

	h.cacheMutex.Lock()
	h.cachedStats = &SystemStatsResponse{
		TotalSeries:        seriesCount,
		DownloadedChapters: completedCount,
		MissingChapters:    missingCount,
		SizeOnDiskBytes:    sizeOnDisk,
	}
	h.lastCached = time.Now()
	h.cacheMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.cachedStats)
	log.Printf("[System API] Cache updated successfully. Total execution time: %v", time.Since(start))
}
