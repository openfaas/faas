// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/openfaas/faas/gateway/types"
)

// functionMatcher parses out the service name (group 1) and rest of path (group 2).
var functionMatcher = regexp.MustCompile("^/?(?:async-)?function/([^/?]+)([^?]*)")

// Indices and meta-data for functionMatcher regex parts
const (
	hasPathCount = 3
	routeIndex   = 0 // routeIndex corresponds to /function/ or /async-function/
	nameIndex    = 1 // nameIndex is the function name
	pathIndex    = 2 // pathIndex is the path i.e. /employee/:id/
)

// BaseURLResolver URL resolver for upstream requests
type BaseURLResolver interface {
	Resolve(r *http.Request) string
}

// URLPathTransformer Transform the incoming URL path for upstream requests
type URLPathTransformer interface {
	Transform(r *http.Request) string
}

// MakeForwardingProxyHandler create a handler which forwards HTTP requests
func MakeForwardingProxyHandler(proxy *types.HTTPClientReverseProxy, notifiers []HTTPNotifier, baseURLResolver BaseURLResolver, urlPathTransformer URLPathTransformer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseURL := baseURLResolver.Resolve(r)
		originalURL := r.URL.String()

		requestURL := urlPathTransformer.Transform(r)

		start := time.Now()

		statusCode, err := forwardRequest(w, r, proxy.Client, baseURL, requestURL, proxy.Timeout)

		seconds := time.Since(start)
		if err != nil {
			log.Printf("error with upstream request to: %s, %s\n", requestURL, err.Error())
		}

		for _, notifier := range notifiers {
			notifier.Notify(r.Method, requestURL, originalURL, statusCode, seconds)
		}
	}
}

func buildUpstreamRequest(r *http.Request, baseURL string, requestURL string) *http.Request {
	url := baseURL + requestURL

	if len(r.URL.RawQuery) > 0 {
		url = fmt.Sprintf("%s?%s", url, r.URL.RawQuery)
	}

	upstreamReq, _ := http.NewRequest(r.Method, url, nil)

	copyHeaders(upstreamReq.Header, &r.Header)

	if len(r.Host) > 0 && upstreamReq.Header.Get("X-Forwarded-Host") == "" {
		upstreamReq.Header["X-Forwarded-Host"] = []string{r.Host}
	}
	if upstreamReq.Header.Get("X-Forwarded-For") == "" {
		upstreamReq.Header["X-Forwarded-For"] = []string{r.RemoteAddr}
	}

	if r.Body != nil {
		upstreamReq.Body = r.Body
	}

	return upstreamReq
}

func forwardRequest(w http.ResponseWriter, r *http.Request, proxyClient *http.Client, baseURL string, requestURL string, timeout time.Duration) (int, error) {

	upstreamReq := buildUpstreamRequest(r, baseURL, requestURL)
	if upstreamReq.Body != nil {
		defer upstreamReq.Body.Close()
	}

	if _, exists := os.LookupEnv("write_request_uri"); exists {
		log.Printf("forwardRequest: %s %s\n", upstreamReq.Host, upstreamReq.URL.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, resErr := proxyClient.Do(upstreamReq.WithContext(ctx))
	if resErr != nil {
		badStatus := http.StatusBadGateway
		w.WriteHeader(badStatus)
		return badStatus, resErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	copyHeaders(w.Header(), &res.Header)

	// Write status code
	w.WriteHeader(res.StatusCode)

	if res.Body != nil {
		// Copy the body over
		io.CopyBuffer(w, res.Body, nil)
	}

	return res.StatusCode, nil
}

func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		(destination)[k] = vClone
	}
}

// SingleHostBaseURLResolver resolves URLs against a single BaseURL
type SingleHostBaseURLResolver struct {
	BaseURL string
}

// Resolve the base URL for a request
func (s SingleHostBaseURLResolver) Resolve(r *http.Request) string {

	baseURL := s.BaseURL

	if strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL[0 : len(baseURL)-1]
	}
	return baseURL
}

// FunctionAsHostBaseURLResolver resolves URLs using a function from the URL as a host
type FunctionAsHostBaseURLResolver struct {
	FunctionSuffix string
}

// Resolve the base URL for a request
func (f FunctionAsHostBaseURLResolver) Resolve(r *http.Request) string {
	svcName := getServiceName(r.URL.Path)

	const watchdogPort = 8080
	var suffix string
	if len(f.FunctionSuffix) > 0 {
		suffix = "." + f.FunctionSuffix
	}

	return fmt.Sprintf("http://%s%s:%d", svcName, suffix, watchdogPort)
}

// TransparentURLPathTransformer passes the requested URL path through untouched.
type TransparentURLPathTransformer struct {
}

// Transform returns the URL path unchanged.
func (f TransparentURLPathTransformer) Transform(r *http.Request) string {
	return r.URL.Path
}

// FunctionPrefixTrimmingURLPathTransformer removes the "/function/servicename/" prefix from the URL path.
type FunctionPrefixTrimmingURLPathTransformer struct {
}

// Transform removes the "/function/servicename/" prefix from the URL path.
func (f FunctionPrefixTrimmingURLPathTransformer) Transform(r *http.Request) string {
	ret := r.URL.Path

	if ret != "" {
		// When forwarding to a function, since the `/function/xyz` portion
		// of a path like `/function/xyz/rest/of/path` is only used or needed
		// by the Gateway, we want to trim it down to `/rest/of/path` for the
		// upstream request.  In the following regex, in the case of a match
		// the r.URL.Path will be at `0`, the function name at `1` and the
		// rest of the path (the part we are interested in) at `2`.
		matcher := functionMatcher.Copy()
		parts := matcher.FindStringSubmatch(ret)
		if len(parts) == hasPathCount {
			ret = parts[pathIndex]
		}
	}

	return ret
}
