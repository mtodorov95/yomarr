package indexer

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type NyaaResult struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Guid        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Seeders     int    `xml:"https://nyaa.si/xmlns/nyaa seeders"`
	Leechers    int    `xml:"https://nyaa.si/xmlns/nyaa leechers"`
	Downloads   int    `xml:"https://nyaa.si/xmlns/nyaa downloads"`
	InfoHash    string `xml:"https://nyaa.si/xmlns/nyaa infoHash"`
	Size        string `xml:"https://nyaa.si/xmlns/nyaa size"`
}

type nyaaRSS struct {
	XMLName xml.Name `xml:"rss"`
	Items   []NyaaResult `xml:"channel>item"`
}

type NyaaIndexer struct {
	Client *http.Client
}

func NewNyaaIndexer() *NyaaIndexer {
	return &NyaaIndexer{
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *NyaaIndexer) Search(query string) ([]NyaaResult, error) {
	apiURL := fmt.Sprintf("https://nyaa.si/?page=rss&q=%s&c=3_1&f=0", url.QueryEscape(query))

	resp, err := n.Client.Get(apiURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nyaa error code: %d", resp.StatusCode)
	}

	var rss nyaaRSS
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		return nil, err
	}

	return rss.Items, nil
}
