package util

import (
	"io/ioutil"
	"log"
	"sync"
	"time"
)

type DedupingDownloader struct {
	wrappedDownloader RemoteFileFetcher
	downloadPool      *DownloadPool
}

type DownloadError struct {
	message string
}

type DownloadPool struct {
	mu   sync.Mutex
	urls map[string]time.Time
}

func (err DownloadError) Error() string {
	return err.message
}

func (dd *DedupingDownloader) downloader(url string) ([]byte, error) {
	if dd.downloadPool.IsInTransit(url) {
		log.Println("Cannot download", url, "because it is already in transit.")
		return nil, DownloadError{"Url already being downloaded"}
	}
	dd.downloadPool.Download(url)
	body, error := dd.wrappedDownloader(url)
	dd.downloadPool.Finished(url)
	return body, error
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

func NewDownloadPool() *DownloadPool {
	downloadPool := new(DownloadPool)
	downloadPool.mu = sync.Mutex{}
	downloadPool.urls = make(map[string]time.Time)
	return downloadPool
}

func DedupeWrapDownloader(downloader RemoteFileFetcher) RemoteFileFetcher {
	dedupingDownloader := new(DedupingDownloader)
	dedupingDownloader.wrappedDownloader = downloader
	dedupingDownloader.downloadPool = NewDownloadPool()
	return dedupingDownloader.downloader
}

func DefaultRemoteFileFetcher(url string) ([]byte, error) {
	httpClient := NewHttpClient(false, 30*time.Second)
	resp, err := httpClient.Get(url)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return body, nil
}
