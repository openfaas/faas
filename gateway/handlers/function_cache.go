// Copyright (c) OpenFaaS Project. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"sync"
	"time"
)

// FunctionMeta holds the last refresh and any other
// meta-data needed for caching.
type FunctionMeta struct {
	LastRefresh time.Time
	Replicas    uint64
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
func (fc *FunctionCache) Set(functionName string, replicas uint64) {
	fc.Sync.Lock()

	if _, exists := fc.Cache[functionName]; !exists {
		fc.Cache[functionName] = &FunctionMeta{}
	}

	entry := fc.Cache[functionName]
	entry.LastRefresh = time.Now()
	entry.Replicas = replicas

	fc.Sync.Unlock()
}

// Get replica count for functionName
func (fc *FunctionCache) Get(functionName string) (uint64, bool) {
	replicas := uint64(0)
	hit := false
	fc.Sync.Lock()

	if val, exists := fc.Cache[functionName]; exists {
		replicas = val.Replicas
		hit = !val.Expired(fc.Expiry)
	}

	fc.Sync.Unlock()
	return replicas, hit
}
