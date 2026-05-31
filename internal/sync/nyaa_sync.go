package sync

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/download"
	"github.com/mtodorov95/yomarr/internal/indexer"
)

type NyaaSyncEngine struct {
	ChapterStore db.ChapterStore
	SeriesStore  db.SeriesStore
	Indexer      *indexer.NyaaIndexer
	Downloader   download.DownloadClient
}

func NewNyaaSyncEngine(cs db.ChapterStore, ss db.SeriesStore, idx *indexer.NyaaIndexer, dl download.DownloadClient) *NyaaSyncEngine {
	return &NyaaSyncEngine{
		ChapterStore: cs,
		SeriesStore:  ss,
		Indexer:      idx,
		Downloader:   dl,
	}
}

func translatePath(originalPath string) string {
	originalPath = strings.TrimSpace(originalPath)

	localRoot := os.Getenv("MANGA_ROOT")
	if localRoot == "" {
	log.Printf("Path: %s", originalPath)
		return originalPath
	}

	if strings.HasPrefix(originalPath, "/mnt/manga") {
		relPath := strings.TrimPrefix(originalPath, "/mnt/manga")
		relPath = strings.TrimPrefix(relPath, "/")

	log.Printf("Path: %s/%s", localRoot, relPath)
		return filepath.Join(localRoot, relPath)
	}
	log.Printf("Path: %s", originalPath)

	return originalPath
}

func (e *NyaaSyncEngine) FindMissingChapters(seriesID int64) error {
	series, err := e.SeriesStore.GetById(seriesID)
	if err != nil {
		log.Printf("Err: %v", err)
		return err
	}

	missing, err := e.ChapterStore.GetMissingBySeriesID(seriesID)
	if err != nil || len(missing) == 0 {
		return err
	}

	searchQueries := append([]string{series.Title}, series.AltTitles...)
	var results []indexer.NyaaResult

	for _, queryTitle := range searchQueries {
		log.Printf("Executing broad search on Nyaa for: %s", queryTitle)
		results, err = e.Indexer.Search(queryTitle)
		if err == nil && len(results) > 0 {
			log.Printf("Found %d results using title variant: %s", len(results), queryTitle)
			break
		}
	}

	log.Printf("Results: %v", results)

	if len(results) == 0 {
		log.Printf("Total search blackout on Nyaa across all known titles for: %s", series.Title)
		return nil
	}

	downloadedTorrents := make(map[string]bool)

	for _, ch := range missing {
		var bestTorrent *indexer.NyaaResult
		maxSeeders := -1

		for _, res := range results {
			parsed, ok := indexer.ParseTorrentTitle(res.Title)

			isMatch := false

			if !ok {
				if strings.Contains(strings.ToLower(res.Title), "complete") || strings.Contains(strings.ToLower(res.Title), "digital") {
					isMatch = true
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
				// Already queued
				ch.Status = "Downloading"
				// _ = e.ChapterStore.Update(ch)
				continue
			}

			targetPath := translatePath(series.Path)

			log.Printf("Preparing download folder on disk: %s", targetPath)

			err := os.MkdirAll(targetPath, 0755)
			if err != nil {
				log.Printf("Failed to create local directory %s: %v", targetPath, err)
			}

			log.Printf("Found optimal release for %s Ch %g -> %s (Seeds: %d)", series.Title, ch.Number, bestTorrent.Title, bestTorrent.Seeders)

			err = e.Downloader.AddTorrentFromURL(bestTorrent.Link, targetPath)
			if err != nil {
				log.Printf("Failed to dispatch torrent to client: %v", err)
				continue
			}

			downloadedTorrents[bestTorrent.InfoHash] = true
			ch.Status = "Downloading"
			// _ = e.ChapterStore.Update(ch)
		} else {
			log.Printf("No available candidate on Nyaa matches %s Ch %g (Vol %v)", series.Title, ch.Number, ch.Volume)
		}
	}

	return nil
}
