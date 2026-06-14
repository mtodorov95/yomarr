package sync

import (
	"fmt"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/download"
	"github.com/mtodorov95/yomarr/internal/indexer"
)

type DynamicManager struct {
	IndexerStore db.IndexerStore
	ClientStore  db.DownloadClientStore
}

var _ indexer.Indexer = (*DynamicManager)(nil)
var _ download.DownloadClient = (*DynamicManager)(nil)

func NewDynamicManager(idxStore db.IndexerStore, dcStore db.DownloadClientStore) *DynamicManager {
	return &DynamicManager{
		IndexerStore: idxStore,
		ClientStore:  dcStore,
	}
}

func (m *DynamicManager) getPrimaryClient() (download.DownloadClient, error) {
	clients, err := m.ClientStore.GetAll()
	if err != nil || len(clients) == 0 {
		return nil, fmt.Errorf("no download clients configured in database")
	}

	config := clients[0]

	// Check config.Name in the future 
	return download.NewQBittorrentClient(config)
}

func (m *DynamicManager) AddTorrentFromURL(url string, savePath string, seedTime int, language string) error {
	client, err := m.getPrimaryClient()
	if err != nil {
		return err
	}
	return client.AddTorrentFromURL(url, savePath, seedTime, language)
}

func (m *DynamicManager) AddTorrentFromMagnet(magnet string, savePath string, seedTime int, language string) error {
	client, err := m.getPrimaryClient()
	if err != nil {
		return err
	}
	return client.AddTorrentFromMagnet(magnet, savePath, seedTime, language)
}

func (m *DynamicManager) GetActiveDownloads() ([]download.TorrentInfo, error) {
	client, err := m.getPrimaryClient()
	if err != nil {
		return nil, nil 
	}
	return client.GetActiveDownloads()
}

func (m *DynamicManager) MarkAsImported(hash string) error {
	client, err := m.getPrimaryClient()
	if err != nil {
		return err
	}
	return client.MarkAsImported(hash)
}

func (m *DynamicManager) Search(query string) ([]indexer.SearchResult, error) {
	indexers, err := m.IndexerStore.GetAll()
	if err != nil || len(indexers) == 0 {
		return nil, fmt.Errorf("No indexers configured yet: %v", err)
	}

	var aggregatedResults []indexer.SearchResult

	for _, idxConfig := range indexers {
		if !idxConfig.EnableSearch {
			continue
		}

		var activeIndexer indexer.Indexer
		if idxConfig.Name == "Nyaa" || idxConfig.Name == "nyaa" {
			activeIndexer = indexer.NewNyaaIndexer(idxConfig) 
		}

		if activeIndexer == nil {
			continue
		}

		results, err := activeIndexer.Search(query)
		if err != nil {
			continue 
		}
		aggregatedResults = append(aggregatedResults, results...)
	}

	return aggregatedResults, nil
}
