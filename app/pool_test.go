package app

import (
	"sync"
	"testing"
	"time"
)

type mockDownloader struct {
	payloads map[string][]byte
	counts   map[string]int
	mu       sync.Mutex
}

type stringError struct {
	message string
}

func (err stringError) Error() string {
	return err.message
}

func (mockDownloader *mockDownloader) download(url string) ([]byte, error) {
	payload, hasPayload := mockDownloader.payloads[url]
	if hasPayload {
		mockDownloader.mu.Lock()
		time.Sleep(2000)
		value, hasValue := mockDownloader.counts[url]
		if !hasValue {
			value = 0
		}
		value += 1
		mockDownloader.counts[url] = value
		mockDownloader.mu.Unlock()
		return payload, nil
	}
	return nil, stringError{"No url in mock downloader."}
}

func TestInTransit(t *testing.T) {
	dp := NewDownloadPool()
	if dp.IsInTransit("http://localhost/1") == true {
		t.Error("Url 'http://localhost/1' should not be in transit.")
	}

	dp.Download("http://localhost/1")

	if dp.IsInTransit("http://localhost/1") == false {
		t.Error("Url 'http://localhost/1' should be in transit.")
	}

	dp.Finished("http://localhost/1")

	if dp.IsInTransit("http://localhost/1") == true {
		t.Error("Url 'http://localhost/1' should not be in transit.")
	}

}

func TestPoolMany(t *testing.T) {
	dp := NewDownloadPool()
	if dp.IsInTransit("http://localhost/1") == true {
		t.Error("Url 'http://localhost/1' should not be in transit.")
	}
	if dp.IsInTransit("http://localhost/2") == true {
		t.Error("Url 'http://localhost/2' should not be in transit.")
	}

	dp.Download("http://localhost/1")
	dp.Download("http://localhost/1")
	dp.Download("http://localhost/2")

	if dp.IsInTransit("http://localhost/1") == false {
		t.Error("Url 'http://localhost/1' should be in transit.")
	}
	if dp.IsInTransit("http://localhost/2") == false {
		t.Error("Url 'http://localhost/2' should be in transit.")
	}

	dp.Finished("http://localhost/1")

	if dp.IsInTransit("http://localhost/1") == true {
		t.Error("Url 'http://localhost/1' should not be in transit.")
	}
	if dp.IsInTransit("http://localhost/2") == false {
		t.Error("Url 'http://localhost/2' should be in transit.")
	}
}

func TestDedupe(t *testing.T) {
	md := new(mockDownloader)
	md.mu = sync.Mutex{}
	md.counts = make(map[string]int)
	md.payloads = make(map[string][]byte)
	md.payloads["http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"] = []byte("/")
	md.payloads["http://localhost:3001/ef090dcea7b507772498cd2e67f2b148ae2609f6"] = []byte("/tram")
	md.payloads["http://localhost:3001/a11f846da74df08c2e93ede56beefdde735ccc05"] = []byte("/tram-chef-cookbook")

	downloader := DedupeWrapDownloader(md.download)

	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		_, err := downloader("http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8")
		if err != nil {
			t.Error(err.Error())
		}
		defer wg.Done()
	}()
	go func() {
		wg.Add(1)
		_, err := downloader("http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8")
		if err == nil {
			t.Error(err.Error())
		}
		defer wg.Done()
	}()
	wg.Wait()
}
