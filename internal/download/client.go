package download

import (
	"errors"
)

var ErrTorrentExists = errors.New("torrent already downloading or exists in client")

type TorrentInfo struct {
	Hash     string
	Name     string
	Progress float64 
	Tags     []string
}

type DownloadClient interface {
	AddTorrentFromMagnet(magnet string, savePath string, seedTime int, language string) error
	AddTorrentFromURL(url string, savePath string, seedTime int, language string) error
	GetActiveDownloads() ([]TorrentInfo, error)
	MarkAsImported(hash string) error
}
