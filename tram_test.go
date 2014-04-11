package main

import (
	"github.com/codegangsta/martini"
	"github.com/ngerakines/tram/app"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Test_Basic(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	cacheDirectory := filepath.Join(pwd, ".cache1")
	os.MkdirAll(cacheDirectory, 00777)

	newDiskFileCache := app.NewDiskFileCache(cacheDirectory)
	defer newDiskFileCache.Close()

	m := martini.Classic()
	m.Use(app.NewFileCacheMiddleware(newDiskFileCache))

	m.Any("/", app.HandleIndex)

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

	os.RemoveAll(cacheDirectory)
}

func Test_Get(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	cacheDirectory := filepath.Join(pwd, ".cache2")
	os.MkdirAll(cacheDirectory, 00777)

	newDiskFileCache := app.NewDiskFileCache(cacheDirectory)
	defer newDiskFileCache.Close()

	m := martini.Classic()
	m.Use(app.NewFileCacheMiddleware(newDiskFileCache))

	m.Any("/", app.HandleIndex)
	time.Sleep(1000)

	func() {
		uri := "/?url=" + url.QueryEscape("http://www.github.com/") + "&alias=github"
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", uri, nil)
		m.ServeHTTP(res, req)

		if res.Code != 200 {
			t.Error("File should be downloaded.")
		}
	}()

	os.RemoveAll(cacheDirectory)
}
