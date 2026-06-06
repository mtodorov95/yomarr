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
	ChapterStore   db.ChapterStore
	SeriesStore    db.SeriesStore
	QBClient       *download.QBittorrentClient
	importedHashes map[string]bool
}

type QBTorrentInfo struct {
	Hash     string  `json:"hash"`
	Name     string  `json:"name"`
	Progress float64 `json:"progress"`
	Status   string  `json:"status"`
	Tags     string  `json:"tags"`
}

func NewDownloadMonitor(cs db.ChapterStore, ss db.SeriesStore, qb *download.QBittorrentClient) *DownloadMonitor {
	return &DownloadMonitor{
		ChapterStore:   cs,
		SeriesStore:    ss,
		QBClient:       qb,
		importedHashes: make(map[string]bool),
	}
}

func torrentMatchesSeries(torrentNameLower string, series models.Series) bool {
	if strings.Contains(torrentNameLower, strings.ToLower(series.Title)) {
		return true
	}
	for _, alt := range series.AltTitles {
		if alt != "" && strings.Contains(torrentNameLower, strings.ToLower(alt)) {
			return true
		}
	}
	return false
}

func (m *DownloadMonitor) importToLibrary(seriesTitle string, torrentName string, language string) (string, error) {
	downloadRoot := os.Getenv("MANGA_DOWNLOAD_ROOT")
	if downloadRoot == "" {
		downloadRoot = "/downloads"
	}
	libraryRoot := os.Getenv("MANGA_LIBRARY_ROOT")
	if libraryRoot == "" {
		libraryRoot = "/Manga"
	}

	srcPath := filepath.Join(downloadRoot, torrentName)

	langDir := "EN"
	if strings.ToLower(language) == "raw" {
		langDir = "RAW"
	}

	destDir := filepath.Join(libraryRoot, seriesTitle, langDir)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}

	finalDestPath := filepath.Join(destDir, torrentName)

	log.Printf("[Importer] Moving media asset from %s to permanent library %s", srcPath, finalDestPath)

	err := os.Rename(srcPath, finalDestPath)
	if err != nil {
		log.Printf("[Importer] Direct move failed (cross-device link), falling back to deep file copy...")
		err = copyFileOrDir(srcPath, finalDestPath)
		if err != nil {
			return "", err
		}
	}

	return finalDestPath, nil
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

	allChapters, err := m.ChapterStore.GetByStatus(string(models.ChapterDownloading))
	if err != nil {
		return err
	}

	for _, torrent := range torrents {
		if torrent.Progress != 1.0 || m.importedHashes[torrent.Hash] {
			continue
		}

		torrentNameLower := strings.ToLower(torrent.Name)
		parsed, ok := indexer.ParseTorrentTitle(torrent.Name)

		torrentLang := "en"
		if strings.Contains(strings.ToLower(torrent.Tags), "raw") {
			torrentLang = "raw"
		}

		if !ok {
			for _, series := range allSeries {
				if torrentMatchesSeries(torrentNameLower, series) &&
					(strings.Contains(torrentNameLower, "complete") || strings.Contains(torrentNameLower, "digital") || strings.Contains(torrentNameLower, "batch")) {

					log.Printf("[Monitor] Full series batch finished for: %s! Processing all chapters...", series.Title)

					finalLibraryPath, err := m.importToLibrary(series.Title, torrent.Name, torrentLang)
					if err != nil {
						log.Printf("[Monitor Error] Failed importing file to library: %v", err)
						continue
					}

					for i := range allChapters {
						ch := &allChapters[i]
						chLang := ch.Language
						if chLang == "" {
							chLang = "en"
						}

						if ch.SeriesID == series.ID && ch.Status == models.ChapterDownloading && chLang == torrentLang {
							ch.Status = models.ChapterDownloaded
							ch.FilePath = &finalLibraryPath
							if err := m.ChapterStore.Update(ch); err != nil {
								log.Printf("[Monitor Error] Failed batch updating Ch %g: %v", ch.Number, err)
							}
						}
					}

					m.importedHashes[torrent.Hash] = true
				}
			}
			continue
		}

		torrentImported := false
		for i := range allChapters {
			ch := &allChapters[i]

			series, exists := seriesMap[ch.SeriesID]
			if !exists {
				continue
			}

			if !torrentMatchesSeries(torrentNameLower, series) {
				continue
			}

			chLang := ch.Language
			if chLang == "" {
				chLang = "en"
			}
			if chLang != torrentLang {
				continue
			}

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

				finalLibraryPath, err := m.importToLibrary(series.Title, torrent.Name, torrentLang)
				if err != nil {
					log.Printf("[Monitor Error] Failed importing file to library: %v", err)
					continue
				}

				ch.Status = models.ChapterDownloaded
				ch.FilePath = &finalLibraryPath

				if err := m.ChapterStore.Update(ch); err != nil {
					log.Printf("[Monitor Error] Failed updating database for Ch %g: %v", ch.Number, err)
				}

				torrentImported = true
			}
		}

		if torrentImported {
			m.importedHashes[torrent.Hash] = true
		}
	}

	return nil
}
