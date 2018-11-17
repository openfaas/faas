package types

import "net/http"

// HandlerSet can be initialized with handlers for binding to mux
type HandlerSet struct {
	Proxy          http.HandlerFunc
	DeployFunction http.HandlerFunc
	DeleteFunction http.HandlerFunc
	ListFunctions  http.HandlerFunc
	Alert          http.HandlerFunc
	RoutelessProxy http.HandlerFunc
	UpdateFunction http.HandlerFunc

	// QueryFunction - queries the metdata for a function
	QueryFunction http.HandlerFunc

	// QueuedProxy - queue work and return synchronous response
	QueuedProxy http.HandlerFunc

	// AsyncReport - report a deferred execution result
	AsyncReport http.HandlerFunc

	// ScaleFunction allows a function to be scaled
	ScaleFunction http.HandlerFunc

	// InfoHandler provides version and build info
	InfoHandler http.HandlerFunc

	// SecretHandler allows secrets to be managed
	SecretHandler http.HandlerFunc
}
