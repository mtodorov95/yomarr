package db

import (
	"testing"
	"time"
	"github.com/mtodorov95/yomarr/internal/models"
)

func TestChapterStore(t *testing.T) {
	testDB := Init(":memory:")
	defer testDB.Close()

	store := &SQLiteChapterStore{DB: testDB}
	
	t.Run("InsertAndGetBySeries", func(t *testing.T) {
		vol := 1
		path := "/test/ch1.5.cbz"
		now := time.Now()

		c := &models.Chapters{
			SeriesID:    1,
			Number:      1.5,
			Volume:      &vol,  
			FilePath:    &path,
			Status:      "Available",
			ReleaseDate: &now,
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
