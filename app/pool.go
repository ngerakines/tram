package app

import (
	"sync"
	"time"
)

type DownloadPool struct {
	mu   sync.Mutex
	urls map[string]time.Time
}

func (d *DownloadPool) Download(url string) {
	d.mu.Lock()
	d.urls[url] = time.Now()
	d.mu.Unlock()
}

func (d *DownloadPool) Finished(url string) {
	d.mu.Lock()
	delete(d.urls, url)
	d.mu.Unlock()
}

func (d *DownloadPool) IsInTransit(url string) (result bool) {
	d.mu.Lock()
	_, result = d.urls[url]
	d.mu.Unlock()
	return
}
