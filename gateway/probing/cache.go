// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package probing

import (
	"fmt"
	"sync"
	"time"
)

// ProbeCacher queries functions and caches the results
type ProbeCacher interface {
	Set(functionName, namespace string, result *FunctionProbeResult)
	Get(functionName, namespace string) (result *FunctionProbeResult, hit bool)
}

// ProbeCache provides a cache of Probe replica counts
type ProbeCache struct {
	Cache  map[string]*FunctionProbeResult
	Expiry time.Duration
	Sync   sync.RWMutex
}

// NewProbeCache creates a function cache to query function metadata
func NewProbeCache(cacheExpiry time.Duration) ProbeCacher {
	return &ProbeCache{
		Cache:  make(map[string]*FunctionProbeResult),
		Expiry: cacheExpiry,
	}
}

// Set replica count for functionName
func (fc *ProbeCache) Set(functionName, namespace string, result *FunctionProbeResult) {
	fc.Sync.Lock()
	defer fc.Sync.Unlock()

	fc.Cache[functionName+"."+namespace] = result
}

func (fc *ProbeCache) Get(functionName, namespace string) (*FunctionProbeResult, bool) {

	result := &FunctionProbeResult{
		Available: false,
		Error:     fmt.Errorf("unavailable in cache"),
	}

	hit := false
	fc.Sync.RLock()
	defer fc.Sync.RUnlock()

	if val, exists := fc.Cache[functionName+"."+namespace]; exists {
		hit = val.Expired(fc.Expiry) == false
		result = val
	}

	return result, hit
}
