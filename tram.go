package main

import (
	"bytes"
	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/regretable"
	"github.com/ngerakines/tram/util"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	_ "time"
)

type WriteOnClose struct {
	Path     string
	ReadData []byte
	R        io.ReadCloser
}

func (c *WriteOnClose) Read(data []byte) (n int, err error) {
	c.ReadData = append(c.ReadData, data...)
	n, err = c.R.Read(data)
	return
}
func (c WriteOnClose) Close() error {
	ioutil.WriteFile(c.Path, c.ReadData, 0777)
	return c.R.Close()
}

func main() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	cwd := util.CWD()
	cachePath := filepath.Join(cwd, ".cache")
	err := os.MkdirAll(cachePath, 0777)
	if err != nil {
		panic(err)
	}

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			requestCachePath := filepath.Join(cachePath, util.Hash([]byte(ctx.Req.URL.String())))
			info, err := os.Stat(requestCachePath)
			if err == nil && !info.IsDir() {
				data, err := ioutil.ReadFile(requestCachePath)
				if err == nil {
					return r, NewBinaryResponse(r, data)
				}
			}
			return r, nil
		})

	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		regret := regretable.NewRegretableReaderCloser(resp.Body)
		resp.Body = regret

		requestCachePath := filepath.Join(cachePath, util.Hash([]byte(ctx.Req.URL.String())))
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			regret.Regret()
			return resp
		}
		ioutil.WriteFile(requestCachePath, data, 0777)

		buf := bytes.NewBuffer([]byte{})
		resp.Body = ioutil.NopCloser(buf)
		return resp
	})
	log.Fatal(http.ListenAndServe(":7040", proxy))
}

func NewBinaryResponse(r *http.Request, body []byte) *http.Response {
	resp := &http.Response{}
	resp.Request = r
	resp.TransferEncoding = r.TransferEncoding
	resp.Header = make(http.Header)
	resp.Header.Add("Content-Type", "application/octet-stream")
	resp.StatusCode = http.StatusOK
	buf := bytes.NewBuffer(body)
	resp.ContentLength = int64(buf.Len())
	resp.Body = ioutil.NopCloser(buf)
	return resp
}
