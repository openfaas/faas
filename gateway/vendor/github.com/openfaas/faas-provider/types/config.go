package types

import (
	"net/http"
	"time"
)

// FaaSHandlers provide handlers for OpenFaaS
type FaaSHandlers struct {
	FunctionReader http.HandlerFunc
	DeployHandler  http.HandlerFunc
	// FunctionProxy provides the function invocation proxy logic.  Use proxy.NewHandlerFunc to
	// use the standard OpenFaaS proxy implementation or provide completely custom proxy logic.
	FunctionProxy  http.HandlerFunc
	DeleteHandler  http.HandlerFunc
	ReplicaReader  http.HandlerFunc
	ReplicaUpdater http.HandlerFunc
	SecretHandler  http.HandlerFunc

	// Optional: Update an existing function
	UpdateHandler http.HandlerFunc
	Health        http.HandlerFunc
	InfoHandler   http.HandlerFunc
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
