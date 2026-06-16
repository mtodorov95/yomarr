package sync

import (
	"log"
	"strings"
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
				releaseTitleLower := strings.ToLower(release.Title)

				for _, series := range monitoredSeries {

					if !TorrentMatchesSeries(releaseTitleLower, series) {
						continue
					}

					missingChapters, _ := e.ChapterStore.GetMissingBySeriesID(series.ID)

					for _, ch := range missingChapters {
						if ch.Language != release.Language {
							continue
						}
						if IsChapterMatch(release, ch) {
							log.Printf("[RSS] Found : %s Ch %g [%s]", series.Title, ch.Number, ch.Language)

							err := e.Manager.AddTorrentFromURL(release.Link, getDownloadsPath(), release.SeedTime, release.Language)
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
