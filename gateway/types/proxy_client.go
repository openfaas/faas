package types

import (
	"net"
	"net/http"
	"net/url"
	"time"
)

func NewHttpClientReverseProxy(baseURL *url.URL, timeout time.Duration) *HttpClientReverseProxy {
	h := HttpClientReverseProxy{
		BaseURL: baseURL,
	}

	h.Client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 1 * time.Second,
			}).DialContext,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}
	return &h
}

type HttpClientReverseProxy struct {
	BaseURL *url.URL
	Client  *http.Client
}
