package types

import "net/http"

// HandlerSet can be initialized with handlers for binding to mux
type HandlerSet struct {
	// Proxy invokes a function
	Proxy http.HandlerFunc

	// DeployFunction deploys a new function that isn't already deployed
	DeployFunction http.HandlerFunc

	// DeleteFunction deletes a function that is already deployed
	DeleteFunction http.HandlerFunc

	// ListFunctions lists all deployed functions in a namespace
	ListFunctions http.HandlerFunc

	// Alert handles alerts triggered from AlertManager
	Alert http.HandlerFunc

	// UpdateFunction updates an existing function
	UpdateFunction http.HandlerFunc

	// FunctionStatus returns the status of an already deployed function
	FunctionStatus http.HandlerFunc

	// QueuedProxy queue work and return synchronous response
	QueuedProxy http.HandlerFunc

	// ScaleFunction enables a function to be scaled
	ScaleFunction http.HandlerFunc

	// InfoHandler provides version and build info
	InfoHandler http.HandlerFunc

	// SecretHandler enables secrets to be managed
	SecretHandler http.HandlerFunc

	// LogProxyHandler enables streaming of logs for functions
	LogProxyHandler http.HandlerFunc

	// NamespaceListerHandler lists namespaces
	NamespaceListerHandler http.HandlerFunc

	NamespaceMutatorHandler http.HandlerFunc
}
