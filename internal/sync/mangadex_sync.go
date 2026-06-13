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

	existingMap := make(map[string]*models.Chapters)
	for i := range existing {
		ch := &existing[i]
		key := fmt.Sprintf("%f-%s", ch.Number, ch.Language)
		existingMap[key] = ch
	}

	type chapterMeta struct {
		Volume *int
	}
	globalChapterNumbers := make(map[float64]chapterMeta)

	for _, rCh := range remoteChapters {
		var volPtr *int
		if rCh.Volume != nil && *rCh.Volume != "" {
			if v, err := strconv.Atoi(*rCh.Volume); err == nil {
				volPtr = &v
			}
		}

		meta, exists := globalChapterNumbers[rCh.Number]
		if !exists || (meta.Volume == nil && volPtr != nil) {
			globalChapterNumbers[rCh.Number] = chapterMeta{Volume: volPtr}
		}
	}

	var insertedCount int
	var updatedCount int

	for chNum, meta := range globalChapterNumbers {
		enKey := fmt.Sprintf("%f-en", chNum)
		enCh, exists := existingMap[enKey]
		if !exists {
			ch, err := e.insertMissingChapter(seriesID, chNum, meta.Volume, "en")
			if err != nil {
				return err
			}

			existingMap[enKey] = ch
			insertedCount++
		} else {
			if enCh.Volume == nil && meta.Volume != nil {
				enCh.Volume = meta.Volume
				if err := e.ChapterStore.Update(enCh); err != nil {
					return fmt.Errorf("failed backfilling english chapter volume: %w", err)
				}
				updatedCount++
			}
		}

		rawKey := fmt.Sprintf("%f-raw", chNum)
		rawCh, exists := existingMap[rawKey]
		if !exists {
			ch, err := e.insertMissingChapter(seriesID, chNum, meta.Volume, "raw")
			if err != nil {
				return fmt.Errorf("failed to insert parallel raw variant: %w", err)
			}
			existingMap[rawKey] = ch
			insertedCount++
		} else {
			if rawCh.Volume == nil {
				rawCh.Volume = meta.Volume
				if err := e.ChapterStore.Update(rawCh); err != nil {
					return fmt.Errorf("failed backfilling raw chapter volume: %w", err)
				}
				updatedCount++
			}
		}
	}

	log.Printf("Sync complete. Added %d new missing chapters for series %d", insertedCount, seriesID)
	return nil
}

func (e *MangaDexSyncEngine) insertMissingChapter(seriesID int64, chNum float64, volume *int, lang string) (*models.Chapters, error) {
	chapter := &models.Chapters{
		SeriesID: seriesID,
		Number:   chNum,
		Volume:   volume,
		Status:   models.ChapterMissing,
		FilePath: nil,
		Language: lang,
	}

	if err := e.ChapterStore.Insert(chapter); err != nil {
		return nil, err
	}

	return chapter, nil
}
