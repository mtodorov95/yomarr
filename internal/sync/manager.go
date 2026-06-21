package sync

import (
	"fmt"
	"log"

	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/download"
	"github.com/mtodorov95/yomarr/internal/indexer"
	"github.com/mtodorov95/yomarr/internal/models"
)

type DynamicManager struct {
	IndexerStore db.IndexerStore
	ClientStore  db.DownloadClientStore
	QueueStore db.QueueStore
}

var _ indexer.Indexer = (*DynamicManager)(nil)
var _ download.DownloadClient = (*DynamicManager)(nil)

func NewDynamicManager(idxStore db.IndexerStore, dcStore db.DownloadClientStore, qStore db.QueueStore) *DynamicManager {
	return &DynamicManager{
		IndexerStore: idxStore,
		ClientStore:  dcStore,
		QueueStore: qStore,
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

func (m *DynamicManager) AddTorrentFromURL(url string, savePath string, seedTime int, language string, seriesID int64, release indexer.ParsedRelease) (string, error) {
	client, err := m.getPrimaryClient()
	if err != nil {
		return "", err
	}
	hash, err := client.AddTorrentFromURL(url, savePath, seedTime, language, seriesID, release)
	if err != nil {
		return "", err
	}

	if hash != "" {
		queueItem := &models.QueueItem{
			TorrentHash: hash,
			SeriesID:    seriesID,
			ReleaseType: release.Type,
			StartNum:    release.StartNum,
			EndNum:      release.EndNum,
			Language:    language,
			Status:      models.QueueDownloading,
		}
		
		_ = m.QueueStore.Insert(queueItem)
	}

	return hash, nil
}

func (m *DynamicManager) AddTorrentFromMagnet(magnet string, savePath string, seedTime int, language string, seriesID int64, release indexer.ParsedRelease) (string, error) {
	client, err := m.getPrimaryClient()
	if err != nil {
		return "", err
	}

	hash, err := client.AddTorrentFromMagnet(magnet, savePath, seedTime, language, seriesID, release)
	if err != nil {
		return "", err
	}

	if hash != "" {
		queueItem := &models.QueueItem{
			TorrentHash: hash,
			SeriesID:    seriesID,
			ReleaseType: release.Type,
			StartNum:    release.StartNum,
			EndNum:      release.EndNum,
			Language:    language,
			Status:      models.QueueDownloading,
		}
		_ = m.QueueStore.Insert(queueItem)
	}

	return hash, nil
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

	if err := client.MarkAsImported(hash); err != nil {
        log.Printf("[Manager] Downloader client failed to finalize %s: %v", hash, err)
    }

	if err := m.QueueStore.Remove(hash); err != nil {
        return fmt.Errorf("failed to delete queue item after import: %w", err)
    }

	return nil
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

func (m *DynamicManager) FetchLatestRSS() ([]indexer.SearchResult, error) {
    indexers, err := m.IndexerStore.GetAll()
    if err != nil || len(indexers) == 0 {
        return nil, fmt.Errorf("no indexers configured yet: %v", err)
    }

    var aggregatedResults []indexer.SearchResult

    for _, idxConfig := range indexers {
        if !idxConfig.EnableRSS {
            continue
        }

        var activeIndexer indexer.Indexer
        if idxConfig.Name == "Nyaa" || idxConfig.Name == "nyaa" {
            activeIndexer = indexer.NewNyaaIndexer(idxConfig) 
        }

        if activeIndexer == nil {
            continue
        }

        results, err := activeIndexer.FetchLatestRSS() 
        if err != nil {
            continue 
        }
        aggregatedResults = append(aggregatedResults, results...)
    }

    return aggregatedResults, nil
}
