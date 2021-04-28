// Package internal provides ...
package internal

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func newRequest(method, rawurl string, data interface{}) (*http.Request, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	originHost := u.Host
	// 获取主机IP
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		addrErr := err.(*net.AddrError)
		if addrErr.Err != "missing port in address" {
			return nil, err
		}
		// set default value
		host = originHost
		switch u.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		}
	}
	ips, err := net.LookupHost(host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("http: not found ip(%s)", u.Host)
	}
	host = net.JoinHostPort(ips[0], port)
	u.Host = host
	// 创建HTTP Request
	var req *http.Request
	switch raw := data.(type) {
	case url.Values:
		req, err = http.NewRequest(method, u.String(),
			strings.NewReader(raw.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	case []byte:
		req, err = http.NewRequest(method, u.String(),
			bytes.NewReader(raw))
	case nil:
		req, err = http.NewRequest(method, u.String(), nil)
	default:
		return nil, fmt.Errorf("http: unsupported data type: %T", data)
	}
	if err != nil {
		return nil, err
	}
	// 设置Host
	req.Host = originHost
	return req, nil
}

// httpHead HTTP HEAD请求
func httpHead(rawurl string) (*http.Response, error) {
	req, err := newRequest(http.MethodHead, rawurl, nil)
	if err != nil {
		return nil, err
	}
	return httpClient.Do(req)
}

// httpGet HTTP GET请求
func httpGet(rawurl string) (*http.Response, error) {
	req, err := newRequest(http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, err
	}
	// 发起请求
	return httpClient.Do(req)
}

// httpPost HTTP POST请求, 自动识别是否是form
func httpPost(rawurl string, data interface{}) (*http.Response, error) {
	req, err := newRequest(http.MethodPost, rawurl, data)
	if err != nil {
		return nil, err
	}
	// 发起请求
	return httpClient.Do(req)
}

// httpPostHeader HTTP POST请求，自定义Header
func httpPostHeader(rawurl string, data interface{},
	header http.Header) (*http.Response, error) {

	req, err := newRequest(http.MethodPost, rawurl, data)
	if err != nil {
		return nil, err
	}
	// set header
	req.Header = header
	// 发起请求
	return httpClient.Do(req)
}

// httpPut HTTP PUT请求
func httpPut(rawurl string, data interface{}) (*http.Response, error) {
	req, err := newRequest(http.MethodPut, rawurl, data)
	if err != nil {
		return nil, err
	}
	// 发起请求
	return httpClient.Do(req)
}
