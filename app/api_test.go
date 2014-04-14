package app

import (
	"github.com/codegangsta/martini"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type MockDownloader struct {
	payloads map[string][]byte
	counts   map[string]int
	mu       sync.Mutex
}

type StringError struct {
	message string
}

func (err StringError) Error() string {
	return err.message
}

func (mockDownloader *MockDownloader) download(url string) ([]byte, error) {
	payload, hasPayload := mockDownloader.payloads[url]
	if hasPayload {
		mockDownloader.mu.Lock()
		value, hasValue := mockDownloader.counts[url]
		if !hasValue {
			value = 0
		}
		value += 1
		mockDownloader.counts[url] = value
		mockDownloader.mu.Unlock()
		return payload, nil
	}
	return nil, StringError{"No url in mock downloader."}
}

func buildConfig(cacheDir string) (DiskFileCacheConfig, *MockDownloader) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	cacheDirectory := filepath.Join(pwd, cacheDir)
	os.MkdirAll(cacheDirectory, 00777)

	mockDownloader := new(MockDownloader)
	mockDownloader.mu = sync.Mutex{}
	mockDownloader.counts = make(map[string]int)
	mockDownloader.payloads = make(map[string][]byte)
	mockDownloader.payloads["http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"] = []byte("/")
	mockDownloader.payloads["http://localhost:3001/ef090dcea7b507772498cd2e67f2b148ae2609f6"] = []byte("/tram")
	mockDownloader.payloads["http://localhost:3001/a11f846da74df08c2e93ede56beefdde735ccc05"] = []byte("/tram-chef-cookbook")
	mockDownloader.payloads["http://localhost:3001/974a0682b121adda8d8bb4503d07672c6d65319c"] = []byte("/commitment")
	mockDownloader.payloads["http://localhost:3001/f6cd811a6b41ae88b1aa5d0e7ab8e16cda349917"] = []byte("/erlang_twitter")
	mockDownloader.payloads["http://localhost:3001/72a57722a7d426ca21bf52348f3a83c96a3cc72b"] = []byte("/erlang_protobuffs")
	mockDownloader.payloads["http://localhost:3001/fca08fb51c04e2cf57d2d78e229a28ff10a2c2d7"] = []byte("/erlang_couchdb")
	mockDownloader.payloads["http://localhost:3001/db28cadbe22a8dcbd65659c833ea91b8ad1246e5"] = []byte("/etap")
	mockDownloader.payloads["http://localhost:3001/81720de9afb950d0e5f04d01c9e99e03e6425af0"] = []byte("/wallbase-cc-downloader")
	mockDownloader.payloads["http://localhost:3001/6ce8a69121f4fdcb156772ff00c3828ae542f00b"] = []byte("/miniature-dubstep")
	mockDownloader.payloads["http://localhost:3001/ffdefcd0d443d73f4058e33bec27c5cabb2ac1c1"] = []byte("/elasticservices")


	return DiskFileCacheConfig{DedupeWrapDownloader(mockDownloader.download), cacheDirectory, 24}, mockDownloader
}

func TestEmpty(t *testing.T) {
	config, mockDownloader := buildConfig(".cache1")
	newDiskFileCache := NewDiskFileCache(config)
	defer newDiskFileCache.Close()

	m := martini.Classic()
	m.Use(NewFileCacheMiddleware(newDiskFileCache))

	m.Any("/", HandleIndex)

	func() {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/?query=foo", nil)
		m.ServeHTTP(res, req)

		if res.Code != 404 {
			t.Error("No file should be found.")
		}
	}()

	func() {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", "/?url=foo", nil)
		m.ServeHTTP(res, req)

		if res.Code != 404 {
			t.Error("No file should be found.")
		}
	}()

	func() {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/", strings.NewReader("url=nope&alias="))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		m.ServeHTTP(res, req)
		if res.Code != 202 {
			t.Error("All requests should be accepted.")
		}
	}()

	if len(mockDownloader.counts) > 0 {
		t.Error("mockDownloader.counts should be empty")
	}

	os.RemoveAll(newDiskFileCache.config.basePath)
}

