package sync

import (
	"fmt"
	"log"
	"strconv"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/metadata"
	"github.com/mtodorov95/yomarr/internal/models"
)

type MangaDexSyncEngine struct {
	ChapterStore db.ChapterStore
	MDProvider   *metadata.MangaDexProvider
}

func NewMangaDexSyncEngine(cs db.ChapterStore, mdp *metadata.MangaDexProvider) *MangaDexSyncEngine {
	return &MangaDexSyncEngine{ChapterStore: cs, MDProvider: mdp}
}

func (e *MangaDexSyncEngine) SyncChapters(seriesID int64, mangadexID string) error {
	if mangadexID == "" {
		return fmt.Errorf("cannot sync: series has no MangaDex ID mapped")
	}

	remoteChapters, err := e.MDProvider.GetChapterFeed(mangadexID)
	if err != nil {
		return fmt.Errorf("metadata fetch failed: %w", err)
	}

	existing, err := e.ChapterStore.GetBySeriesId(seriesID)
	if err != nil {
		return fmt.Errorf("failed fetching local cache: %w", err)
	}

	existingMap := make(map[string]bool)
	for _, ch := range existing {
		key := fmt.Sprintf("%f-%s", ch.Number, ch.Language)
		existingMap[key] = true
	}

	var insertedCount int
	for _, rCh := range remoteChapters {
		remoteKey := fmt.Sprintf("%f-%s", rCh.Number, rCh.Language)

		if !existingMap[remoteKey] {
			var volPtr *int
			if rCh.Volume != nil && *rCh.Volume != "" {
				if v, err := strconv.Atoi(*rCh.Volume); err == nil {
					volPtr = &v
				}
			}

			newChapter := &models.Chapters{
				SeriesID: seriesID,
				Number:   rCh.Number,
				Volume:   volPtr,
				Status:   "Missing",
				FilePath: nil,
				Language: rCh.Language,
			}

			if err := e.ChapterStore.Insert(newChapter); err != nil {
				return err
			}
			existingMap[remoteKey] = true
			insertedCount++
		}
	}

	log.Printf("Sync complete. Added %d new missing chapters for series %d", insertedCount, seriesID)
	return nil
}
