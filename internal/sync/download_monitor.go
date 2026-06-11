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
	"github.com/mtodorov95/yomarr/internal/utils"
)

const ImportedTag = "imported"

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
	Tags     string  `json:"tags"`
}

func NewDownloadMonitor(cs db.ChapterStore, ss db.SeriesStore, qb *download.QBittorrentClient) *DownloadMonitor {
	return &DownloadMonitor{
		ChapterStore: cs,
		SeriesStore:  ss,
		QBClient:     qb,
	}
}

func torrentMatchesSeries(torrentNameLower string, series models.Series) bool {
	normalize := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "-", "")
		s = strings.ReplaceAll(s, "_", "")
		s = strings.ReplaceAll(s, ".", "")
		return s
	}

	cleanTorrent := normalize(torrentNameLower)

	if cleanSeries := normalize(series.Title); cleanSeries != "" {
		if strings.Contains(cleanTorrent, cleanSeries) {
			return true
		}
	}

	for _, alt := range series.AltTitles {
		if cleanAlt := normalize(alt); cleanAlt != "" {
			if strings.Contains(cleanTorrent, cleanAlt) {
				return true
			}
		}
	}
	return false
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
		if torrent.Progress != 1.0 {
			continue
		}

		isAlreadyImported := false
		torrentLang := "en"

		tagList := strings.Split(torrent.Tags, ",")
		for _, tag := range tagList {
			if strings.TrimSpace(tag) == ImportedTag {
				isAlreadyImported = true
			}
			if strings.TrimSpace(tag) == "raw" {
				torrentLang = "raw"
			}
		}

		if isAlreadyImported {
			continue
		}

		torrentNameLower := strings.ToLower(torrent.Name)
		parsed, ok := indexer.ParseTorrentTitle(torrent.Name)

		isMultiVolumePack := ok && parsed.Type == indexer.TypeVolume && parsed.EndNum > parsed.StartNum
		isBatchText := strings.Contains(torrentNameLower, "complete") || strings.Contains(torrentNameLower, "digital") || strings.Contains(torrentNameLower, "batch")

		if !ok || isMultiVolumePack || isBatchText {
			for _, series := range allSeries {
				if torrentMatchesSeries(torrentNameLower, series) {

					if isMultiVolumePack {
						hasMatchingVolChapter := false
						for _, ch := range allChapters {
							if ch.SeriesID == series.ID && ch.Volume != nil {
								chVol := float64(*ch.Volume)
								if chVol >= parsed.StartNum && chVol <= parsed.EndNum {
									hasMatchingVolChapter = true
									break
								}
							}
						}
						if !hasMatchingVolChapter {
							continue
						}

					}

					log.Printf("[Monitor] Multi-volume or batch release finished for: %s! Processing library mapping...", series.Title)

					finalLibraryPath, err := m.importToLibrary(series, torrent.Name, torrentLang)
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
							if isMultiVolumePack && ch.Volume != nil {
								chVol := float64(*ch.Volume)
								if chVol < parsed.StartNum || chVol > parsed.EndNum {
									continue
								}
							}

							ch.Status = models.ChapterDownloaded
							ch.FilePath = &finalLibraryPath
							if err := m.ChapterStore.Update(ch); err != nil {
								log.Printf("[Monitor Error] Failed batch updating Ch %g: %v", ch.Number, err)
							}
						}
					}

					if err := m.QBClient.AddTorrentTags(torrent.Hash, ImportedTag); err != nil {
						log.Printf("[Monitor Error] Failed tagging batch torrent %s: %v", torrent.Hash, err)
					}
				}
			}
			continue
		}

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

				finalLibraryPath, err := m.importToLibrary(series, torrent.Name, torrentLang)
				if err != nil {
					log.Printf("[Monitor Error] Failed importing file to library: %v", err)
					continue
				}

				ch.Status = models.ChapterDownloaded
				ch.FilePath = &finalLibraryPath

				if err := m.ChapterStore.Update(ch); err != nil {
					log.Printf("[Monitor Error] Failed updating database for Ch %g: %v", ch.Number, err)
				}

				if err := m.QBClient.AddTorrentTags(torrent.Hash, ImportedTag); err != nil {
					log.Printf("[Monitor Error] Failed tagging single/volume torrent %s: %v", torrent.Hash, err)
				}
			}
		}
	}

	return nil
}
