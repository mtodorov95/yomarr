package sync

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/mtodorov95/yomarr/internal/models"
)

type MockStore struct {
	chapters []models.Chapters
	updates  []models.Chapters
	inserts  []models.Chapters
}

func (m *MockStore) GetBySeriesId(id int64) ([]models.Chapters, error) {
	return m.chapters, nil
}

func (m *MockStore) Update(ch *models.Chapters) error {
	m.updates = append(m.updates, *ch)
	return nil
}

func (m *MockStore) Insert(ch *models.Chapters) error {
	m.inserts = append(m.inserts, *ch)
	return nil
}

func createMockCBZ(t *testing.T, targetPath string, internalFiles []string) {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		t.Fatalf("Failed to create mock subdirs: %v", err)
	}

	f, err := os.Create(targetPath)
	if err != nil {
		t.Fatalf("Failed to create mock cbz file: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	for _, file := range internalFiles {
		_, err := zw.Create(file)
		if err != nil {
			t.Fatalf("Failed to write mock internal zip track: %v", err)
		}
	}
}

func TestExtractChaptersFromArchive(t *testing.T) {
	ls := &LibraryScanner{}

	tests := []struct {
		name          string
		archiveName   string
		internalFiles []string
		expectedChaps map[float64]int
	}{
		{
			name:        "Flat Digital Release Layout (1r0n style)",
			archiveName: "The Anemone Feels the Heat v02 (2025) (Digital).cbz",
			internalFiles: []string{
				"The Anemone Feels the Heat - c007 (v02) - p000 [Yen Press].jpg",
				"The Anemone Feels the Heat - c007 (v02) - p001 [Yen Press].jpg",
				"The Anemone Feels the Heat - c008 (v02) - p000 [Yen Press].jpg",
			},
			expectedChaps: map[float64]int{7: 2, 8: 2},
		},
		{
			name:        "MangaDex Chapter Folders Layout",
			archiveName: "Fuufu Ijou, Koibito Miman. v02.cbz",
			internalFiles: []string{
				"[MangaDex] Vol. 02 Ch. 009.5 - Special/001.png",
				"[MangaDex] Vol. 02 Ch. 009.5 - Special/002.png",
				"[MangaDex] Vol. 02 Ch. 010 [The Abandoned]/0187.jpg",
			},
			expectedChaps: map[float64]int{9.5: 2, 10: 2},
		},
		{
			name:        "Flat Raw Volume Pages Layout (Should be ignored)",
			archiveName: "Fuufu Ijou, Koibito Miman. v02_RAW_FLAT.cbz",
			internalFiles: []string{
				"0186.jpg",
				"0187.jpg",
				"0197.jpg",
			},
			expectedChaps: map[float64]int{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			archivePath := filepath.Join(tmpDir, tc.archiveName)

			createMockCBZ(t, archivePath, tc.internalFiles)

			res, err := extractChaptersFromArchiveWorkaround(ls, archivePath, 2)
			if err != nil {
				t.Fatalf("Unexpected archive parsing error: %v", err)
			}

			if len(res) != len(tc.expectedChaps) {
				t.Errorf("Extracted map size mismatch. Expected %d entries, got %d (%v)", len(tc.expectedChaps), len(res), res)
			}

			for ch, vol := range tc.expectedChaps {
				if gotVol, found := res[ch]; !found || gotVol != vol {
					t.Errorf("Expected chapter %g to point to Vol %d, got Vol %d (found: %t)", ch, vol, gotVol, found)
				}
			}
		})
	}
}

func extractChaptersFromArchiveWorkaround(ls *LibraryScanner, path string, fallbackVol int) (map[float64]int, error) {
	return ls.extractChaptersFromArchive(path, fallbackVol)
}
