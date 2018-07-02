package types

import (
	"bytes"
	"net/http"
)

// GatewayInfo provides information about the gateway and it's connected components
type GatewayInfo struct {
	Provider *ProviderInfo `json:"provider"`
	Version  *VersionInfo  `json:"version"`
}

// ProviderInfo provides information about the configured provider
type ProviderInfo struct {
	Name          string       `json:"provider"`
	Version       *VersionInfo `json:"version"`
	Orchestration string       `json:"orchestration"`
}

// VersionInfo provides the commit message, sha and release version number
type VersionInfo struct {
	CommitMessage string `json:"commit_message,omitempty"`
	SHA           string `json:"sha"`
	Release       string `json:"release"`
}

// StringResponseWriter captures the handlers HTTP response in a buffer
type StringResponseWriter struct {
	body       *bytes.Buffer
	headerCode int
	header     http.Header
}

// NewStringResponseWriter create a new StringResponseWriter
func NewStringResponseWriter() *StringResponseWriter {
	return &StringResponseWriter{body: &bytes.Buffer{}, header: make(http.Header)}
}

// Header capture the Header information
func (s StringResponseWriter) Header() http.Header {
	return s.header
}

// Write captures the response data
func (s StringResponseWriter) Write(data []byte) (int, error) {
	return s.body.Write(data)
}

// WriteHeader captures the status code of the response
func (s StringResponseWriter) WriteHeader(statusCode int) {
	s.headerCode = statusCode
}

// Body returns the response body bytes
func (s StringResponseWriter) Body() []byte {
	return s.body.Bytes()
}
