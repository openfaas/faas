// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package scaling

import (
	"sync"
	"time"
)

// FunctionMeta holds the last refresh and any other
// meta-data needed for caching.
type FunctionMeta struct {
	LastRefresh          time.Time
	ServiceQueryResponse ServiceQueryResponse
}

// Expired find out whether the cache item has expired with
// the given expiry duration from when it was stored.
func (fm *FunctionMeta) Expired(expiry time.Duration) bool {
	return time.Now().After(fm.LastRefresh.Add(expiry))
}

// FunctionCache provides a cache of Function replica counts
type FunctionCache struct {
	Cache  map[string]*FunctionMeta
	Expiry time.Duration
	Sync   sync.RWMutex
}

// Set replica count for functionName
func (fc *FunctionCache) Set(functionName, namespace string, serviceQueryResponse ServiceQueryResponse) {
	fc.Sync.Lock()
	defer fc.Sync.Unlock()

	if _, exists := fc.Cache[functionName+"."+namespace]; !exists {
		fc.Cache[functionName+"."+namespace] = &FunctionMeta{}
	}

	fc.Cache[functionName+"."+namespace].LastRefresh = time.Now()
	fc.Cache[functionName+"."+namespace].ServiceQueryResponse = serviceQueryResponse
	// entry.LastRefresh = time.Now()
	// entry.ServiceQueryResponse = serviceQueryResponse
}

// Get replica count for functionName
func (fc *FunctionCache) Get(functionName, namespace string) (ServiceQueryResponse, bool) {
	replicas := ServiceQueryResponse{
		AvailableReplicas: 0,
	}

	hit := false
	fc.Sync.RLock()
	defer fc.Sync.RUnlock()

	if val, exists := fc.Cache[functionName+"."+namespace]; exists {
		replicas = val.ServiceQueryResponse
		hit = !val.Expired(fc.Expiry)
	}

	return replicas, hit
}
