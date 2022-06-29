package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
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
	BuildURL(function, namespace, healthPath string, directFunctions bool) string
}

// URLPathTransformer Transform the incoming URL path for upstream requests
type URLPathTransformer interface {
	Transform(r *http.Request) string
}

// SingleHostBaseURLResolver resolves URLs against a single BaseURL
type SingleHostBaseURLResolver struct {
	BaseURL string
}

func (s SingleHostBaseURLResolver) BuildURL(function, namespace, healthPath string, directFunctions bool) string {
	u, _ := url.Parse(s.BaseURL)

	base := fmt.Sprintf("/function/%s.%s/", function, namespace)

	if len(healthPath) > 0 {
		u.Path = path.Join(base, healthPath)
	} else {
		u.Path = base
	}

	return u.String()
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
	svcName := GetServiceName(r.URL.Path)

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

func (f FunctionAsHostBaseURLResolver) BuildURL(function, namespace, healthPath string, directFunctions bool) string {
	svcName := function

	const watchdogPort = 8080
	var suffix string

	if len(f.FunctionSuffix) > 0 {
		suffix = strings.Replace(f.FunctionSuffix, f.FunctionNamespace, namespace, 1)
	}

	u, _ := url.Parse(fmt.Sprintf("http://%s.%s:%d", svcName, suffix, watchdogPort))
	if len(healthPath) > 0 {
		u.Path = healthPath
	}

	return u.String()
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

func GetServiceName(urlValue string) string {
	var serviceName string
	forward := "/function/"
	if strings.HasPrefix(urlValue, forward) {
		// With a path like `/function/xyz/rest/of/path?q=a`, the service
		// name we wish to locate is just the `xyz` portion.  With a positive
		// match on the regex below, it will return a three-element slice.
		// The item at index `0` is the same as `urlValue`, at `1`
		// will be the service name we need, and at `2` the rest of the path.
		matcher := functionMatcher.Copy()
		matches := matcher.FindStringSubmatch(urlValue)
		if len(matches) == hasPathCount {
			serviceName = matches[nameIndex]
		}
	}
	return strings.Trim(serviceName, "/")
}
