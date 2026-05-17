package db

import (
	"testing"
	"time"
	"github.com/mtodorov95/yomarr/internal/models"
)

func TestChapterStore(t *testing.T) {
	Init(":memory:")
	defer DB.Close()

	store := &SQLiteChapterStore{}
	
	t.Run("InsertAndGetBySeries", func(t *testing.T) {
		c := &models.Chapters{
			SeriesID:    1,
			Number:      1.5,
			Volume:      1,
			FilePath:    "/test/ch1.5.cbz",
			Status:      "Available",
			ReleaseDate: time.Now(),
		}

		if err := store.Insert(c); err != nil {
			t.Fatalf("Insert fail: %v", err)
		}

		list, err := store.GetBySeriesId(1)
		if err != nil {
			t.Fatalf("Get fail: %v", err)
		}
		if len(list) != 1 || list[0].Number != 1.5 {
			t.Errorf("Data mismatch. Got number: %f", list[0].Number)
		}
	})
}
