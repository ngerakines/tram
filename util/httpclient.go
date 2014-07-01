package util

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

func NewHttpClient(verifySsl bool, timeout time.Duration) *http.Client {
	tr := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: !verifySsl},
		ResponseHeaderTimeout: timeout,
		Dial: TimeoutDialer(5*time.Second, 30*time.Second),
	}
	return &http.Client{Transport: tr}
}

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}
