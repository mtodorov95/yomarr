package download

import "errors"

var ErrTorrentExists = errors.New("torrent already downloading or exists in client")

type DownloadClient interface {
	AddTorrentFromMagnet(magnet string, savePath string) error
	AddTorrentFromURL(url string, savePath string) error
}
