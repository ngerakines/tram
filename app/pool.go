package app

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type RemoteFileFetcher func(url string) ([]byte, error)

type DedupingDownloader struct {
	downloader   RemoteFileFetcher
	downloadPool *DownloadPool
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

func (dd *DedupingDownloader) download(url string) ([]byte, error) {
	if dd.downloadPool.IsInTransit(url) {
		fmt.Println("Cannot download", url, "because it is already in transit.")
		return nil, DownloadError{"Url already being downloaded"}
	}
	dd.downloadPool.Download(url)
	body, error := dd.downloader(url)
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
	dedupingDownloader.downloader = downloader
	dedupingDownloader.downloadPool = NewDownloadPool()
	return dedupingDownloader.downloader
}

func defaultRemoteFileFetcher(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return body, nil
}
