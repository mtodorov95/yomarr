package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mtodorov95/yomarr/internal/models"
)

func EnsureSeriesDirectory(basePath string, s *models.Series) (string, error) {
	if basePath == "" {
		return "", fmt.Errorf("base library path cannot be empty")
	}

	cleanTitle := strings.TrimSpace(s.Title)
	
	folderName := cleanTitle
	if s.Year != nil && *s.Year > 0 {
		folderName = fmt.Sprintf("%s (%d)", cleanTitle, *s.Year)
	}

	fullPath := filepath.Join(basePath, folderName)

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return "", fmt.Errorf("failed creating series base directory: %w", err)
	}

	return fullPath, nil
}

func EnsureLanguageDirectory(seriesPath string, language string) (string, error) {
	if seriesPath == "" {
		return "", fmt.Errorf("series destination path cannot be empty")
	}

	langDir := "EN"
	if strings.ToLower(language) == "raw" {
		langDir = "RAW"
	}

	destDir := filepath.Join(seriesPath, langDir)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed creating localized library path: %w", err)
	}

	return destDir, nil
}

func DeleteSeriesFile(seriesPath string, targetFile string) error {
	if seriesPath == "" || targetFile == "" {
		return fmt.Errorf("invalid paths specified for file deletion")
	}

	fullPath := filepath.Join(seriesPath, targetFile)
	
	return os.Remove(fullPath)
}
