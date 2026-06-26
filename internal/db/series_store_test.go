package db

import (
	"testing"

	"github.com/mtodorov95/yomarr/internal/models"
)

func TestSeriesStore(t *testing.T) {
	testDB := Init(":memory:")
	defer testDB.Close()

	store := &SQLiteSeriesStore{DB: testDB}

	t.Run("InsertAndGet", func(t *testing.T) {
		s := &models.Series{
			Title: "Test Manga",
			AnilistID: ToPtr("231"),
			MangadexID: ToPtr("abc"),
			Path: "/test/path",
			Status: "Monitored",
		}

		err := store.Insert(s)
		if err != nil {
			t.Fatalf("Insert fail: %v", err)
		}

		if s.ID == 0 {
			t.Error("ID not set after insert")
		}

		got, err := store.GetById(s.ID)
		if err != nil {
			t.Fatalf("GetById fail: %v", err)
		}
		if got.Title != s.Title {
			t.Errorf("Title mismatch: got %s, want %s", got.Title, s.Title)
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		list, err := store.GetAll()
		if err != nil {
			t.Fatalf("GetAll fail: %v", err)
		}
		if len(list) == 0 {
			t.Error("List empty, want 1 item")
		}
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		got, err := store.GetById(999)
		if err != nil {
			t.Fatalf("Error on missing: %v", err)
		}
		if got != nil {
			t.Error("Want nil for missing ID")
		}
	})
}
