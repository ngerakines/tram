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
)

type MockDownloader struct {
	payloads map[string][]byte
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
		return payload, nil
	}
	return nil, StringError{"No url in mock downloader."}
}

func buildConfig(cacheDir string) DiskFileCacheConfig {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	cacheDirectory := filepath.Join(pwd, cacheDir)
	os.MkdirAll(cacheDirectory, 00777)

	mockDownloader := new(MockDownloader)
	mockDownloader.payloads = make(map[string][]byte)
	mockDownloader.payloads["http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8"] = []byte("/")
	mockDownloader.payloads["http://localhost:3001/ef090dcea7b507772498cd2e67f2b148ae2609f6"] = []byte("/tram")
	mockDownloader.payloads["http://localhost:3001/a11f846da74df08c2e93ede56beefdde735ccc05"] = []byte("/tram-chef-cookbook")

	return DiskFileCacheConfig{mockDownloader.download, cacheDirectory}
}

func Test_Basic(t *testing.T) {
	newDiskFileCache := NewDiskFileCache(buildConfig(".cache1"))
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

	os.RemoveAll(newDiskFileCache.config.basePath)
}

func Test_Get(t *testing.T) {
	config := buildConfig(".cache2")
	newDiskFileCache := NewDiskFileCache(config)
	defer newDiskFileCache.Close()

	m := martini.Classic()
	m.Use(NewFileCacheMiddleware(newDiskFileCache))

	m.Any("/", HandleIndex)
	time.Sleep(500)

	func() {
		uri := "/?url=" + url.QueryEscape("http://localhost:3001/42099b4af021e53fd8fd4e056c2568d7c2e3ffa8") + "&alias=github:ngerakines"
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

	os.RemoveAll(config.basePath)
}

