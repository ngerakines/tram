package app

import (
	"github.com/ngerakines/tram/storage"
	"testing"
	"time"
)

func TestShouldNotify(t *testing.T) {
	if shouldNotify(test_createCachedFile("350ea0dd9", "http://github.com/", "350ea0dd9", []string{"github"}), test_createDownloadListener("http://www.google.com/", []string{})) {
		t.Error("No url or aliases")
	}
	if !shouldNotify(test_createCachedFile("350ea0dd9", "http://github.com/", "350ea0dd9", []string{"github"}), test_createDownloadListener("http://github.com/", []string{})) {
		t.Error("Url matches")
	}
	if !shouldNotify(test_createCachedFile("350ea0dd9", "http://asset/350ea0dd9", "350ea0dd9", []string{"350ea0dd9"}), test_createDownloadListener("http://asset/350ea0dd9?ts=1", []string{"350ea0dd9"})) {
		t.Error("Alias exists matches")
	}
}

func test_createCachedFile(hash, url, path string, aliases []string) storage.CachedFile {
	return storage.NewLocalCachedFile(hash, path, []string{url}, aliases)
}

func test_createDownloadListener(url string, aliases []string) DownloadListener {
	return DownloadListener{time.Now(), url, aliases, make(chan storage.CachedFile)}
}
