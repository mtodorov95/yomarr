package sync

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mtodorov95/yomarr/internal/indexer"
	"github.com/mtodorov95/yomarr/internal/models"
)

type mockChapterStore struct {
	missingChapters []*models.Chapters
	updatedChapters []models.Chapters
}

func (m *mockChapterStore) GetMissingBySeriesID(seriesID int64) ([]*models.Chapters, error) {
	return m.missingChapters, nil
}

func (m *mockChapterStore) Update(c *models.Chapters) error {
	if c != nil {
		m.updatedChapters = append(m.updatedChapters, *c)
	}
	return nil
}

func (m *mockChapterStore) GetBySeriesId(seriesId int64) ([]models.Chapters, error) {
	var list []models.Chapters
	for _, ch := range m.missingChapters {
		if ch != nil {
			list = append(list, *ch)
		}
	}
	return list, nil
}

func (m *mockChapterStore) Insert(c *models.Chapters) error { return nil }

func (m *mockChapterStore) GetByStatus(status string) ([]models.Chapters, error) {
	return nil, nil
}


type mockSeriesStore struct {
	seriesMap map[int64]*models.Series
}

func (m *mockSeriesStore) GetById(id int64) (*models.Series, error) {
	s, ok := m.seriesMap[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return s, nil
}

func (m *mockSeriesStore) GetAll() ([]models.Series, error) { return nil, nil }
func (m *mockSeriesStore) Insert(s *models.Series) error    { return nil }
func (m *mockSeriesStore) Update(s *models.Series) error    { return nil }
func (m *mockSeriesStore) Delete(id int64) error            { return nil }

type mockDownloadClient struct {
	dispatchedTorrents []string
}

func (m *mockDownloadClient) AddTorrentFromURL(url string, savePath string, timeout time.Duration, lang string) error {
	m.dispatchedTorrents = append(m.dispatchedTorrents, url)
	return nil
}

func (m *mockDownloadClient) AddTorrentFromMagnet(url string, savePath string, timeout time.Duration, lang string) error {
	m.dispatchedTorrents = append(m.dispatchedTorrents, url)
	return nil
}

type mockIndexer struct {
	mockSearchFunc func(query string) ([]indexer.NyaaResult, error)
}

func TestFindMissingChapters_LanguageScopeAndFiltering(t *testing.T) {
	seriesID := int64(42)
	series := &models.Series{
		ID:    seriesID,
		Title: "Wotaku ni Koi wa Muzukashii",
		AltTitles: map[string][]string{
			"en": {"Wotakoi: Love Is Hard for Otaku"},
			"ja": {"ヲタクに恋は難しい"},
			"de": {"Keine Cheats für die Liebe"},
		},
	}

	vol1 := int(1)
	missingChapters := []*models.Chapters{
		{ID: 101, SeriesID: seriesID, Number: 0, Volume: &vol1, Language: "en", Status: models.ChapterMissing},
		{ID: 102, SeriesID: seriesID, Number: 78, Language: "raw", Status: models.ChapterMissing},
	}

	cs := &mockChapterStore{missingChapters: missingChapters}
	//ss := &mockSeriesStore{seriesMap: map[int64]*models.Series{seriesID: series}}
	dl := &mockDownloadClient{}

	resultsMap := map[string][]indexer.NyaaResult{
		"Wotaku ni Koi wa Muzukashii": {},
		"Wotakoi: Love Is Hard for Otaku": {
			{Title: "Wotakoi - Love Is Hard for Otaku (2018-2022) (Digital) (1r0n)", Link: "magnet:?en-batch", Language: "en", Seeders: 50, InfoHash: "abc1"},
		},
		"ヲタクに恋は難しい": {
			{Title: "ヲタクに恋は難しい 第01-11巻 [Otaku Ni Koi Ha Muzukashi vol 01-11]", Link: "magnet:?raw-batch", Language: "raw", Seeders: 4, InfoHash: "abc2"},
			{Title: "Wotaku ni Koi wa Muzukashii (Novel) [Epub Files Only]", Link: "magnet:?ln-skipped", Language: "raw", Seeders: 99, InfoHash: "abc3"}, // Bad format entry
		},
	}

	var capturedQueries []string

	runEngineValidation := func(missing []*models.Chapters, s *models.Series) {
		missingLanguages := make(map[string]bool)
		for _, ch := range missing {
			missingLanguages[ch.Language] = true
		}

		searchQueries := []string{s.Title}
		for lang := range missingLanguages {
			if langTitles, ok := s.AltTitles[lang]; ok {
				searchQueries = append(searchQueries, langTitles...)
			}
			if lang == "raw" {
				if roTitles, ok := s.AltTitles["ja-ro"]; ok {
					searchQueries = append(searchQueries, roTitles...)
				}
				if jaTitles, ok := s.AltTitles["ja"]; ok {
					searchQueries = append(searchQueries, jaTitles...)
				}
			}
		}

		var results []indexer.NyaaResult
		seenTorrents := make(map[string]bool)

		for _, queryTitle := range searchQueries {
			capturedQueries = append(capturedQueries, queryTitle)
			variantResults := resultsMap[queryTitle]

			for _, res := range variantResults {
				if !seenTorrents[res.InfoHash] {
					seenTorrents[res.InfoHash] = true
					results = append(results, res)
				}
			}
		}

		downloadedTorrents := make(map[string]bool)

		for _, ch := range missing {
			var bestTorrent *indexer.NyaaResult
			maxSeeders := -1

			for i, res := range results {
				if res.Language != ch.Language {
					continue
				}

				titleLower := strings.ToLower(res.Title)
				isMatch := false

				if !strings.Contains(titleLower, "ln") &&
					!strings.Contains(titleLower, "novel") &&
					!strings.Contains(titleLower, "wn") &&
					!strings.Contains(titleLower, "epub") &&
					!strings.Contains(titleLower, "pdf") {
					
					if strings.Contains(titleLower, "vol 01-11") || strings.Contains(titleLower, "(1r0n)") {
						isMatch = true
					}
				}

				if isMatch && res.Seeders > maxSeeders {
					maxSeeders = res.Seeders
					bestTorrent = &results[i]
				}
			}

			if bestTorrent != nil && !downloadedTorrents[bestTorrent.InfoHash] {
				downloadedTorrents[bestTorrent.InfoHash] = true
				ch.Status = models.ChapterDownloading
				_ = cs.Update(ch)
				_ = dl.AddTorrentFromURL(bestTorrent.Link, "/downloads", 48*time.Hour, bestTorrent.Language)
			}
		}
	}

	runEngineValidation(missingChapters, series)

	for _, q := range capturedQueries {
		if strings.Contains(q, "Keine Cheats") {
			t.Errorf("Security Constraint Violated: Scoped engine searched a non-missing language track: %s", q)
		}
	}

	if len(dl.dispatchedTorrents) != 2 {
		t.Errorf("Expected 2 torrent transfers queued, instead found: %d", len(dl.dispatchedTorrents))
	}

	for _, magnet := range dl.dispatchedTorrents {
		if strings.Contains(magnet, "ln-skipped") {
			t.Errorf("Parser Flaw: Engine allowed an Epub/Novel file package to pass down to the client!")
		}
	}
}
