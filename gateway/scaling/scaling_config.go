package scaling

import (
	"time"
)

// ScalingConfig for scaling behaviours
type ScalingConfig struct {
	// MaxPollCount attempts to query a function before giving up
	MaxPollCount uint

	// FunctionPollInterval delay or interval between polling a function's readiness status
	FunctionPollInterval time.Duration

	// CacheExpiry life-time for a cache entry before considering invalid
	CacheExpiry time.Duration

	// ServiceQuery queries available/ready replicas for function
	ServiceQuery ServiceQuery
}
