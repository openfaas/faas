// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"
	"time"

	fhttputil "github.com/openfaas/faas-provider/httputil"
	"github.com/openfaas/faas/gateway/pkg/middleware"
	"github.com/openfaas/faas/gateway/types"
)

// MakeForwardingProxyHandler create a handler which forwards HTTP requests
func MakeForwardingProxyHandler(proxy *types.HTTPClientReverseProxy,
	notifiers []HTTPNotifier,
	baseURLResolver middleware.BaseURLResolver,
	urlPathTransformer middleware.URLPathTransformer,
	serviceAuthInjector middleware.AuthInjector) http.HandlerFunc {

	writeRequestURI := false
	if _, exists := os.LookupEnv("write_request_uri"); exists {
		writeRequestURI = exists
	}

	reverseProxy := makeRewriteProxy(baseURLResolver, urlPathTransformer)

	return func(w http.ResponseWriter, r *http.Request) {

		baseURL := baseURLResolver.Resolve(r)
		originalURL := r.URL.String()
		requestURL := urlPathTransformer.Transform(r)

		for _, notifier := range notifiers {
			notifier.Notify(r.Method, requestURL, originalURL, http.StatusProcessing, "started", time.Second*0)
		}

		start := time.Now()

		statusCode, err := forwardRequest(w, r, proxy.Client, baseURL, requestURL, proxy.Timeout, writeRequestURI, serviceAuthInjector, reverseProxy)
		if err != nil {
			log.Printf("error with upstream request to: %s, %s\n", requestURL, err.Error())
		}

		seconds := time.Since(start)

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
	serviceAuthInjector middleware.AuthInjector,
	reverseProxy *httputil.ReverseProxy) (int, error) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	upstreamReq := buildUpstreamRequest(r, baseURL, requestURL)

	if serviceAuthInjector != nil {
		serviceAuthInjector.Inject(upstreamReq)
	}

	if writeRequestURI {
		log.Printf("forwardRequest: %s %s\n", upstreamReq.Host, upstreamReq.URL.String())
	}

	if strings.HasPrefix(r.Header.Get("Accept"), "text/event-stream") {

		return handleEventStream(w, r, reverseProxy, upstreamReq, timeout)
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	res, err := proxyClient.Do(upstreamReq.WithContext(ctx))
	if err != nil {
		badStatus := http.StatusBadGateway
		w.WriteHeader(badStatus)
		return badStatus, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	copyHeaders(w.Header(), &res.Header)

	w.WriteHeader(res.StatusCode)

	if res.Body != nil {
		io.Copy(w, res.Body)
	}

	return res.StatusCode, nil
}

func handleEventStream(w http.ResponseWriter, r *http.Request, reverseProxy *httputil.ReverseProxy, upstreamReq *http.Request, timeout time.Duration) (int, error) {
	ww := fhttputil.NewHttpWriteInterceptor(w)

	ctx, cancel := context.WithTimeoutCause(r.Context(), timeout, http.ErrHandlerTimeout)
	defer cancel()

	r = r.WithContext(ctx)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			if r := recover(); r != nil {
				if errors.Is(r.(error), http.ErrAbortHandler) {
					log.Printf("Aborted [%s] for: %s", upstreamReq.Method, upstreamReq.URL.Path)

				} else {
					log.Printf("Recovered from panic in reverseproxy: %v", r)
				}
			}
		}()

		reverseProxy.ServeHTTP(ww, r)
	}()

	wg.Wait()

	return ww.Status(), nil
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

func makeRewriteProxy(baseURLResolver middleware.BaseURLResolver, urlPathTransformer middleware.URLPathTransformer) *httputil.ReverseProxy {

	return &httputil.ReverseProxy{

		ErrorLog: log.New(io.Discard, "proxy:", 0),
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
		},
		Director: func(r *http.Request) {

			baseURL := baseURLResolver.Resolve(r)
			baseURLu, _ := r.URL.Parse(baseURL)

			requestURL := urlPathTransformer.Transform(r)

			r.URL.Scheme = "http"
			r.URL.Path = requestURL
			r.URL.Host = baseURLu.Host
		},
	}
}
