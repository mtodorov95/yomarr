package sync

import (
	"log"
	"time"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/models"
)

type RssEngine struct {
	ChapterStore db.ChapterStore
	SeriesStore  db.SeriesStore
	Manager      *DynamicManager
}

func NewRssEngine(cs db.ChapterStore, ss db.SeriesStore, dm *DynamicManager) *RssEngine {
	return &RssEngine{
		ChapterStore: cs,
		SeriesStore:  ss,
		Manager:      dm,
	}
}

func (e *RssEngine) StartBackgroundRssCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		log.Printf("[RSS] Background RSS engine starting. Interval: %v", interval)
		for range ticker.C {
			monitoredSeries, err := e.SeriesStore.GetAllMonitored()
			if err != nil || len(monitoredSeries) == 0 {
				continue
			}

			latestReleases, err := e.Manager.FetchLatestRSS()
			if err != nil {
				log.Printf("[RSS Error] Failed to fetch stream: %v", err)
				continue
			}

			for _, release := range latestReleases {
				for _, series := range monitoredSeries {
					missingChapters, _ := e.ChapterStore.GetMissingBySeriesID(series.ID)
					if len(missingChapters) == 0 {
						continue
					}

					for _, ch := range missingChapters {
						parsedRelease, matched := IsChapterMatch(release, ch, &series)
						if matched {
							log.Printf("[RSS] Found : %s Ch %g [%s]", series.Title, ch.Number, ch.Language)

							_, err := e.Manager.AddTorrentFromURL(
								release.Link,
								getDownloadsPath(),
								release.SeedTime,
								release.Language,
								series.ID,
								parsedRelease,
							)
							if err != nil {
								log.Printf("[RSS] Failed to pass torrent to downloader client: %v", err)
								continue
							}

							ch.Status = models.ChapterDownloading
							_ = e.ChapterStore.Update(ch)
							break
						}
					}
				}
			}
		}
	}()
}