func TestGet(t *testing.T) {
	config, mockDownloader := buildConfig(".cache2")
	newDiskFileCache := NewDiskFileCache(config)
	defer newDiskFileCache.Close()

	m := martini.Classic()
	m.Use(NewFileCacheMiddleware(newDiskFileCache))

	m.Any("/", HandleIndex)
	time.Sleep(500)

	func() {
		uri := "/?url=" + url.QueryEscape("http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8") + "&alias=base"
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", uri, nil)
		m.ServeHTTP(res, req)

		if res.Code != 200 {
			t.Error("File should be downloaded.")
		}
		filePath := filepath.Join(config.basePath, "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("File doesn't exist at " + filePath)
		}
		if _, err := os.Stat(filePath + ".metadata"); os.IsNotExist(err) {
			t.Error("File metadata doesn't exist at " + filePath + ".metadata")
		}
	}()

	func() {
		uri := "/?url=" + url.QueryEscape("http://localhost:3001/ef090dcea7b507772498cd2e67f2b148ae2609f6") + "&alias=tram"
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", uri, nil)
		m.ServeHTTP(res, req)

		if res.Code != 200 {
			t.Error("File should be downloaded.")
		}
		filePath := filepath.Join(config.basePath, "ef090dcea7b507772498cd2e67f2b148ae2609f6")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("File doesn't exist at " + filePath)
		}
		if _, err := os.Stat(filePath + ".metadata"); os.IsNotExist(err) {
			t.Error("File metadata doesn't exist at " + filePath + ".metadata")
		}
	}()

	func() {
		uri := "/?url=" + url.QueryEscape("http://localhost:3001/a11f846da74df08c2e93ede56beefdde735ccc05") + "&alias=tram-chef-cookbook"
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", uri, nil)
		m.ServeHTTP(res, req)

		if res.Code != 200 {
			t.Error("File should be downloaded.")
		}
		filePath := filepath.Join(config.basePath, "a11f846da74df08c2e93ede56beefdde735ccc05")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("File doesn't exist at " + filePath)
		}
		if _, err := os.Stat(filePath + ".metadata"); os.IsNotExist(err) {
			t.Error("File metadata doesn't exist at " + filePath + ".metadata")
		}
	}()

	if len(mockDownloader.counts) != 3 {
		t.Error("mockDownloader.counts should have 3 entries")
	}
	value, hasValue := mockDownloader.counts["http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}
	value, hasValue = mockDownloader.counts["http://localhost:3001/ef090dcea7b507772498cd2e67f2b148ae2609f6"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}
	value, hasValue = mockDownloader.counts["http://localhost:3001/a11f846da74df08c2e93ede56beefdde735ccc05"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}

	os.RemoveAll(config.basePath)
}

