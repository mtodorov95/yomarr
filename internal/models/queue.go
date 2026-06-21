package models

import (
	"time"
)

type ReleaseType string

const (
	TypeSingle ReleaseType = "single"
	TypeRange  ReleaseType = "range"
	TypeVolume ReleaseType = "volume"
)

type QueueStatus string

const (
	QueueDownloading  QueueStatus = "downloading"
	QueueImporting    QueueStatus = "importing"
	QueueFailedImport QueueStatus = "failed_import"
)

type QueueItem struct {
	TorrentHash  string      `json:"torrent_hash"`
	SeriesID     int64       `json:"series_id"`
	ReleaseType  ReleaseType `json:"release_type"`
	StartNum     float64     `json:"start_num"`
	EndNum       float64     `json:"end_num"`
	Language     string      `json:"language"`
	Status       QueueStatus `json:"status"`
	ErrorMessage *string     `json:"error_message,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}
