package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/download"
	"github.com/mtodorov95/yomarr/internal/indexer"
	"github.com/mtodorov95/yomarr/internal/models"
)

type DownloadMonitor struct {
	ChapterStore db.ChapterStore
	SeriesStore  db.SeriesStore
	QBClient     *download.QBittorrentClient
}

type QBTorrentInfo struct {
	Hash     string  `json:"hash"`
	Name     string  `json:"name"`
	Progress float64 `json:"progress"`
	Status   string  `json:"status"`
}

func NewDownloadMonitor(cs db.ChapterStore, ss db.SeriesStore, qb *download.QBittorrentClient) *DownloadMonitor {
	return &DownloadMonitor{
		ChapterStore: cs,
		SeriesStore:  ss,
		QBClient:     qb,
	}
}

func (m *DownloadMonitor) importToLibrary(seriesTitle string, torrentName string) (string, error) {
	downloadRoot := os.Getenv("MANGA_DOWNLOAD_ROOT")
	if downloadRoot == "" {
		downloadRoot = "/mnt/downloads"
	}
	libraryRoot := os.Getenv("MANGA_LIBRARY_ROOT")
	if libraryRoot == "" {
		libraryRoot = "/mnt/manga"
	}

	srcPath := filepath.Join(downloadRoot, seriesTitle, torrentName)

	destDir := filepath.Join(libraryRoot, seriesTitle)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}

	destPath := filepath.Join(destDir, torrentName)

	log.Printf("[Importer] Moving media asset from %s to permanent library %s", srcPath, destPath)

	err := os.Rename(srcPath, destPath)
	if err != nil {
		// If os.Rename fails because src and dest are on separate physical drives, fallback to a manual stream copy
		log.Printf("[Importer] Direct move failed (cross-device link), falling back to deep file copy...")
		err = copyFileOrDir(srcPath, destPath)
		if err != nil {
			return "", err
		}
		// Clean up the original download workspace if you want, or leave it for qbit to seed
		// os.RemoveAll(srcPath)
	}

	return destPath, nil
}

func copyFileOrDir(src string, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(src, dest)
	}
	return copyFile(src, dest)
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dest string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())
		if err := copyFileOrDir(srcPath, destPath); err != nil {
			return err
		}
	}
	return nil
}

func (m *DownloadMonitor) Start() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		log.Println("[Monitor] Post-Download Sync Daemon initialized")
		for range ticker.C {
			if err := m.CheckActiveDownloads(); err != nil {
				log.Printf("[Monitor Error] Failed to check downloads: %v", err)
			}
		}
	}()
}

func (m *DownloadMonitor) CheckActiveDownloads() error {
	if m.QBClient == nil {
		return nil
	}

	infoURL := fmt.Sprintf("%s/api/v2/torrents/info?category=yomarr", m.QBClient.BaseURL)
	resp, err := m.QBClient.Client.Get(infoURL)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("qbit returned status code %d", resp.StatusCode)
	}

	var torrents []QBTorrentInfo
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return err
	}

	if len(torrents) == 0 {
		return nil
	}

	allSeries, err := m.SeriesStore.GetAll()
	if err != nil {
		return err
	}

	seriesMap := make(map[int64]models.Series)
	for _, s := range allSeries {
		seriesMap[s.ID] = s
	}

	allChapters, err := m.ChapterStore.GetByStatus("Downloading")
	if err != nil {
		return err
	}
	for _, torrent := range torrents {
		if torrent.Progress != 1.0 {
			continue
		}
		torrentNameLower := strings.ToLower(torrent.Name)
		parsed, ok := indexer.ParseTorrentTitle(torrent.Name)

		if !ok {
			for _, series := range allSeries {
				seriesTitleLower := strings.ToLower(series.Title)

				if strings.Contains(torrentNameLower, seriesTitleLower) &&
					(strings.Contains(torrentNameLower, "complete") || strings.Contains(torrentNameLower, "digital") || strings.Contains(torrentNameLower, "batch")) {

					log.Printf("[Monitor] Full series batch finished for: %s! Processing all chapters...", series.Title)

					for i := range allChapters {
						ch := allChapters[i]
						if ch.SeriesID == series.ID && ch.Status == "Downloading" {

							finalLibraryPath, err := m.importToLibrary(series.Title, torrent.Name)
							if err != nil {
								log.Printf("[Monitor Error] Failed importing file to library: %v", err)
								continue
							}

							ch.Status = "Downloaded"
							ch.FilePath = &finalLibraryPath

							if err := m.ChapterStore.Update(ch); err != nil {
								log.Printf("[Monitor Error] Failed batch updating Ch %g: %v", ch.Number, err)
							}
						}
					}
				}
			}
			continue
		}

		for i := range allChapters {
			ch := allChapters[i]

			isMatch := false
			switch parsed.Type {
			case indexer.TypeSingle:
				if parsed.StartNum == ch.Number {
					isMatch = true
				}
			case indexer.TypeRange:
				if ch.Number >= parsed.StartNum && ch.Number <= parsed.EndNum {
					isMatch = true
				}
			case indexer.TypeVolume:
				if ch.Volume != nil && float64(*ch.Volume) >= parsed.StartNum && float64(*ch.Volume) <= parsed.EndNum {
					isMatch = true
				}
			}

			if isMatch {
				log.Printf("[Monitor] Torrent finished! Marking Ch %g as Downloaded: %s", ch.Number, torrent.Name)

				series, exists := seriesMap[ch.SeriesID]
				if !exists {
					continue
				}

				finalLibraryPath, err := m.importToLibrary(series.Title, torrent.Name)
				if err != nil {
					log.Printf("[Monitor Error] Failed importing file to library: %v", err)
					continue
				}

				ch.Status = "Downloaded"
				ch.FilePath = &finalLibraryPath

				if err := m.ChapterStore.Update(ch); err != nil {
					log.Printf("[Monitor Error] Failed updating database for Ch %g: %v", ch.Number, err)
				}
			}
		}
	}

	return nil
}
