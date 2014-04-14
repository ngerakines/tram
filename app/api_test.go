package app

import (
	"github.com/codegangsta/martini"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"sync"
)

type MockDownloader struct {
	payloads map[string][]byte
	counts map[string]int
	mu   sync.Mutex
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

	return DiskFileCacheConfig{mockDownloader.download, cacheDirectory}, mockDownloader
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
	value, hasValue := mockDownloader.counts["http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"]; if (!hasValue || value != 1) {
		t.Error("Url should have been requested just once.")
	}
	value, hasValue = mockDownloader.counts["http://localhost:3001/ef090dcea7b507772498cd2e67f2b148ae2609f6"]; if (!hasValue || value != 1) {
		t.Error("Url should have been requested just once.")
	}
	value, hasValue = mockDownloader.counts["http://localhost:3001/a11f846da74df08c2e93ede56beefdde735ccc05"]; if (!hasValue || value != 1) {
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
	value, hasValue := mockDownloader.counts["http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"]; if (!hasValue || value != 1) {
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
		req, _ := http.NewRequest("POST", "/", strings.NewReader("url=" + url.QueryEscape("http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8") + "&alias="))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		m.ServeHTTP(res, req)
		if res.Code != 202 {
			t.Error("All requests should be accepted.")
		}
	}()

	if len(mockDownloader.counts) != 1 {
		t.Error("mockDownloader.counts should have 1 entry")
	}
	value, hasValue := mockDownloader.counts["http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"]; if (!hasValue || value != 1) {
		t.Error("Url should have been requested just once.")
	}

	os.RemoveAll(config.basePath)
}
