// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.

package types

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/openfaas/faas/gateway/version"
)

// NewHTTPClientReverseProxy proxies to an upstream host through the use of a http.Client
func NewHTTPClientReverseProxy(baseURL *url.URL, timeout time.Duration, maxIdleConns, maxIdleConnsPerHost int) *HTTPClientReverseProxy {
	h := HTTPClientReverseProxy{
		BaseURL: baseURL,
		Timeout: timeout,
	}

	h.Client = http.DefaultClient
	h.Timeout = timeout
	h.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// These overrides for the default client enable re-use of connections and prevent
	// CoreDNS from rate limiting the gateway under high traffic
	//
	// See also two similar projects where this value was updated:
	// https://github.com/prometheus/prometheus/pull/3592
	// https://github.com/minio/minio/pull/5860

	// Taken from http.DefaultTransport in Go 1.11
	h.Client.Transport = &proxyTransport{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: timeout,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          maxIdleConns,
			MaxIdleConnsPerHost:   maxIdleConnsPerHost,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	return &h
}

// HTTPClientReverseProxy proxy to a remote BaseURL using a http.Client
type HTTPClientReverseProxy struct {
	BaseURL *url.URL
	Client  *http.Client
	Timeout time.Duration
}

// proxyTransport is an http.RoundTripper for the reverse proxy client.
// It ensures default headers like the `User-Agent` are set on requests.
type proxyTransport struct {
	// Transport is the underlying HTTP transport to use when making requests.
	Transport http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (t *proxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if _, ok := req.Header["User-Agent"]; !ok {
		req.Header.Set("User-Agent", fmt.Sprintf("openfaas-ce-gateway/%s", version.BuildVersion()))
	}

	return t.Transport.RoundTrip(req)
}
