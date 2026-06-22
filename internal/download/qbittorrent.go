package download

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mtodorov95/yomarr/internal/indexer"
	"github.com/mtodorov95/yomarr/internal/models"
)

type qbitTorrentResponse struct {
	Hash     string  `json:"hash"`
	Name     string  `json:"name"`
	Progress float64 `json:"progress"`
	Tags     string  `json:"tags"`
}

type QBittorrentClient struct {
	BaseURL string
	Client  *http.Client
	Cfg     models.DownloadClient
}

func NewQBittorrentClient(cfg models.DownloadClient) (*QBittorrentClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	scheme := "http"
	if cfg.UseSSL {
		scheme = "https"
	}

	cleanHost := strings.TrimPrefix(cfg.Host, "http://")
	cleanHost = strings.TrimPrefix(cleanHost, "https://")
	cleanHost = strings.TrimSuffix(cleanHost, "/")

	computedURL := fmt.Sprintf("%s://%s:%d", scheme, cleanHost, cfg.Port)

	qb := &QBittorrentClient{
		BaseURL:  computedURL,
		Client: &http.Client{
			Timeout: 10 * time.Second,
			Jar:     jar,
		},
	}

	if err := qb.login(); err != nil {
		return nil, fmt.Errorf("qbittorrent auth failed: %w", err)
	}

	return qb, nil
}

func (q *QBittorrentClient) login() error {
	loginURL := fmt.Sprintf("%s/api/v2/auth/login", q.BaseURL)
	data := url.Values{}
	data.Set("username", q.Cfg.Username)
	data.Set("password", q.Cfg.Password)

	resp, err := q.Client.PostForm(loginURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (q *QBittorrentClient) AddTorrentFromURL(torrentURL string, savePath string, seedTime int, language string, seriesID int64, release indexer.ParsedRelease) (string, error) {
	addURL := fmt.Sprintf("%s/api/v2/torrents/add", q.BaseURL)

	categoryTarget := q.Cfg.Category
	if categoryTarget == "" {
		categoryTarget = "yomarr"
	}

	data := url.Values{}
	data.Set("urls", torrentURL)
	data.Set("savepath", savePath)
	data.Set("category", categoryTarget)
	data.Set("ratioLimit", "-2")
	data.Set("seedingTimeLimit", strconv.Itoa(seedTime))

	resp, err := q.Client.PostForm(addURL, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to add torrent, code: %d", resp.StatusCode)
	}

	infoHash := ""
	if strings.HasPrefix(torrentURL, "magnet:") {
		if u, err := url.Parse(torrentURL); err == nil {
			xt := u.Query().Get("xt")
			if strings.HasPrefix(xt, "urn:btih:") {
				infoHash = strings.ToLower(strings.TrimPrefix(xt, "urn:btih:"))
			}
		}
	}

	if infoHash == "" {
		time.Sleep(150 * time.Millisecond)

		downloads, err := q.GetActiveDownloads()
		if err == nil && len(downloads) > 0 {
			infoHash = downloads[len(downloads)-1].Hash
		}
	}

	return infoHash, nil
}

func (q *QBittorrentClient) AddTorrentFromMagnet(magnet string, savePath string, seedTime int, language string, seriesID int64, release indexer.ParsedRelease) (string, error) {
	return q.AddTorrentFromURL(magnet, savePath, seedTime, language, seriesID, release)
}

func (q *QBittorrentClient) GetActiveDownloads() ([]TorrentInfo, error) {
	categoryTarget := q.Cfg.Category
	if categoryTarget == "" {
		categoryTarget = "yomarr"
	}

	infoURL := fmt.Sprintf("%s/api/v2/torrents/info?category=%s", q.BaseURL, categoryTarget)
	resp, err := q.Client.Get(infoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qbittorrent returned status code %d", resp.StatusCode)
	}

	var rawTorrents []qbitTorrentResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawTorrents); err != nil {
		return nil, err
	}

	results := make([]TorrentInfo, len(rawTorrents))
	for i, t := range rawTorrents {
		var tagsSlice []string
		if t.Tags != "" {
			for _, tag := range strings.Split(t.Tags, ",") {
				tagsSlice = append(tagsSlice, strings.TrimSpace(tag))
			}
		}

		results[i] = TorrentInfo{
			Hash:     t.Hash,
			Name:     t.Name,
			Progress: t.Progress,
			Tags:     tagsSlice,
		}
	}

	return results, nil
}

func (q *QBittorrentClient) MarkAsImported(hash string) error {
	return nil
}
