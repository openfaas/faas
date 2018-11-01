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
	Sync   sync.Mutex
}

// Set replica count for functionName
func (fc *FunctionCache) Set(functionName string, serviceQueryResponse ServiceQueryResponse) {
	fc.Sync.Lock()
	defer fc.Sync.Unlock()

	if _, exists := fc.Cache[functionName]; !exists {
		fc.Cache[functionName] = &FunctionMeta{}
	}

	entry := fc.Cache[functionName]
	entry.LastRefresh = time.Now()
	entry.ServiceQueryResponse = serviceQueryResponse

}

// Get replica count for functionName
func (fc *FunctionCache) Get(functionName string) (ServiceQueryResponse, bool) {

	fc.Sync.Lock()
	defer fc.Sync.Unlock()

	replicas := ServiceQueryResponse{
		AvailableReplicas: 0,
	}

	hit := false
	if val, exists := fc.Cache[functionName]; exists {
		replicas = val.ServiceQueryResponse
		hit = !val.Expired(fc.Expiry)
	}

	return replicas, hit
}
