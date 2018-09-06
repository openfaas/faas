package types

import (
	"net/http"
	"time"
)

// FaaSHandlers provide handlers for OpenFaaS
type FaaSHandlers struct {
	FunctionReader http.HandlerFunc
	DeployHandler  http.HandlerFunc
	DeleteHandler  http.HandlerFunc
	ReplicaReader  http.HandlerFunc
	FunctionProxy  http.HandlerFunc
	ReplicaUpdater http.HandlerFunc

	// Optional: Update an existing function
	UpdateHandler http.HandlerFunc
	Health        http.HandlerFunc
	InfoHandler   http.HandlerFunc
}

// FaaSConfig set config for HTTP handlers
type FaaSConfig struct {
	TCPPort      *int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	EnableHealth bool
}
