// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"net"
	"net/http"
	"net/url"
	"time"
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
	h.Client.Transport = &http.Transport{
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
	}

	return &h
}

// HTTPClientReverseProxy proxy to a remote BaseURL using a http.Client
type HTTPClientReverseProxy struct {
	BaseURL *url.URL
	Client  *http.Client
	Timeout time.Duration
}
