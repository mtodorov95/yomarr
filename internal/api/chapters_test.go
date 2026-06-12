package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/mtodorov95/yomarr/internal/models"
)

type MockChapterStore struct{}

func (m *MockChapterStore) GetBySeriesId(id int64) ([]models.Chapters, error) {
	return []models.Chapters{{ID: 1, SeriesID: id}}, nil
}
func (m *MockChapterStore) Insert(c *models.Chapters) error { return nil }

func (m *MockChapterStore) GetByStatus(status string) ([]models.Chapters, error) {
	return []models.Chapters{}, nil
}

func (m *MockChapterStore) Update(c *models.Chapters) error { return nil }

func (m *MockChapterStore) GetMissingBySeriesID(seriesID int64) ([]*models.Chapters, error) {
	return []*models.Chapters{}, nil
}

func TestHandleChapters(t *testing.T) {
	handler := &ChapterHandler{Store: &MockChapterStore{}}

	t.Run("GET requires series_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/chapters", nil)
		rr := httptest.NewRecorder()
		handler.HandleChapters(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
	})

	t.Run("GET with series_id works", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/chapters?series_id=1", nil)
		rr := httptest.NewRecorder()
		handler.HandleChapters(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", rr.Code)
		}
	})
}
