package sync

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/download"
	"github.com/mtodorov95/yomarr/internal/models"
	"github.com/mtodorov95/yomarr/internal/utils"
)

type DownloadMonitor struct {
	ChapterStore db.ChapterStore
	SeriesStore  db.SeriesStore
	QueueStore   db.QueueStore
	Downloader   download.DownloadClient
	Hub          *EventHub
}

func NewDownloadMonitor(cs db.ChapterStore, ss db.SeriesStore, dl download.DownloadClient, qs db.QueueStore, hub *EventHub) *DownloadMonitor {
	return &DownloadMonitor{
		ChapterStore: cs,
		SeriesStore:  ss,
		Downloader:   dl,
		QueueStore:   qs,
		Hub:          hub,
	}
}

func (m *DownloadMonitor) importToLibrary(series models.Series, torrentName string, language string) (string, error) {
	downloadRoot := os.Getenv("MANGA_DOWNLOAD_ROOT")
	if downloadRoot == "" {
		downloadRoot = "/downloads"
	}

	srcPath := filepath.Join(downloadRoot, torrentName)

	destDir, err := utils.EnsureLanguageDirectory(series.Path, language)
	if err != nil {
		return "", err
	}

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to inspect download source: %w", err)
	}

	if !srcInfo.IsDir() {
		finalDestPath := filepath.Join(destDir, torrentName)
		log.Printf("[Importer] Moving single file: %s -> %s", torrentName, finalDestPath)

		if err := moveOrCopyFile(srcPath, finalDestPath); err != nil {
			return "", err
		}
		return finalDestPath, nil
	}

	log.Printf("[Importer] Parsing folder structure for %s...", torrentName)

	var primaryTrackedPath string

	err = filepath.WalkDir(srcPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".cbz" || ext == ".cbr" || ext == ".zip" || ext == ".rar" {
			fileName := filepath.Base(path)
			finalDestPath := filepath.Join(destDir, fileName)

			log.Printf("[Importer] Found nested archive: %s -> %s", fileName, destDir)
			if err := moveOrCopyFile(path, finalDestPath); err != nil {
				return err
			}

			if primaryTrackedPath == "" {
				primaryTrackedPath = finalDestPath
			}
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error flattening download directory: %w", err)
	}

	if primaryTrackedPath == "" {
		primaryTrackedPath = destDir
	}

	return primaryTrackedPath, nil
}

func moveOrCopyFile(src, dest string) error {
	err := os.Rename(src, dest)
	if err == nil {
		return nil
	}

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

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	// Clean up original source file after copy success
	//  return os.Remove(src)
	return nil
}

