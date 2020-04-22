// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package scaling

import (
	"sync"
	"time"
)

// FunctionCacher queries functions and caches the results
type FunctionCacher interface {
	Set(functionName, namespace string, serviceQueryResponse ServiceQueryResponse)
	Get(functionName, namespace string) (ServiceQueryResponse, bool)
}

// FunctionCache provides a cache of Function replica counts
type FunctionCache struct {
	Cache  map[string]*FunctionMeta
	Expiry time.Duration
	Sync   sync.RWMutex
}

// NewFunctionCache creates a function cache to query function metadata
func NewFunctionCache(cacheExpiry time.Duration) FunctionCacher {
	return &FunctionCache{
		Cache:  make(map[string]*FunctionMeta),
		Expiry: cacheExpiry,
	}
}

// Set replica count for functionName
func (fc *FunctionCache) Set(functionName, namespace string, queryRes ServiceQueryResponse) {
	fc.Sync.Lock()
	defer fc.Sync.Unlock()

	if _, exists := fc.Cache[functionName+"."+namespace]; !exists {
		fc.Cache[functionName+"."+namespace] = &FunctionMeta{}
	}

	fc.Cache[functionName+"."+namespace].LastRefresh = time.Now()
	fc.Cache[functionName+"."+namespace].ServiceQueryResponse = queryRes
}

// Get replica count for functionName
func (fc *FunctionCache) Get(functionName, namespace string) (ServiceQueryResponse, bool) {
	queryRes := ServiceQueryResponse{
		AvailableReplicas: 0,
	}

	hit := false
	fc.Sync.RLock()
	defer fc.Sync.RUnlock()

	if val, exists := fc.Cache[functionName+"."+namespace]; exists {
		queryRes = val.ServiceQueryResponse
		hit = !val.Expired(fc.Expiry)
	}

	return queryRes, hit
}
