package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mtodorov95/yomarr/internal/models"
)

func DownloadSeriesCovers(client *http.Client, seriesPath string, remoteThumbnail string, remoteHistoricals []models.VolumeCover) (string, []models.VolumeCover, error) {
	if seriesPath == "" {
		return "", nil, fmt.Errorf("cannot download covers: series path is empty")
	}

	coversDir := filepath.Join(seriesPath, "Covers")
	if err := os.MkdirAll(coversDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed creating covers directory: %w", err)
	}

	localThumbnailRel := ""
	if remoteThumbnail != "" {
		parts := strings.Split(remoteThumbnail, "/")
		fileName := parts[len(parts)-1]

		if fileName != "" {
			destPath := filepath.Join(coversDir, fileName)
			if err := downloadFile(client, remoteThumbnail, destPath); err == nil {
				localThumbnailRel = filepath.Join("Covers", fileName)
			}
		}
	}

	var localHistoricalsRel []models.VolumeCover
	for _, remoteCover := range remoteHistoricals {
		if remoteCover.URL == "" {
			continue
		}

		parts := strings.Split(remoteCover.URL, "/")
		fileName := parts[len(parts)-1]
		if fileName == "" {
			continue
		}

		destPath := filepath.Join(coversDir, fileName)
		if err := downloadFile(client, remoteCover.URL, destPath); err == nil {
			localHistoricalsRel = append(localHistoricalsRel, models.VolumeCover{
				Volume: remoteCover.Volume,
				URL:    filepath.Join("Covers", fileName),
			})
		}
	}

	return localThumbnailRel, localHistoricalsRel, nil
}

func downloadFile(client *http.Client, url string, dest string) error {
	if info, err := os.Stat(dest); err == nil && info.Size() > 0 {
		return nil
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Yomarr/1.0.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status downloading asset: %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
