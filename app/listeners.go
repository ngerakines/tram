package app

import "time"

type DownloadListeners struct {
	listeners []DownloadListener
}

type DownloadListener struct {
	when time.Time
	url string
	aliases []string
	channel chan *CachedFile
}

func NewDownloadListeners() *DownloadListeners {
	downloadListeners := new(DownloadListeners)
	downloadListeners.listeners = make([]DownloadListener, 0, 0)
	return downloadListeners
}

func (downloadListeners *DownloadListeners) Add(url string, aliases []string, channel chan *CachedFile) {
	downloadListener := DownloadListener{when: time.Now(), url: url, aliases: aliases, channel: channel }
	downloadListeners.listeners = append(downloadListeners.listeners, downloadListener)
}

func (downloadListeners *DownloadListeners) Notify(cachedFile *CachedFile) {
	for _, downloadListener := range downloadListeners.listeners {
		if shouldNotify(cachedFile, downloadListener) {
			downloadListener.channel <- cachedFile
		}
	}
}

func shouldNotify(cachedFile *CachedFile, downloadListener DownloadListener) bool {
	if downloadListener.url == cachedFile.Url {
		return true;
	}
	// NKG: This can be improved.
	for _, alias := range cachedFile.Aliases {
		for _, alias2 := range downloadListener.aliases {
			if alias == alias2 {
				return true
			}
		}
	}
	return false
}
