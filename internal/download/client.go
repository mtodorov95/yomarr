package download

import (
	"errors"
	"time"
)

var ErrTorrentExists = errors.New("torrent already downloading or exists in client")

type DownloadClient interface {
	AddTorrentFromMagnet(magnet string, savePath string, seedDuration time.Duration, language string) error
	AddTorrentFromURL(url string, savePath string, seedDuration time.Duration, language string) error
}
