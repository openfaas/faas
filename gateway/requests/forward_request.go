package requests

import "fmt"
import "net/url"

// ForwardRequest for proxying incoming requests
type ForwardRequest struct {
	RawPath  string
	RawQuery string
	Method   string
}

// NewForwardRequest create a ForwardRequest
func NewForwardRequest(method string, url url.URL) ForwardRequest {
	return ForwardRequest{
		Method:   method,
		RawQuery: url.RawQuery,
		RawPath:  url.Path,
	}
}

// ToURL create formatted URL
func (f *ForwardRequest) ToURL(addr string, watchdogPort int) string {
	if len(f.RawQuery) > 0 {
		return fmt.Sprintf("http://%s:%d%s?%s", addr, watchdogPort, f.RawPath, f.RawQuery)
	}
	return fmt.Sprintf("http://%s:%d%s", addr, watchdogPort, f.RawPath)

}
