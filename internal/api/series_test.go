package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtodorov95/yomarr/internal/models"
)

type MockStore struct {
	series []models.Series
}

func (m *MockStore) GetAll() ([]models.Series, error) { return m.series, nil }
func (m *MockStore) GetById(id int64) (*models.Series, error) {
	for _, s := range m.series {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, nil
}
func (m *MockStore) Insert(s *models.Series) error {
	s.ID = int64(len(m.series) + 1)
	m.series = append(m.series, *s)
	return nil
}

func TestHandleSeries(t *testing.T) {
	mock := &MockStore{}
	handler := &SeriesHandler{Store: mock}

	t.Run("POST create series", func(t *testing.T) {
		s := models.Series{Title: "Test"}
		body, _ := json.Marshal(s)
		req := httptest.NewRequest("POST", "/api/series", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.HandleSeries(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("Status mismatch: got %d", rr.Code)
		}
	})

	t.Run("GET returns list", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/series", nil)
		rr := httptest.NewRecorder()

		handler.HandleSeries(rr, req)

		var list []models.Series
		json.NewDecoder(rr.Body).Decode(&list)
		if len(list) != 1 {
			t.Errorf("List size mismatch: got %d", len(list))
		}
	})
}
