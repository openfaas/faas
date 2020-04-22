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

	"github.com/openfaas/faas/gateway/pkg/middleware"
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
func MakeForwardingProxyHandler(proxy *types.HTTPClientReverseProxy,
	notifiers []HTTPNotifier,
	baseURLResolver BaseURLResolver,
	urlPathTransformer URLPathTransformer,
	serviceAuthInjector middleware.AuthInjector) http.HandlerFunc {

	writeRequestURI := false
	if _, exists := os.LookupEnv("write_request_uri"); exists {
		writeRequestURI = exists
	}

	return func(w http.ResponseWriter, r *http.Request) {
		baseURL := baseURLResolver.Resolve(r)
		originalURL := r.URL.String()
		requestURL := urlPathTransformer.Transform(r)

		for _, notifier := range notifiers {
			notifier.Notify(r.Method, requestURL, originalURL, http.StatusProcessing, "started", time.Second*0)
		}

		start := time.Now()

		statusCode, err := forwardRequest(w, r, proxy.Client, baseURL, requestURL, proxy.Timeout, writeRequestURI, serviceAuthInjector)

		seconds := time.Since(start)
		if err != nil {
			log.Printf("error with upstream request to: %s, %s\n", requestURL, err.Error())
		}

		for _, notifier := range notifiers {
			notifier.Notify(r.Method, requestURL, originalURL, statusCode, "completed", seconds)
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
	deleteHeaders(&upstreamReq.Header, &hopHeaders)

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

func forwardRequest(w http.ResponseWriter,
	r *http.Request,
	proxyClient *http.Client,
	baseURL string,
	requestURL string,
	timeout time.Duration,
	writeRequestURI bool,
	serviceAuthInjector middleware.AuthInjector) (int, error) {

	upstreamReq := buildUpstreamRequest(r, baseURL, requestURL)
	if upstreamReq.Body != nil {
		defer upstreamReq.Body.Close()
	}

	if serviceAuthInjector != nil {
		serviceAuthInjector.Inject(upstreamReq)
	}

	if writeRequestURI {
		log.Printf("forwardRequest: %s %s\n", upstreamReq.Host, upstreamReq.URL.String())
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
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

func deleteHeaders(target *http.Header, exclude *[]string) {
	for _, h := range *exclude {
		target.Del(h)
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
	FunctionSuffix    string
	FunctionNamespace string
}

// Resolve the base URL for a request
func (f FunctionAsHostBaseURLResolver) Resolve(r *http.Request) string {
	svcName := getServiceName(r.URL.Path)

	const watchdogPort = 8080
	var suffix string

	if len(f.FunctionSuffix) > 0 {
		if index := strings.LastIndex(svcName, "."); index > -1 && len(svcName) > index+1 {
			suffix = strings.Replace(f.FunctionSuffix, f.FunctionNamespace, "", -1)
		} else {
			suffix = "." + f.FunctionSuffix
		}
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

// Hop-by-hop headers. These are removed when sent to the backend.
// As of RFC 7230, hop-by-hop headers are required to appear in the
// Connection header field. These are the headers defined by the
// obsoleted RFC 2616 (section 13.5.1) and are used for backward
// compatibility.
// Copied from: https://golang.org/src/net/http/httputil/reverseproxy.go
var hopHeaders = []string{
	"Connection",
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}
