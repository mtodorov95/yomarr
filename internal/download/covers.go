package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DownloadSeriesCovers(client *http.Client, seriesPath string, remoteThumbnail string, remoteHistoricals []string) (string, []string, error) {
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

	var localHistoricalsRel []string
	for _, remoteURL := range remoteHistoricals {
		if remoteURL == "" {
			continue
		}
		
		parts := strings.Split(remoteURL, "/")
		fileName := parts[len(parts)-1]
		if fileName == "" {
			continue
		}

		destPath := filepath.Join(coversDir, fileName)
		if err := downloadFile(client, remoteURL, destPath); err == nil {
			localHistoricalsRel = append(localHistoricalsRel, filepath.Join("Covers", fileName))
		}
	}

	return localThumbnailRel, localHistoricalsRel, nil
}

func downloadFile(client *http.Client, url string, dest string) error {
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
