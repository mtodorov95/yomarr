package sync

import (
	"archive/zip"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/download"
	"github.com/mtodorov95/yomarr/internal/indexer"
	"github.com/mtodorov95/yomarr/internal/metadata"
	"github.com/mtodorov95/yomarr/internal/models"
)

type LibraryScanner struct {
	ChapterStore db.ChapterStore
	SeriesStore  db.SeriesStore
	Metadata     metadata.Provider
	SyncEngine   *MangaDexSyncEngine
}

func NewLibraryScanner(cs db.ChapterStore, ss db.SeriesStore, md metadata.Provider, se *MangaDexSyncEngine) *LibraryScanner {
	return &LibraryScanner{
		ChapterStore: cs,
		SeriesStore:  ss,
		Metadata:     md,
		SyncEngine:   se,
	}
}

func (ls *LibraryScanner) StartBackgroundMetadataRefresher(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		log.Printf("[Scanner] Background metadata refresher started. Interval: %v", interval)
		for range ticker.C {
			log.Println("[Scanner] Starting global metadata and chapter refresh...")

			allSeries, err := ls.SeriesStore.GetAll()
			if err != nil {
				log.Printf("[Scanner Error] Could not fetch series lists: %v", err)
				continue
			}

			for _, s := range allSeries {
				log.Printf("[Scanner] Auto-refreshing details for: %s", s.Title)
				if _, err := ls.RefreshSeriesMetadata(s.ID); err != nil {
					log.Printf("[Scanner Error] Upstream sync failed for %s: %v", s.Title, err)
				}

				time.Sleep(5 * time.Second)
			}
			log.Println("[Scanner] Global metadata sync cycle complete.")
		}
	}()
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

		currentDirPath := filepath.Clean(filepath.Join(libraryRoot, dir.Name()))
		var matchedSeries *models.Series

		for i := range allSeries {
			dbPath := filepath.Clean(allSeries[i].Path)

			if dbPath == currentDirPath {
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
				folderPath := filepath.Join(libraryRoot, dir.Name())

				newSeries := models.Series{
					Title:       cleanTitle,
					Year:        seriesYear,
					Status:      models.SeriesUnknown,
					Downloading: false,
					Monitored:   false,
					Path:        folderPath,
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

			cleanedName := cleanLooseFilename(fileName)

			parsed, ok := indexer.ParseTorrentTitle(cleanedName)
			if !ok {
				continue
			}

			start := parsed.StartNum
			end := parsed.EndNum
			currentPath := filepath.Join(subFolderPath, fileName)

			if parsed.Type == models.TypeVolume {
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
						Language: lang,
					}

					if err := ls.ChapterStore.Insert(&newCh); err != nil {
						log.Printf("[Scanner Error] Failed insert Ch %g: %v", num, err)
					} else {
						log.Printf("[Scanner] Created missing row for Ch %g", num)
						chapterMap[diskKey] = &newCh
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

var looseVolumeLetterRegex = regexp.MustCompile(`(?i)\b(v\d+)[a-z](-\d+)?\b`)

func cleanLooseFilename(filename string) string {
	if looseVolumeLetterRegex.MatchString(filename) {
		filename = looseVolumeLetterRegex.ReplaceAllString(filename, "$1$2")
	}
	return filename
}

// Matches: "- p086-087", "- p000", "_p050", " page 12", " p.012"
var pageCleanerRegex = regexp.MustCompile(`(?i)[-_\s]p(?:age|[\s.])?\d+(?:\s*-\s*\d+)?`)

// Matches 4-digit years bound by spaces, dashes, or brackets: e.g., "- 2018", "[2025]", " 1999 "
var yearCleanerRegex = regexp.MustCompile(`(?i)[-_\s]*\[?\b(19\d{2}|20\d{2})\b\]?`)

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
		cleanedName = yearCleanerRegex.ReplaceAllString(cleanedName, " ")

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

		if parsed.Type == models.TypeSingle || parsed.Type == models.TypeRange {
			end := parsed.StartNum
			if parsed.Type == models.TypeRange {
				end = parsed.EndNum
			}

			if end-parsed.StartNum > 500 {
				log.Printf("[Scanner Warning] Skipping suspicious chapter range : %v to %v in %s", parsed.StartNum, end, tokenToParse)
				continue
			}

			for num := parsed.StartNum; num <= end; num++ {
				chapterToVolumeMap[num] = targetVolume
			}
		}
	}

	return chapterToVolumeMap, nil
}

func (ls *LibraryScanner) RefreshSeriesMetadata(id int64) (*models.Series, error) {
	s, err := ls.SeriesStore.GetById(id)
	if err != nil {
		return nil, fmt.Errorf("database lookup failure: %w", err)
	}
	if s == nil {
		return nil, fmt.Errorf("series with ID %d not found", id)
	}

	var targetMDID string
	if s.MangadexID != nil && *s.MangadexID != "" {
		targetMDID = *s.MangadexID
	} else {
		log.Printf("[Metadata Service] Missing MangaDex ID for %s. Executing search fallback...", s.Title)
		results, err := ls.Metadata.Search(s.Title)
		if err != nil {
			return nil, fmt.Errorf("upstream search failed: %w", err)
		}
		if len(results) > 0 {
			targetMDID = *results[0].MangadexID
		}
	}

	if targetMDID == "" {
		return nil, fmt.Errorf("could not map series %q against external MangaDex records", s.Title)
	}

	extSeries, err := ls.Metadata.GetDetails(targetMDID)
	if err != nil {
		return nil, fmt.Errorf("metadata aggregation fetch fail: %w", err)
	}

	s.Status = extSeries.Status
	s.LastChapter = extSeries.LastChapter
	s.LastVolume = extSeries.LastVolume
	s.TotalChapters = extSeries.TotalChapters
	s.AltTitles = extSeries.AltTitles

	if (s.MangadexID == nil || *s.MangadexID == "") && extSeries.MangadexID != nil {
		s.MangadexID = extSeries.MangadexID
	}
	if (s.AnilistID == nil || *s.AnilistID == "") && extSeries.AnilistID != nil {
		s.AnilistID = extSeries.AnilistID
	}
	if (s.Artist == nil || *s.Artist == "") && extSeries.Artist != nil {
		s.Artist = extSeries.Artist
	}
	if (s.Author == nil || *s.Author == "") && extSeries.Author != nil {
		s.Author = extSeries.Author
	}
	if (s.Description == nil || *s.Description == "") && extSeries.Description != nil {
		s.Description = extSeries.Description
	}
	if s.Year == nil && extSeries.Year != nil {
		s.Year = extSeries.Year
	}
	if len(s.Genres) == 0 && len(extSeries.Genres) > 0 {
		s.Genres = extSeries.Genres
	}

	if s.Path != "" {
		existingFiles := make(map[string]bool)
		for _, hc := range s.HistoricalCovers {
			fileName := filepath.Base(hc.URL)
			existingFiles[fileName] = true
		}

		var newHistoricalCoversToDownload []models.VolumeCover
		for _, extHC := range extSeries.HistoricalCovers {
			parts := strings.Split(extHC.URL, "/")
			remoteFileName := parts[len(parts)-1]

			if !existingFiles[remoteFileName] {
				newHistoricalCoversToDownload = append(newHistoricalCoversToDownload, extHC)
			}
		}

		if s.Thumbnail == "" || len(newHistoricalCoversToDownload) > 0 {
			log.Printf("[Metadata Service] Gathering missing tracking artwork down to: %s/Covers", s.Path)

			localThumb, localHists, err := download.DownloadSeriesCovers(
				http.DefaultClient,
				s.Path,
				extSeries.Thumbnail,
				newHistoricalCoversToDownload,
			)
			if err != nil {
				log.Printf("[Metadata Service Warning] Cover refresh incomplete for %s: %v", s.Title, err)
			} else {
				if localThumb != "" {
					s.Thumbnail = localThumb
				}
				s.HistoricalCovers = append(s.HistoricalCovers, localHists...)
			}
		}
	}

	if err := ls.SeriesStore.Update(s); err != nil {
		return nil, fmt.Errorf("database updates sync failure: %w", err)
	}

	if s.MangadexID != nil && *s.MangadexID != "" {
		if err := ls.SyncEngine.SyncChapters(s.ID, *s.MangadexID); err != nil {
			log.Printf("[Metadata Warning] Non-fatal chapter tracking sync error for %s: %v", s.Title, err)
		}
	}

	return s, nil
}
