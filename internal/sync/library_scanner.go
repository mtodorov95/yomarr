package sync

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/indexer"
	"github.com/mtodorov95/yomarr/internal/models"
)

type LibraryScanner struct {
	ChapterStore db.ChapterStore
	SeriesStore  db.SeriesStore
}

func NewLibraryScanner(cs db.ChapterStore, ss db.SeriesStore) *LibraryScanner {
	return &LibraryScanner{
		ChapterStore: cs,
		SeriesStore:  ss,
	}
}

func (ls *LibraryScanner) StartBackgroundScan(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		log.Printf("[Scanner] Background library sync daemon started. Interval: %v", interval)
		for range ticker.C {
			log.Println("[Scanner] Starting periodic library scan...")
			if err := ls.ScanLibrary(); err != nil {
				log.Printf("[Scanner Error] Periodic scan failed: %v", err)
			}
		}
	}()
}

func (ls *LibraryScanner) ScanLibrary() error {
	libraryRoot := os.Getenv("MANGA_LIBRARY_ROOT")
	if libraryRoot == "" {
		libraryRoot = "/Manga"
	}

	dirs, err := os.ReadDir(libraryRoot)
	if err != nil {
		return err
	}

	allSeries, err := ls.SeriesStore.GetAll()
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		dirNameLower := strings.ToLower(dir.Name())
		var matchedSeries *models.Series

		for i := range allSeries {
			if strings.ToLower(allSeries[i].Title) == dirNameLower {
				matchedSeries = &allSeries[i]
				break
			}
		}

		if matchedSeries == nil {
			continue
		}

		if err := ls.scanSeriesFolder(matchedSeries, filepath.Join(libraryRoot, dir.Name())); err != nil {
			log.Printf("[Scanner Error] Failed scan %s: %v", matchedSeries.Title, err)
		}
	}

	return nil
}

func (ls *LibraryScanner) scanSeriesFolder(series *models.Series, folderPath string) error {
	log.Printf("[Scanner] Starting sync for: %s", series.Title)

	dbChapters, err := ls.ChapterStore.GetBySeriesId(series.ID)
	if err != nil {
		return err
	}

	chapterMap := make(map[string]*models.Chapters)
	for i := range dbChapters {
		lang := strings.ToLower(dbChapters[i].Language)
		if lang == "" {
			lang = "en"
		}
		key := fmt.Sprintf("%g_%s", dbChapters[i].Number, lang)
		chapterMap[key] = &dbChapters[i]
	}

	foundOnDisk := make(map[string]bool)
	languages := []string{"en", "raw"}

	for _, lang := range languages {
		subFolderPath := filepath.Join(folderPath, strings.ToUpper(lang))
		files, err := os.ReadDir(subFolderPath)

		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			log.Printf("[Scanner Error] Failed to read subfolder %s: %v", subFolderPath, err)
			continue
		}

		for _, file := range files {
			fileName := file.Name()
			if strings.HasPrefix(fileName, ".") {
				continue
			}

			parsed, ok := indexer.ParseTorrentTitle(fileName)
			if !ok {
				continue
			}

			start := parsed.StartNum
			end := parsed.StartNum
			if parsed.Type == indexer.TypeRange {
				end = parsed.EndNum
			}
			currentPath := filepath.Join(subFolderPath, fileName)

			if parsed.Type == indexer.TypeVolume {
				for volNum := start; volNum <= end; volNum++ {
					for i := range dbChapters {
						ch := &dbChapters[i]
						chLang := strings.ToLower(ch.Language)
						if chLang == "" {
							chLang = "en"
						}

						if ch.Volume != nil && float64(*ch.Volume) == volNum && chLang == lang {
							diskKey := fmt.Sprintf("%g_%s", ch.Number, lang)
							foundOnDisk[diskKey] = true
							if ch.Status != models.ChapterDownloaded {
								ch.Status = models.ChapterDownloaded
								ch.FilePath = &currentPath
								if err := ls.ChapterStore.Update(ch); err != nil {
									log.Printf("[Scanner Error] Fail update Ch %g via volume: %v", ch.Number, err)
								}
							}
						}
					}
				}
				continue
			}

			for num := start; num <= end; num++ {
				diskKey := fmt.Sprintf("%g_%s", num, lang)
				foundOnDisk[diskKey] = true

				ch, exists := chapterMap[diskKey]
				if !exists {
					newCh := models.Chapters{
						SeriesID: series.ID,
						Number:   num,
						Status:   models.ChapterDownloaded,
						FilePath: &currentPath,
					}

					if err := ls.ChapterStore.Insert(&newCh); err != nil {
						log.Printf("[Scanner Error] Failed insert Ch %g: %v", num, err)
					} else {
						log.Printf("[Scanner] Created missing row for Ch %g", num)
					}
				} else {
					if ch.Status != models.ChapterDownloaded {
						ch.Status = models.ChapterDownloaded
						ch.FilePath = &currentPath
						if err := ls.ChapterStore.Update(ch); err != nil {
							log.Printf("[Scanner Error] Failed update status Ch %g: %v", num, err)
						}
					}
				}
			}
		}

	}

	for key, ch := range chapterMap {
		if ch.Status == models.ChapterDownloaded && !foundOnDisk[key] {
			log.Printf("[Scanner] Chapter %g [%s] marked Downloaded but file missing! Reverting status", ch.Number, ch.Language)
			ch.Status = models.ChapterMissing
			ch.FilePath = nil
			if err := ls.ChapterStore.Update(ch); err != nil {
				log.Printf("[Scanner Error] Failed revert status Ch %g [%s]: %v", ch.Number, ch.Language, err)
			}
		}
	}

	return nil
}
