package metadata

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtodorov95/yomarr/internal/models"
)

func TestMapMDStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected models.SeriesStatus
	}{
		{"ongoing", models.SeriesOngoing},
		{"hiatus", models.SeriesHiatus},
		{"completed", models.SeriesCompleted},
		{"cancelled", models.SeriesUnmonitored},
		{"unknown_garbage", models.SeriesOngoing},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			actual := mapMDStatus(tc.input)
			if actual != tc.expected {
				t.Errorf("For status %q: expected %v, got %v", tc.input, tc.expected, actual)
			}
		})
	}
}

func TestGetDetails(t *testing.T) {
	mockDetailsJSON := `{
		"data": {
			"id": "mock-manga-123",
			"attributes": {
				"title": { "en": "Chainsaw Man" },
				"altTitles": [ { "ja": "チェンソーマン" } ],
				"status": "hiatus",
				"year": 2018,
				"lastChapter": "145",
				"lastVolume": "15",
				"description": { "en": "A legendary dark fantasy manga series." },
				"links": { "al": "105778" },
				"tags": [
					{ "attributes": { "name": { "en": "Action" }, "group": "genre" } }
				]
			},
			"relationships": [
				{
					"id": "author-id",
					"type": "author",
					"attributes": { "name": "Tatsuki Fujimoto" }
				},
				{
					"id": "artist-id",
					"type": "artist",
					"attributes": { "name": "Tatsuki Fujimoto" }
				}
			]
		}
	}`

	mockCoversJSON := `{
		"data": [
			{
				"attributes": {
					"fileName": "cover_vol1.jpg",
					"volume": "1"
				}
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		if r.URL.Path == "/manga/mock-manga-123" {
			_, _ = w.Write([]byte(mockDetailsJSON))
			return
		}
		if r.URL.Path == "/cover" {
			_, _ = w.Write([]byte(mockCoversJSON))
			return
		}
		
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer server.Close()

	var res mdDetailsResponse
	if err := json.Unmarshal([]byte(mockDetailsJSON), &res); err != nil {
		t.Fatalf("Failed unmarshaling mock data fixture: %v", err)
	}

	var yearVal int
	if res.Data.Attributes.Year != nil {
		yearVal = int(*res.Data.Attributes.Year)
	}

	s := &models.Series{
		MangadexID:  &res.Data.ID,
		Title:       getMDTitle(res.Data.Attributes.Title, res.Data.Attributes.AltTitles),
		Status:      mapMDStatus(res.Data.Attributes.Status),
		Author:      getMDAuthor(res.Data.Relationships),
		Artist:      getMDArtist(res.Data.Relationships),
		Year:        &yearVal,
		LastChapter: res.Data.Attributes.LastChapter,
		LastVolume:  res.Data.Attributes.LastVolume,
	}

	if *s.MangadexID != "mock-manga-123" {
		t.Errorf("Expected ID mock-manga-123, got %s", *s.MangadexID)
	}
	if s.Title != "Chainsaw Man" {
		t.Errorf("Expected title Chainsaw Man, got %s", s.Title)
	}
	if s.Status != models.SeriesHiatus {
		t.Errorf("Expected status SeriesHiatus, got %v", s.Status)
	}
	if s.Author == nil || *s.Author != "Tatsuki Fujimoto" {
		t.Errorf("Expected author Tatsuki Fujimoto, got %v", s.Author)
	}
	if s.Artist == nil || *s.Artist != "Tatsuki Fujimoto" {
		t.Errorf("Expected artist Tatsuki Fujimoto, got %v", s.Artist)
	}
	if s.Year == nil || *s.Year != 2018 {
		t.Errorf("Expected year 2018, got %v", s.Year)
	}
	if s.LastChapter == nil || *s.LastChapter != "145" {
		t.Errorf("Expected last chapter 145, got %v", s.LastChapter)
	}
	if s.LastVolume == nil || *s.LastVolume != "15" {
		t.Errorf("Expected last volume 15, got %v", s.LastVolume)
	}
}
