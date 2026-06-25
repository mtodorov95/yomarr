package sync

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/download"
	"github.com/mtodorov95/yomarr/internal/indexer"
	"github.com/mtodorov95/yomarr/internal/models"
)

type SearchEngine struct {
	ChapterStore db.ChapterStore
	SeriesStore  db.SeriesStore
	Indexer      indexer.Indexer
	Downloader   download.DownloadClient
}

func NewSearchEngine(cs db.ChapterStore, ss db.SeriesStore, idx indexer.Indexer, dl download.DownloadClient) *SearchEngine {
	return &SearchEngine{
		ChapterStore: cs,
		SeriesStore:  ss,
		Indexer:      idx,
		Downloader:   dl,
	}
}

func getDownloadsPath() string {
	downloadRoot := os.Getenv("MANGA_DOWNLOAD_ROOT")
	if downloadRoot == "" {
		return "/downloads"
	}

	return downloadRoot
}

func (e *SearchEngine) StartBackgroundSearcher(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		log.Printf("[Automation] Background backlog searcher engine starting. Interval: %v", interval)
		for range ticker.C {
			monitoredSeries, err := e.SeriesStore.GetAllMonitored()
			if err != nil {
				log.Printf("[Automation] Failed to retrieve monitored series: %v", err)
				continue
			}

			for _, series := range monitoredSeries {
				log.Printf("[Automation] Triggering missing chapter check for: %s", series.Title)

				if err := e.FindMissingChapters(series.ID); err != nil {
					log.Printf("[Automation] Search run failed for %s: %v", series.Title, err)
				}

				time.Sleep(15 * time.Second)
			}
			log.Println("[Automation] Global backlog search cycle complete.")
		}
	}()
}

func (e *SearchEngine) FindMissingChapters(seriesID int64) error {
	if e.Indexer == nil || e.Downloader == nil {
		log.Println("[Sync] Indexer or Downloader not configured yet.")
		return nil
	}

	series, err := e.SeriesStore.GetById(seriesID)
	if err != nil {
		log.Printf("Err: %v", err)
		return err
	}

	if series.Status == models.SeriesUnmonitored {
		log.Printf("[Sync] Skipping search. Series '%s' is explicitly Unmonitored.", series.Title)
		return nil
	}

	missing, err := e.ChapterStore.GetMissingBySeriesID(seriesID)
	if err != nil || len(missing) == 0 {
		return err
	}

	missingLanguages := make(map[string]bool)
	for _, ch := range missing {
		missingLanguages[ch.Language] = true
	}

	searchQueries := []string{series.Title}

	for lang := range missingLanguages {
		if langTitles, ok := series.AltTitles[lang]; ok {
			searchQueries = append(searchQueries, langTitles...)
		}

		if lang == "raw" {
			if roTitles, ok := series.AltTitles["ja-ro"]; ok {
				searchQueries = append(searchQueries, roTitles...)
			}
			if jaTitles, ok := series.AltTitles["ja"]; ok {
				searchQueries = append(searchQueries, jaTitles...)
			}
		}
	}

	if len(searchQueries) == 1 && series.AltTitles == nil {
		log.Printf("[Sync Warning] Series %s might contain legacy unmapped alternative titles.", series.Title)
	}

	var results []indexer.SearchResult
	seenTorrents := make(map[string]bool)

	for _, queryTitle := range searchQueries {
		log.Printf("Executing targeted search for: %s", queryTitle)
		variantResults, err := e.Indexer.Search(queryTitle)
		if err != nil {
			log.Printf("Search error for variant '%s': %v", queryTitle, err)
			continue
		}

		if len(variantResults) > 0 {
			log.Printf("Found %d results using title variant: %s", len(variantResults), queryTitle)
			for _, res := range variantResults {
				if !seenTorrents[res.InfoHash] {
					seenTorrents[res.InfoHash] = true
					results = append(results, res)
				}
			}
		}
	}

	if len(results) == 0 {
		log.Printf("[Search] Total search blackout across all known titles for: %s", series.Title)
		return nil
	}

	downloadedTorrents := make(map[string]bool)

	for _, ch := range missing {
		var bestTorrent *indexer.SearchResult
		var bestParsed indexer.ParsedRelease
		maxSeeders := -1

		for _, res := range results {

			parsedRelease, matched := IsChapterMatch(res, ch, series)
			if !matched {
				continue
			}

			if res.Seeders > maxSeeders {
				maxSeeders = res.Seeders
				tmp := res
				bestTorrent = &tmp
				bestParsed = parsedRelease
			}
		}

		if bestTorrent != nil {
			if downloadedTorrents[bestTorrent.InfoHash] {
				if bestTorrent.Language == ch.Language {
					ch.Status = models.ChapterDownloading
					_ = e.ChapterStore.Update(ch)
				}
				continue
			}

			targetPath := getDownloadsPath()
			err := os.MkdirAll(targetPath, 0755)
			if err != nil {
				log.Printf("[Search] Failed to create local directory %s: %v", targetPath, err)
			}

			log.Printf("[Search] Found optimal release for %s Ch %g [%s] -> %s (Seeds: %d)",
				series.Title, ch.Number, bestTorrent.Language, bestTorrent.Title, bestTorrent.Seeders)

			err = e.Downloader.AddTorrentFromURL(
				bestTorrent.Link,
				targetPath,
				bestTorrent.SeedTime,
				bestTorrent.Language,
				series.ID,
				bestParsed,
				bestTorrent.InfoHash,
			)
			if err != nil {
				log.Printf("[Search] Failed to dispatch torrent to client: %v", err)
				continue
			}

			downloadedTorrents[bestTorrent.InfoHash] = true
			if bestTorrent.Language == ch.Language {
				ch.Status = models.ChapterDownloading
				_ = e.ChapterStore.Update(ch)
			}
		} else {
			log.Printf("[Search] No available candidate matches %s Ch %g for language [%s]",
				series.Title, ch.Number, ch.Language)
		}
	}

	return nil
}

