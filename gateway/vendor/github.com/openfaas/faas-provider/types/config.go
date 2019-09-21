package types

import (
	"net/http"
	"time"
)

// FaaSHandlers provide handlers for OpenFaaS
type FaaSHandlers struct {
	// FunctionProxy provides the function invocation proxy logic.  Use proxy.NewHandlerFunc to
	// use the standard OpenFaaS proxy implementation or provide completely custom proxy logic.
	FunctionProxy http.HandlerFunc

	FunctionReader http.HandlerFunc
	DeployHandler  http.HandlerFunc

	DeleteHandler  http.HandlerFunc
	ReplicaReader  http.HandlerFunc
	ReplicaUpdater http.HandlerFunc
	SecretHandler  http.HandlerFunc
	// LogHandler provides streaming json logs of functions
	LogHandler http.HandlerFunc

	// UpdateHandler an existing function/service
	UpdateHandler        http.HandlerFunc
	HealthHandler        http.HandlerFunc
	InfoHandler          http.HandlerFunc
	ListNamespaceHandler http.HandlerFunc
}

// FaaSConfig set config for HTTP handlers
type FaaSConfig struct {
	TCPPort         *int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	EnableHealth    bool
	EnableBasicAuth bool
	SecretMountPath string
}
