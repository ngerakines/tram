package app

import (
	"github.com/ngerakines/tram/util"
	"sync"
	"time"
)

type DownloadListeners struct {
	mu        sync.Mutex
	listeners map[string]DownloadListener
	um        *util.UidManager
}

type DownloadListener struct {
	when    time.Time
	url     string
	aliases []string
	channel chan CachedFile
}

func NewDownloadListeners() *DownloadListeners {
	downloadListeners := new(DownloadListeners)
	downloadListeners.listeners = make(map[string]DownloadListener)
	downloadListeners.mu = sync.Mutex{}
	downloadListeners.um = util.NewUidManager()
	return downloadListeners
}

func (downloadListeners *DownloadListeners) Add(url string, aliases []string, channel chan CachedFile) {
	downloadListener := DownloadListener{when: time.Now(), url: url, aliases: aliases, channel: channel}
	downloadListeners.mu.Lock()
	downloadListeners.listeners[downloadListeners.um.GenerateHex()] = downloadListener
	downloadListeners.mu.Unlock()
}

func (downloadListeners *DownloadListeners) Notify(cachedFile CachedFile) {
	downloadListeners.mu.Lock()
	toRemove := make([]string, 0, 0)
	for key, downloadListener := range downloadListeners.listeners {
		if shouldNotify(cachedFile, downloadListener) {
			downloadListener.channel <- cachedFile
			toRemove = append(toRemove, key)
		}
	}
	for _, key := range toRemove {
		delete(downloadListeners.listeners, key)
	}
	downloadListeners.mu.Unlock()
}

func shouldNotify(cachedFile CachedFile, downloadListener DownloadListener) bool {
	for _, url := range cachedFile.Urls() {
		if downloadListener.url == url {
			return true
		}
	}
	// NKG: This can be improved.
	for _, alias := range cachedFile.Aliases() {
		for _, alias2 := range downloadListener.aliases {
			if alias == alias2 {
				return true
			}
		}
	}
	return false
}
