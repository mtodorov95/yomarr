package download

import (
	"errors"

	"github.com/mtodorov95/yomarr/internal/indexer"
)

var ErrTorrentExists = errors.New("torrent already downloading or exists in client")

type TorrentInfo struct {
	Hash     string
	Name     string
	Progress float64
	Tags     []string
}

type DownloadClient interface {
	AddTorrentFromMagnet(magnet string, savePath string, seedTime int, language string, seriesID int64, release indexer.ParsedRelease) (string, error)
	AddTorrentFromURL(url string, savePath string, seedTime int, language string, seriesID int64, release indexer.ParsedRelease) (string, error)
	GetActiveDownloads() ([]TorrentInfo, error)
	MarkAsImported(hash string) error
}
