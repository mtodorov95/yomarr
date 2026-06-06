package indexer

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type NyaaResult struct {
	Title     string `xml:"title"`
	Link      string `xml:"link"`
	Guid      string `xml:"guid"`
	PubDate   string `xml:"pubDate"`
	Seeders   int    `xml:"https://nyaa.si/xmlns/nyaa seeders"`
	Leechers  int    `xml:"https://nyaa.si/xmlns/nyaa leechers"`
	Downloads int    `xml:"https://nyaa.si/xmlns/nyaa downloads"`
	InfoHash  string `xml:"https://nyaa.si/xmlns/nyaa infoHash"`
	Size      string `xml:"https://nyaa.si/xmlns/nyaa size"`
	Language  string `json:"language"`
}

type nyaaRSS struct {
	XMLName xml.Name     `xml:"rss"`
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
	categories := []struct {
		code string
		lang string
	}{
		{code: "3_1", lang: "en"},
		{code: "3_3", lang: "raw"},
	}

	var wg sync.WaitGroup
	resultChan := make(chan []NyaaResult, len(categories))
	errChan := make(chan error, len(categories))

	for _, cat := range categories {
		wg.Add(1)
		go func(categoryCode, langTag string) {
			defer wg.Done()

			apiURL := fmt.Sprintf("https://nyaa.si/?page=rss&q=%s&c=%s&f=0", url.QueryEscape(query), categoryCode)

			resp, err := n.Client.Get(apiURL)
			if err != nil {
				errChan <- err
				return 
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errChan <- fmt.Errorf("nyaa error code: %d", resp.StatusCode)
				return
			}

			var rss nyaaRSS
			if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
				errChan <- err
				return 
			}

			for i := range rss.Items {
				rss.Items[i].Language = langTag
			}

			resultChan <- rss.Items
		}(cat.code, cat.lang)
	}

	wg.Wait()
	close(resultChan)
	close(errChan)

	if len(errChan) > 0 {
		return nil, <-errChan
	}

	var combinedResults []NyaaResult
	for items := range resultChan {
		combinedResults = append(combinedResults, items...)
	}

	return combinedResults, nil
}