func TestFsIssues(t *testing.T) {
	config, mockDownloader := buildConfig(".cache3")
	newDiskFileCache := NewDiskFileCache(config)
	defer newDiskFileCache.Close()

	m := martini.Classic()
	m.Use(NewFileCacheMiddleware(newDiskFileCache))

	m.Any("/", HandleIndex)
	time.Sleep(500)

	func() {
		uri := "/?url=" + url.QueryEscape("http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8") + "&alias=base"
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", uri, nil)
		m.ServeHTTP(res, req)

		if res.Code != 200 {
			t.Error("File should be downloaded.")
		}
		filePath := filepath.Join(config.basePath, "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("File doesn't exist at " + filePath)
		}
		if _, err := os.Stat(filePath + ".metadata"); os.IsNotExist(err) {
			t.Error("File metadata doesn't exist at " + filePath + ".metadata")
		}
	}()

	func() {
		uri := "/?query=" + url.QueryEscape("http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8")
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", uri, nil)
		m.ServeHTTP(res, req)

		if res.Code != 200 {
			t.Error("File should be downloaded.")
		}
	}()

	os.Remove(filepath.Join(config.basePath, "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"))
	os.Remove(filepath.Join(config.basePath, "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8.metadata"))

	func() {
		uri := "/?query=" + url.QueryEscape("http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8")
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", uri, nil)
		m.ServeHTTP(res, req)

		if res.Code != 404 {
			t.Error("File cannot be downloaded.")
		}
	}()

	func() {
		uri := "/?query=" + url.QueryEscape("http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8")
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", uri, nil)
		m.ServeHTTP(res, req)

		if res.Code != 404 {
			t.Error("File cannot be downloaded.")
		}
	}()

	if len(mockDownloader.counts) != 1 {
		t.Error("mockDownloader.counts should have 1 entry")
	}
	value, hasValue := mockDownloader.counts["http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}

	os.RemoveAll(config.basePath)
}

func TestDoubleDownload(t *testing.T) {
	config, mockDownloader := buildConfig(".cache4")
	newDiskFileCache := NewDiskFileCache(config)
	defer newDiskFileCache.Close()

	m := martini.Classic()
	m.Use(NewFileCacheMiddleware(newDiskFileCache))

	m.Any("/", HandleIndex)
	time.Sleep(500)

	func() {
		uri := "/?url=" + url.QueryEscape("http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8") + "&alias=base"
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", uri, nil)
		m.ServeHTTP(res, req)

		if res.Code != 200 {
			t.Error("File should be downloaded.")
		}
		filePath := filepath.Join(config.basePath, "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("File doesn't exist at " + filePath)
		}
		if _, err := os.Stat(filePath + ".metadata"); os.IsNotExist(err) {
			t.Error("File metadata doesn't exist at " + filePath + ".metadata")
		}
	}()

	func() {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/", strings.NewReader("url="+url.QueryEscape("http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8")+"&alias="))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		m.ServeHTTP(res, req)
		if res.Code != 202 {
			t.Error("All requests should be accepted.")
		}
	}()

	if len(mockDownloader.counts) != 1 {
		t.Error("mockDownloader.counts should have 1 entry")
	}
	value, hasValue := mockDownloader.counts["http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}

	os.RemoveAll(config.basePath)
}

func TestWarm(t *testing.T) {
	config, mockDownloader := buildConfig(".cache5")
	newDiskFileCache := NewDiskFileCache(config)
	defer newDiskFileCache.Close()

	m := martini.Classic()
	m.Use(NewFileCacheMiddleware(newDiskFileCache))

	m.Any("/", HandleIndex)
	time.Sleep(500)

	func() {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/", strings.NewReader("url="+url.QueryEscape("http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8")+"&alias="))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		m.ServeHTTP(res, req)
		if res.Code != 202 {
			t.Error("All requests should be accepted.")
		}
		time.Sleep(250)
		filePath := filepath.Join(config.basePath, "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("File doesn't exist at " + filePath)
		}
		if _, err := os.Stat(filePath + ".metadata"); os.IsNotExist(err) {
			t.Error("File metadata doesn't exist at " + filePath + ".metadata")
		}
	}()

	time.Sleep(1000)

	if len(mockDownloader.counts) != 1 {
		t.Error("mockDownloader.counts should have 1 entry")
	}
	value, hasValue := mockDownloader.counts["http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}

	os.RemoveAll(config.basePath)
}

func TestEviction(t *testing.T) {
	config, mockDownloader := buildConfig(".cache6")
	newDiskFileCache := NewDiskFileCache(config)
	defer newDiskFileCache.Close()

	m := martini.Classic()
	m.Use(NewFileCacheMiddleware(newDiskFileCache))

	m.Any("/", HandleIndex)
	time.Sleep(500)

	request(t, m, "/?url=" + url.QueryEscape("http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8") + "&alias=1", "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", config)
	request(t, m, "/?url=" + url.QueryEscape("http://localhost:3001/ef090dcea7b507772498cd2e67f2b148ae2609f6") + "&alias=2", "ef090dcea7b507772498cd2e67f2b148ae2609f6", config)
	request(t, m, "/?url=" + url.QueryEscape("http://localhost:3001/a11f846da74df08c2e93ede56beefdde735ccc05") + "&alias=3", "a11f846da74df08c2e93ede56beefdde735ccc05", config)
	request(t, m, "/?url=" + url.QueryEscape("http://localhost:3001/974a0682b121adda8d8bb4503d07672c6d65319c") + "&alias=4", "974a0682b121adda8d8bb4503d07672c6d65319c", config)
	request(t, m, "/?url=" + url.QueryEscape("http://localhost:3001/f6cd811a6b41ae88b1aa5d0e7ab8e16cda349917") + "&alias=5", "f6cd811a6b41ae88b1aa5d0e7ab8e16cda349917", config)
	request(t, m, "/?url=" + url.QueryEscape("http://localhost:3001/72a57722a7d426ca21bf52348f3a83c96a3cc72b") + "&alias=6", "72a57722a7d426ca21bf52348f3a83c96a3cc72b", config)
	request(t, m, "/?url=" + url.QueryEscape("http://localhost:3001/fca08fb51c04e2cf57d2d78e229a28ff10a2c2d7") + "&alias=7", "fca08fb51c04e2cf57d2d78e229a28ff10a2c2d7", config)
	request(t, m, "/?url=" + url.QueryEscape("http://localhost:3001/db28cadbe22a8dcbd65659c833ea91b8ad1246e5") + "&alias=8", "db28cadbe22a8dcbd65659c833ea91b8ad1246e5", config)

	time.Sleep(5000)

	if len(mockDownloader.counts) != 8 {
		t.Error("mockDownloader.counts should have 8 entry")
	}
	value, hasValue := mockDownloader.counts["http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}
	value, hasValue = mockDownloader.counts["http://localhost:3001/ef090dcea7b507772498cd2e67f2b148ae2609f6"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}
	value, hasValue = mockDownloader.counts["http://localhost:3001/a11f846da74df08c2e93ede56beefdde735ccc05"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}
	value, hasValue = mockDownloader.counts["http://localhost:3001/974a0682b121adda8d8bb4503d07672c6d65319c"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}
	value, hasValue = mockDownloader.counts["http://localhost:3001/f6cd811a6b41ae88b1aa5d0e7ab8e16cda349917"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}
	value, hasValue = mockDownloader.counts["http://localhost:3001/72a57722a7d426ca21bf52348f3a83c96a3cc72b"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}
	value, hasValue = mockDownloader.counts["http://localhost:3001/fca08fb51c04e2cf57d2d78e229a28ff10a2c2d7"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}
	value, hasValue = mockDownloader.counts["http://localhost:3001/db28cadbe22a8dcbd65659c833ea91b8ad1246e5"]
	if !hasValue || value != 1 {
		t.Error("Url should have been requested just once.")
	}

	checkFileNotExists(t, "42099b4af021e53fd8fd4e056c2568d7c2e3ffa8", config)
	checkFileNotExists(t, "ef090dcea7b507772498cd2e67f2b148ae2609f6", config)
	checkFileNotExists(t, "a11f846da74df08c2e93ede56beefdde735ccc05", config)
	checkFileNotExists(t, "974a0682b121adda8d8bb4503d07672c6d65319c", config)
	checkFileNotExists(t, "f6cd811a6b41ae88b1aa5d0e7ab8e16cda349917", config)
	checkFileNotExists(t, "72a57722a7d426ca21bf52348f3a83c96a3cc72b", config)
	checkFileExists(t, "fca08fb51c04e2cf57d2d78e229a28ff10a2c2d7", config)
	checkFileExists(t, "db28cadbe22a8dcbd65659c833ea91b8ad1246e5", config)

	os.RemoveAll(config.basePath)
}

func request(t *testing.T, m *martini.ClassicMartini, uri, fileName string, config DiskFileCacheConfig) {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", uri, nil)
		m.ServeHTTP(res, req)

		if res.Code != 200 {
			t.Error("Request not successful: %s", uri)
		}
	checkFileExists(t, fileName, config)
}

func checkFileExists(t *testing.T, fileName string, config DiskFileCacheConfig) {
	filePath := filepath.Join(config.basePath, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File doesn't exist at " + filePath)
	}
	if _, err := os.Stat(filePath + ".metadata"); os.IsNotExist(err) {
		t.Error("File metadata doesn't exist at " + filePath + ".metadata")
	}
}

func checkFileNotExists(t *testing.T, fileName string, config DiskFileCacheConfig) {
	filePath := filepath.Join(config.basePath, fileName)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("File doesn't exist at " + filePath)
	}
	if _, err := os.Stat(filePath + ".metadata"); !os.IsNotExist(err) {
		t.Error("File metadata doesn't exist at " + filePath + ".metadata")
	}
}

/*
TODO:

TestDedupe
 - Create a mock client that takes 2-3 seconds to return
 - Send multiple requests for the same url
 - Verify that only one request was made

TestListeners
 - Create a mock client
 - Send multiple requests for two urls
 - Verify that the different requests didn't get the same content
*/
