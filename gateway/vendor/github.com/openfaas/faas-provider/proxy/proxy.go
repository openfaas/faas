// Package proxy provides a default function invocation proxy method for OpenFaaS providers.
//
// The function proxy logic is used by the Gateway when `direct_functions` is set to false.
// This means that the provider will direct call the function and return the results.  This
// involves resolving the function by name and then copying the result into the original HTTP
// request.
//
// openfaas-provider has implemented a standard HTTP HandlerFunc that will handle setting
// timeout values, parsing the request path, and copying the request/response correctly.
// 		bootstrapHandlers := bootTypes.FaaSHandlers{
// 			FunctionProxy:  proxy.NewHandlerFunc(timeout, resolver),
// 			DeleteHandler:  handlers.MakeDeleteHandler(clientset),
// 			DeployHandler:  handlers.MakeDeployHandler(clientset),
// 			FunctionReader: handlers.MakeFunctionReader(clientset),
// 			ReplicaReader:  handlers.MakeReplicaReader(clientset),
// 			ReplicaUpdater: handlers.MakeReplicaUpdater(clientset),
// 			InfoHandler:    handlers.MakeInfoHandler(),
// 		}
//
// proxy.NewHandlerFunc is optional, but does simplify the logic of your provider.
package proxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
)

const (
	watchdogPort           = "8080"
	defaultContentType     = "text/plain"
	errMissingFunctionName = "Please provide a valid route /function/function_name."
)

// BaseURLResolver URL resolver for proxy requests
//
// The FaaS provider implementation is responsible for providing the resolver function implementation.
// BaseURLResolver.Resolve will receive the function name and should return the URL of the
// function service.
type BaseURLResolver interface {
	Resolve(functionName string) (url.URL, error)
}

// NewHandlerFunc creates a standard http.HandlerFunc to proxy function requests.
// The returned http.HandlerFunc will ensure:
//
// 	- proper proxy request timeouts
// 	- proxy requests for GET, POST, PATCH, PUT, and DELETE
// 	- path parsing including support for extracing the function name, sub-paths, and query paremeters
// 	- passing and setting the `X-Forwarded-Host` and `X-Forwarded-For` headers
// 	- logging errors and proxy request timing to stdout
//
// Note that this will panic if `resolver` is nil.
func NewHandlerFunc(timeout time.Duration, resolver BaseURLResolver) http.HandlerFunc {
	if resolver == nil {
		panic("NewHandlerFunc: empty proxy handler resolver, cannot be nil")
	}

	proxyClient := http.Client{
		// these Transport values ensure that the http Client will eventually timeout and prevents
		// infinite retries. The default http.Client configure these timeouts.  The specific
		// values tuned via performance testing/benchmarking
		//
		// Additional context can be found at
		// - https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
		// - https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 1 * time.Second,
			}).DialContext,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		switch r.Method {
		case http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodGet:

			proxyRequest(w, r, proxyClient, resolver)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// proxyRequest handles the actual resolution of and then request to the function service.
func proxyRequest(w http.ResponseWriter, originalReq *http.Request, proxyClient http.Client, resolver BaseURLResolver) {
	ctx := originalReq.Context()

	pathVars := mux.Vars(originalReq)
	functionName := pathVars["name"]
	if functionName == "" {
		writeError(w, http.StatusBadRequest, errMissingFunctionName)
		return
	}

	functionAddr, resolveErr := resolver.Resolve(functionName)
	if resolveErr != nil {
		// TODO: Should record the 404/not found error in Prometheus.
		log.Printf("resolver error: cannot find %s: %s\n", functionName, resolveErr.Error())
		writeError(w, http.StatusNotFound, "Cannot find service: %s.", functionName)
		return
	}

	proxyReq, err := buildProxyRequest(originalReq, functionAddr, pathVars["params"])
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to resolve service: %s.", functionName)
		return
	}
	defer proxyReq.Body.Close()

	start := time.Now()
	response, err := proxyClient.Do(proxyReq.WithContext(ctx))
	seconds := time.Since(start)

	if err != nil {
		log.Printf("error with proxy request to: %s, %s\n", proxyReq.URL.String(), err.Error())

		writeError(w, http.StatusInternalServerError, "Can't reach service for: %s.", functionName)
		return
	}

	log.Printf("%s took %f seconds\n", functionName, seconds.Seconds())

	clientHeader := w.Header()
	copyHeaders(clientHeader, &response.Header)
	w.Header().Set("Content-Type", getContentType(response.Header, originalReq.Header))

	w.WriteHeader(http.StatusOK)
	io.Copy(w, response.Body)
}

// buildProxyRequest creates a request object for the proxy request, it will ensure that
// the original request headers are preserved as well as setting openfaas system headers
func buildProxyRequest(originalReq *http.Request, baseURL url.URL, extraPath string) (*http.Request, error) {

	host := baseURL.Host
	if baseURL.Port() == "" {
		host = baseURL.Host + ":" + watchdogPort
	}

	url := url.URL{
		Scheme:   baseURL.Scheme,
		Host:     host,
		Path:     extraPath,
		RawQuery: originalReq.URL.RawQuery,
	}

	upstreamReq, err := http.NewRequest(originalReq.Method, url.String(), nil)
	if err != nil {
		return nil, err
	}
	copyHeaders(upstreamReq.Header, &originalReq.Header)

	if len(originalReq.Host) > 0 && upstreamReq.Header.Get("X-Forwarded-Host") == "" {
		upstreamReq.Header["X-Forwarded-Host"] = []string{originalReq.Host}
	}
	if upstreamReq.Header.Get("X-Forwarded-For") == "" {
		upstreamReq.Header["X-Forwarded-For"] = []string{originalReq.RemoteAddr}
	}

	if originalReq.Body != nil {
		upstreamReq.Body = originalReq.Body
	}

	return upstreamReq, nil
}

// copyHeaders clones the header values from the source into the destination.
func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		destination[k] = vClone
	}
}

// getContentType resolves the correct Content-Type for a proxied function.
func getContentType(request http.Header, proxyResponse http.Header) (headerContentType string) {
	responseHeader := proxyResponse.Get("Content-Type")
	requestHeader := request.Get("Content-Type")

	if len(responseHeader) > 0 {
		headerContentType = responseHeader
	} else if len(requestHeader) > 0 {
		headerContentType = requestHeader
	} else {
		headerContentType = defaultContentType
	}

	return headerContentType
}

// writeError sets the response status code and write formats the provided message as the
// response body
func writeError(w http.ResponseWriter, statusCode int, msg string, args ...interface{}) {
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf(msg, args...)))
}