func IsChapterMatch(res indexer.SearchResult, ch *models.Chapters, series *models.Series) (indexer.ParsedRelease, bool) {
	titleLower := strings.ToLower(res.Title)

	if !torrentMatchesSeries(titleLower, series) {
		return indexer.ParsedRelease{}, false
	}
	
	if res.Language != ch.Language {
		return indexer.ParsedRelease{}, false
	}

	if strings.Contains(titleLower, "ln") ||
		strings.Contains(titleLower, "novel") ||
		strings.Contains(titleLower, "wn") ||
		strings.Contains(titleLower, "epub") ||
		strings.Contains(titleLower, "pdf") {
		return indexer.ParsedRelease{}, false
	}

	parsed, ok := indexer.ParseTorrentTitle(res.Title)

	if !ok {
		if strings.Contains(titleLower, "complete") || strings.Contains(titleLower, "digital") {
			return indexer.ParsedRelease{
				Type: models.TypeRange,
				StartNum: 0,
				EndNum: 9999,
			}, true
		}
		return indexer.ParsedRelease{}, false
	}

	switch parsed.Type {
	case models.TypeSingle:
		if parsed.StartNum == ch.Number {
            return parsed, true
        }

	case models.TypeRange:
		if ch.Number >= parsed.StartNum && ch.Number <= parsed.EndNum {
            return parsed, true
        }

	case models.TypeVolume:
		if ch.Volume != nil {
			chVol := float64(*ch.Volume)
			if chVol >= parsed.StartNum && chVol <= parsed.EndNum {
                return parsed, true
            }
		}
	}

	return indexer.ParsedRelease{}, false
}

func torrentMatchesSeries(torrentNameLower string, series *models.Series) bool {
	normalize := func(s string) string {
		replacer := strings.NewReplacer(
			" ", "", "-", "", "_", "", ".", "",
			":", "", "?", "", "!", "", `"`, "",
			`'`, "", "`", "", "/", "", "\\", "",
			"(", "", ")", "", "[", "", "]", "",
			"★", "", "☆", "",
		)

		return strings.ToLower(replacer.Replace(s))
	}

	cleanTorrent := normalize(torrentNameLower)

	if cleanSeries := normalize(series.Title); cleanSeries != "" {
		if strings.Contains(cleanTorrent, cleanSeries) {
			return true
		}
	}

	for _, titlesSlice := range series.AltTitles {
		for _, alt := range titlesSlice {
			if cleanAlt := normalize(alt); cleanAlt != "" {
				if strings.Contains(cleanTorrent, cleanAlt) {
					return true
				}
			}
		}
	}
	return false
}

