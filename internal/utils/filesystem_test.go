package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mtodorov95/yomarr/internal/models"
)

func TestEnsureSeriesDirectory(t *testing.T) {
	year2019 := 2019
	zeroYear := 0

	tests := []struct {
		name         string
		basePathFunc func(string) string
		series       models.Series
		expectedSub  string
		expectError  bool
	}{
		{
			name: "Standard Title with Valid Year",
			basePathFunc: func(tmp string) string { return tmp },
			series: models.Series{
				Title: "Shingeki no Eroko-san",
				Year:  &year2019,
			},
			expectedSub: "Shingeki no Eroko-san (2019)",
			expectError: false,
		},
		{
			name: "Fallback handling with Missing Year",
			basePathFunc: func(tmp string) string { return tmp },
			series: models.Series{
				Title: "Chainsaw Man",
				Year:  nil,
			},
			expectedSub: "Chainsaw Man",
			expectError: false,
		},
		{
			name: "Fallback handling with Zero Year",
			basePathFunc: func(tmp string) string { return tmp },
			series: models.Series{
				Title: "Berserk",
				Year:  &zeroYear,
			},
			expectedSub: "Berserk",
			expectError: false,
		},
		{
			name: "Sanitization of Stray Whitespace",
			basePathFunc: func(tmp string) string { return tmp },
			series: models.Series{
				Title: "   Monster   ",
				Year:  &year2019,
			},
			expectedSub: "Monster (2019)",
			expectError: false,
		},
		{
			name: "Empty Base Path Error Enforcement",
			basePathFunc: func(tmp string) string { return "" },
			series: models.Series{
				Title: "One Piece",
			},
			expectedSub: "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			basePath := tc.basePathFunc(tmpDir)

			resultPath, err := EnsureSeriesDirectory(basePath, &tc.series)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but function returned success")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error encountered: %v", err)
			}

			expectedPath := filepath.Join(basePath, tc.expectedSub)
			if resultPath != expectedPath {
				t.Errorf("Path mismatch: expected %q, got %q", expectedPath, resultPath)
			}

			info, err := os.Stat(resultPath)
			if err != nil {
				t.Errorf("Directory was not actually created on disk: %v", err)
			} else if !info.IsDir() {
				t.Errorf("Created path is not a valid directory")
			}
		})
	}
}

func TestEnsureLanguageDirectory(t *testing.T) {
	tests := []struct {
		name        string
		language    string
		expectedSub string
	}{
		{
			name:        "Default English Mapping",
			language:    "en",
			expectedSub: "EN",
		},
		{
			name:        "Case Insensitive English Mapping",
			language:    "En_uS",
			expectedSub: "EN",
		},
		{
			name:        "Raw Japanese Mapping",
			language:    "raw",
			expectedSub: "RAW",
		},
		{
			name:        "Uppercase Raw Mapping",
			language:    "RAW",
			expectedSub: "RAW",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			seriesDir := t.TempDir()

			resultPath, err := EnsureLanguageDirectory(seriesDir, tc.language)
			if err != nil {
				t.Fatalf("Unexpected directory creation error: %v", err)
			}

			expectedPath := filepath.Join(seriesDir, tc.expectedSub)
			if resultPath != expectedPath {
				t.Errorf("Expected path %q, got %q", expectedPath, resultPath)
			}

			info, err := os.Stat(resultPath)
			if err != nil {
				t.Errorf("Language folder missing from disk: %v", err)
			} else if !info.IsDir() {
				t.Errorf("Target is not a directory")
			}
		})
	}

	t.Run("Empty Parent Path Rejection", func(t *testing.T) {
		_, err := EnsureLanguageDirectory("", "en")
		if err == nil {
			t.Error("Expected error on blank parent input path, got nil")
		}
	})
}

func TestDeleteSeriesFile(t *testing.T) {
	t.Run("Successful File Removal", func(t *testing.T) {
		seriesDir := t.TempDir()
		fileName := "cover.jpg"
		fullFilePath := filepath.Join(seriesDir, fileName)

		if err := os.WriteFile(fullFilePath, []byte("fake-image-bytes"), 0644); err != nil {
			t.Fatalf("Failed to initialize test file fixture: %v", err)
		}

		err := DeleteSeriesFile(seriesDir, fileName)
		if err != nil {
			t.Errorf("Unexpected deletion error: %v", err)
		}

		if _, err := os.Stat(fullFilePath); !os.IsNotExist(err) {
			t.Errorf("File should have been removed, but it still exists")
		}
	})

	t.Run("Missing File Propagation Error", func(t *testing.T) {
		seriesDir := t.TempDir()
		err := DeleteSeriesFile(seriesDir, "non_existent.png")
		if err == nil {
			t.Error("Expected an os.ErrNotExist error path, but got no error back")
		}
	})

	t.Run("Empty Path Arguments Validation Rejection", func(t *testing.T) {
		err := DeleteSeriesFile("", "cover.jpg")
		if err == nil {
			t.Error("Expected error for empty path parameter, got nil")
		}
	})
}
