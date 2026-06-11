package sync

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

var folderYearRegex = regexp.MustCompile(`^(.+?)\s*\((\d{4})\)$`)

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
			log.Printf("[Scanner] Found unrecognized folder: %s. Attempting auto-creation...", dir.Name())

			cleanTitle := strings.TrimSpace(dir.Name())
			var seriesYear *int

			if matches := folderYearRegex.FindStringSubmatch(cleanTitle); len(matches) > 2 {
				cleanTitle = strings.TrimSpace(matches[1])
				if y, err := strconv.Atoi(matches[2]); err == nil {
					seriesYear = &y
				}
			}

			for i := range allSeries {
				if strings.ToLower(allSeries[i].Title) == strings.ToLower(cleanTitle) {
					matchedSeries = &allSeries[i]
					break
				}
			}

			if matchedSeries == nil {
				newSeries := models.Series{
					Title:  cleanTitle,
					Year:   seriesYear,
					Status: models.SeriesUnmonitored,
				}

				if err := ls.SeriesStore.Insert(&newSeries); err != nil {
					log.Printf("[Scanner Error] Failed to auto-create missing series row for %s: %v", cleanTitle, err)
					continue
				}

				log.Printf("[Scanner] Successfully registered new series: %s (ID: %d)", newSeries.Title, newSeries.ID)

				allSeries = append(allSeries, newSeries)
				matchedSeries = &allSeries[len(allSeries)-1]
			}
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
					// Update existing rows
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

					// Create missing rows
					ext := strings.ToLower(filepath.Ext(fileName))
					if ext == ".cbz" || ext == ".zip" {
						discoveredChapters, err := ls.extractChaptersFromArchive(currentPath, int(volNum))
						if err != nil {
							log.Printf("[Scanner Error] Failed to look inside volume archive %s: %v", fileName, err)
							continue
						}

						for chNum, assignedVolume := range discoveredChapters {
							currentDiskKey := fmt.Sprintf("%g_%s", chNum, lang)
							foundOnDisk[currentDiskKey] = true

							opposingLang := "raw"
							if lang == "raw" {
								opposingLang = "en"
							}
							opposingKey := fmt.Sprintf("%g_%s", chNum, opposingLang)

							ch, currentExists := chapterMap[currentDiskKey]
							_, opposingExists := chapterMap[opposingKey]

							if !currentExists && !opposingExists {
								vAlloc := assignedVolume

								missingTwin := models.Chapters{
									SeriesID: series.ID,
									Number:   chNum,
									Volume:   &vAlloc,
									Status:   models.ChapterMissing,
									FilePath: nil,
									Language: opposingLang,
								}
								if err := ls.ChapterStore.Insert(&missingTwin); err != nil {
									log.Printf("[Scanner Error] Failed to insert parallel placeholder row for Ch %g [%s]: %v", chNum, opposingLang, err)
								} else {
									chapterMap[opposingKey] = &missingTwin
								}
							}

							if !currentExists {
								vAlloc := assignedVolume
								newCh := models.Chapters{
									SeriesID: series.ID,
									Number:   chNum,
									Volume:   &vAlloc,
									Status:   models.ChapterDownloaded,
									FilePath: &currentPath,
									Language: lang,
								}

								if err := ls.ChapterStore.Insert(&newCh); err != nil {
									log.Printf("[Scanner Error] Failed to insert newly discovered internal Ch %g: %v", chNum, err)
								} else {
									log.Printf("[Scanner] Found chapter Ch %g from Volume %d archive!", chNum, vAlloc)
									chapterMap[currentDiskKey] = &newCh
								}
							} else {
								vAlloc := assignedVolume
								if ch.Status != models.ChapterDownloaded {
									ch.Status = models.ChapterDownloaded
									ch.Volume = &vAlloc
									ch.FilePath = &currentPath
									if err := ls.ChapterStore.Update(ch); err != nil {
										log.Printf("[Scanner Error] Failed updating internal Ch %g: %v", chNum, err)
									}
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

// Matches: "- p086-087", "- p000", "_p050", " page 12", " p.012"
var pageCleanerRegex = regexp.MustCompile(`(?i)[-_\s]p(?:age|[\s.])?\d+(?:\s*-\s*\d+)?`)

func (ls *LibraryScanner) extractChaptersFromArchive(archivePath string, fallbackVol int) (map[float64]int, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	chapterToVolumeMap := make(map[float64]int)

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		internalPath := filepath.ToSlash(f.Name)
		parts := strings.Split(internalPath, "/")

		var tokenToParse string

		if len(parts) > 1 {
			// Nested in a folder
			tokenToParse = parts[len(parts)-2]
		} else {
			// File in root
			tokenToParse = parts[0]
		}

		if strings.HasPrefix(filepath.Base(tokenToParse), ".") {
			continue
		}

		var fileSpecificVol *int
		if volMatches := indexer.VolRegex.FindStringSubmatch(tokenToParse); len(volMatches) > 1 {
			if v, err := strconv.Atoi(volMatches[1]); err == nil {
				fileSpecificVol = &v
			}
		} else if jaVolMatches := indexer.VolJaRegex.FindStringSubmatch(f.Name); len(jaVolMatches) > 1 {
			if v, err := strconv.Atoi(jaVolMatches[1]); err == nil {
				fileSpecificVol = &v
			}
		}

		cleanedName := pageCleanerRegex.ReplaceAllString(tokenToParse, " ")

		if volMatch := indexer.VolRegex.FindString(cleanedName); volMatch != "" {
			cleanedName = strings.ReplaceAll(cleanedName, volMatch, " ")
		}
		if jaVolMatch := indexer.VolJaRegex.FindString(cleanedName); jaVolMatch != "" {
			cleanedName = strings.ReplaceAll(cleanedName, jaVolMatch, " ")
		}

		parsed, ok := indexer.ParseTorrentTitle(cleanedName)
		if !ok {
			continue
		}

		targetVolume := fallbackVol
		if fileSpecificVol != nil {
			targetVolume = *fileSpecificVol
		}

		if parsed.Type == indexer.TypeSingle || parsed.Type == indexer.TypeRange {
			end := parsed.StartNum
			if parsed.Type == indexer.TypeRange {
				end = parsed.EndNum
			}

			for num := parsed.StartNum; num <= end; num++ {
				chapterToVolumeMap[num] = targetVolume
			}
		}
	}

	return chapterToVolumeMap, nil
}