func (m *DownloadMonitor) Start() {
	ticker := time.NewTicker(15 * time.Second)
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
	if m.Downloader == nil {
		return nil
	}

	torrents, err := m.Downloader.GetActiveDownloads()
	if err != nil {
		return err
	}

	if len(torrents) == 0 {
		return nil
	}

	allChapters, err := m.ChapterStore.GetByStatus(string(models.ChapterDownloading))
	if err != nil {
		return err
	}

	for _, torrent := range torrents {
		queueItem, err := m.QueueStore.Get(torrent.Hash)
		if err != nil || queueItem == nil {
			// Torrent isn't tracked in queue
			continue
		}

		series, err := m.SeriesStore.GetById(queueItem.SeriesID)
		if err != nil {
			log.Printf("[Monitor Error] Orphaned queue item %s linked to missing series ID %d", torrent.Hash, queueItem.SeriesID)
			continue
		}

		var releaseDetail string
		switch queueItem.ReleaseType {
		case models.TypeSingle:
			releaseDetail = fmt.Sprintf("Chapter %.1f", queueItem.StartNum)
		case models.TypeRange:
			releaseDetail = fmt.Sprintf("Chapters %.1f - %.1f", queueItem.StartNum, queueItem.EndNum)
		case models.TypeVolume:
			if queueItem.EndNum > queueItem.StartNum {
				releaseDetail = fmt.Sprintf("Volumes %.0f - %.0f", queueItem.StartNum, queueItem.EndNum)
			} else {
				releaseDetail = fmt.Sprintf("Volume %.0f", queueItem.StartNum)
			}
		default:
			releaseDetail = "Processing..."
		}

		if torrent.Progress != 1.0 {
			if queueItem.Status != models.QueueDownloading {
				_ = m.QueueStore.UpdateStatus(queueItem.TorrentHash, models.QueueDownloading, nil)
			}

			if !series.Downloading {
				series.Downloading = true 
				if err := m.SeriesStore.Update(series); err != nil {
					log.Printf("[Monitor Error] Failed to update status to downloading for series %d: %v", series.ID, err)
				}
			}

			m.Hub.Broadcast(models.QueueEvent{
				TorrentHash:   torrent.Hash,
				Status:        models.QueueDownloading,
				Progress:      torrent.Progress,
				Name:          torrent.Name,
				SeriesID:      series.ID,
				SeriesTitle:   series.Title,
				ReleaseDetail: releaseDetail,
				Error: nil,
			})
			continue
		}

		torrentLang := queueItem.Language
		if torrentLang == "" {
			torrentLang = "en"
		}

		finalLibraryPath, err := m.importToLibrary(*series, torrent.Name, torrentLang)
		if err != nil {
			log.Printf("[Monitor Error] Failed importing files to destination paths: %v", err)
			errMsg := err.Error()
			_ = m.QueueStore.UpdateStatus(queueItem.TorrentHash, models.QueueFailedImport, &errMsg)
			m.Hub.Broadcast(models.QueueEvent{
				TorrentHash:   torrent.Hash,
				Status:        models.QueueFailedImport,
				Progress:      torrent.Progress,
				Name:          torrent.Name,
				Error:         &errMsg,
				SeriesID:      series.ID,
				SeriesTitle:   series.Title,
				ReleaseDetail: releaseDetail,
			})
			continue
		}

		for i := range allChapters {
			ch := &allChapters[i]

			if ch.SeriesID != series.ID || ch.Language != torrentLang {
				continue
			}

			isMatch := false
			switch queueItem.ReleaseType {
			case models.TypeSingle:
				if queueItem.StartNum == ch.Number {
					isMatch = true
				}
			case models.TypeRange:
				if ch.Number >= queueItem.StartNum && ch.Number <= queueItem.EndNum {
					isMatch = true
				}
			case models.TypeVolume:
				if ch.Volume != nil {
					chVol := float64(*ch.Volume)
					if chVol >= queueItem.StartNum && chVol <= queueItem.EndNum {
						isMatch = true
					}
				}
			}

			if isMatch {
				ch.Status = models.ChapterDownloaded
				ch.FilePath = &finalLibraryPath
				if err := m.ChapterStore.Update(ch); err != nil {
					log.Printf("[Monitor Error] Failed updating database for Ch %g: %v", ch.Number, err)
				}
			}
		}

		hasMoreDownloads, err := m.QueueStore.HasActiveDownloadsForSeries(series.ID, torrent.Hash)
		if err != nil {
			log.Printf("[Monitor Error] Failed to check remaining queue items for series %d: %v", series.ID, err)
		}

		if !hasMoreDownloads {
			series.Downloading = false 
			if err := m.SeriesStore.Update(series); err != nil {
				log.Printf("[Monitor Error] Failed to reset status for series %d: %v", series.ID, err)
			}
		}

		m.Hub.Broadcast(models.QueueEvent{
			TorrentHash:   torrent.Hash,
			Status:        "Imported",
			Progress:      1.0,
			Name:          torrent.Name,
			SeriesID:      series.ID,
			SeriesTitle:   series.Title,
			ReleaseDetail: releaseDetail,
			Error: nil,
		})

		if err := m.Downloader.MarkAsImported(torrent.Hash); err != nil {
			log.Printf("[Monitor Error] Failed tagging single/volume torrent %s: %v", torrent.Hash, err)
		}
	}

	return nil
}
