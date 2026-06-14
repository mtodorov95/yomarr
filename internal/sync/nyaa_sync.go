package sync

import (
	"log"
	"os"
	"strings"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/download"
	"github.com/mtodorov95/yomarr/internal/indexer"
	"github.com/mtodorov95/yomarr/internal/models"
)

type NyaaSyncEngine struct {
	ChapterStore db.ChapterStore
	SeriesStore  db.SeriesStore
	Indexer      indexer.Indexer
	Downloader   download.DownloadClient
}

func NewNyaaSyncEngine(cs db.ChapterStore, ss db.SeriesStore, idx indexer.Indexer, dl download.DownloadClient) *NyaaSyncEngine {
	return &NyaaSyncEngine{
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

func (e *NyaaSyncEngine) FindMissingChapters(seriesID int64) error {
	if e.Indexer == nil || e.Downloader == nil {
        log.Println("[Sync] Indexer or Downloader not configured yet.")
        return nil
    }

	series, err := e.SeriesStore.GetById(seriesID)
	if err != nil {
		log.Printf("Err: %v", err)
		return err
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
		log.Printf("Executing targeted search on Nyaa for: %s", queryTitle)
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
		log.Printf("Total search blackout on Nyaa across relevant titles for: %s", series.Title)
		return nil
	}

	downloadedTorrents := make(map[string]bool)

	for _, ch := range missing {
		var bestTorrent *indexer.SearchResult
		maxSeeders := -1

		for _, res := range results {
			if res.Language != ch.Language {
				continue
			}

			parsed, ok := indexer.ParseTorrentTitle(res.Title)
			isMatch := false
			titleLower := strings.ToLower(res.Title)

			if !ok {
				if !strings.Contains(titleLower, "ln") &&
					!strings.Contains(titleLower, "novel") &&
					!strings.Contains(titleLower, "wn") &&
					!strings.Contains(titleLower, "epub") &&
					!strings.Contains(titleLower, "pdf") {
					if strings.Contains(titleLower, "complete") || strings.Contains(titleLower, "digital") {
						isMatch = true
					}
				}
			} else {
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
					if ch.Volume != nil {
						chVol := float64(*ch.Volume)
						if chVol >= parsed.StartNum && chVol <= parsed.EndNum {
							isMatch = true
						}
					}
				}
			}

			if isMatch && res.Seeders > maxSeeders {
				maxSeeders = res.Seeders
				tmp := res
				bestTorrent = &tmp
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
				log.Printf("Failed to create local directory %s: %v", targetPath, err)
			}

			log.Printf("Found optimal release for %s Ch %g [%s] -> %s (Seeds: %d)",
				series.Title, ch.Number, bestTorrent.Language, bestTorrent.Title, bestTorrent.Seeders)

			err = e.Downloader.AddTorrentFromURL(bestTorrent.Link, targetPath, bestTorrent.SeedTime, bestTorrent.Language)
			if err != nil {
				log.Printf("Failed to dispatch torrent to client: %v", err)
				continue
			}

			downloadedTorrents[bestTorrent.InfoHash] = true
			if bestTorrent.Language == ch.Language {
				ch.Status = models.ChapterDownloading
				_ = e.ChapterStore.Update(ch)
			}
		} else {
			log.Printf("No available candidate on Nyaa matches %s Ch %g for language [%s]",
				series.Title, ch.Number, ch.Language)
		}
	}

	return nil
}
