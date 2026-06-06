package download

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type QBittorrentClient struct {
	BaseURL  string
	Username string
	Password string
	Client   *http.Client
}

func NewQBittorrentClient(baseURL, username, password string) (*QBittorrentClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	qb := &QBittorrentClient{
		BaseURL:  strings.TrimSuffix(baseURL, "/"),
		Username: username,
		Password: password,
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
	data.Set("username", q.Username)
	data.Set("password", q.Password)

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

func (q *QBittorrentClient) AddTorrentFromURL(torrentURL string, savePath string, seedDuration time.Duration, language string) error {
	addURL := fmt.Sprintf("%s/api/v2/torrents/add", q.BaseURL)
		
	minutesToSeed := int(seedDuration.Minutes())

	langTag := strings.ToLower(language)
	if langTag == "" {
		langTag = "en"
	}

	data := url.Values{}
	data.Set("urls", torrentURL)
	data.Set("savepath", savePath)
	data.Set("category", "yomarr")
	data.Set("tags", langTag)
	data.Set("ratioLimit", "-2")
	data.Set("seedingTimeLimit", strconv.Itoa(minutesToSeed))

	resp, err := q.Client.PostForm(addURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add torrent, code: %d", resp.StatusCode)
	}

	return nil
}

func (q *QBittorrentClient) AddTorrentFromMagnet(magnet string, savePath string, seedDuration time.Duration, language string) error {
	return q.AddTorrentFromURL(magnet, savePath, seedDuration, language) 
}
