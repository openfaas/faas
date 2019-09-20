// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package scaling

import (
	"testing"
	"time"
)

func Test_LastRefreshSet(t *testing.T) {
	before := time.Now()

	fnName := "echo"
	namespace := ""
	cache := FunctionCache{
		Cache:  make(map[string]*FunctionMeta),
		Expiry: time.Millisecond * 1,
	}

	if cache.Cache == nil {
		t.Errorf("Expected cache map to be initialized")
		t.Fail()
	}

	cache.Set(fnName, "", ServiceQueryResponse{AvailableReplicas: 1})

	if _, exists := cache.Cache[fnName+"."+namespace]; !exists {
		t.Errorf("Expected entry to exist after setting %s", fnName)
		t.Fail()
	}

	if cache.Cache[fnName+"."+namespace].LastRefresh.Before(before) {
		t.Errorf("Expected LastRefresh for function to have been after start of test")
		t.Fail()
	}
}

func Test_CacheExpiresIn1MS(t *testing.T) {
	fnName := "echo"
	namespace := ""
	cache := FunctionCache{
		Cache:  make(map[string]*FunctionMeta),
		Expiry: time.Millisecond * 1,
	}

	cache.Set(fnName, namespace, ServiceQueryResponse{AvailableReplicas: 1})
	time.Sleep(time.Millisecond * 2)

	_, hit := cache.Get(fnName, namespace)

	wantHit := false

	if hit != wantHit {
		t.Errorf("hit, want: %v, got %v", wantHit, hit)
	}
}

func Test_CacheGivesHitWithLongExpiry(t *testing.T) {
	fnName := "echo"
	namespace := ""

	cache := FunctionCache{
		Cache:  make(map[string]*FunctionMeta),
		Expiry: time.Millisecond * 500,
	}

	cache.Set(fnName, namespace, ServiceQueryResponse{AvailableReplicas: 1})
	_, hit := cache.Get(fnName, namespace)

	wantHit := true

	if hit != wantHit {
		t.Errorf("hit, want: %v, got %v", wantHit, hit)
	}
}

func Test_CacheFunctionExists(t *testing.T) {
	fnName := "echo"
	namespace := ""

	cache := FunctionCache{
		Cache:  make(map[string]*FunctionMeta),
		Expiry: time.Millisecond * 10,
	}

	cache.Set(fnName, namespace, ServiceQueryResponse{AvailableReplicas: 1})
	time.Sleep(time.Millisecond * 2)

	_, hit := cache.Get(fnName, namespace)

	wantHit := true

	if hit != wantHit {
		t.Errorf("hit, want: %v, got %v", wantHit, hit)
	}
}

func Test_CacheFunctionExistsWithNamespace(t *testing.T) {
	fnName := "echo"
	namespace := "openfaas-fn"

	cache := FunctionCache{
		Cache:  make(map[string]*FunctionMeta),
		Expiry: time.Millisecond * 10,
	}

	cache.Set(fnName, namespace, ServiceQueryResponse{AvailableReplicas: 1})

	_, hit := cache.Get(fnName, namespace)

	wantHit := true

	if hit != wantHit {
		t.Errorf("hit, want: %v, got %v", wantHit, hit)
	}
}

func Test_CacheFunctionNotExist(t *testing.T) {
	fnName := "echo"
	testName := "burt"
	namespace := ""

	cache := FunctionCache{
		Cache:  make(map[string]*FunctionMeta),
		Expiry: time.Millisecond * 10,
	}

	cache.Set(fnName, namespace, ServiceQueryResponse{AvailableReplicas: 1})
	time.Sleep(time.Millisecond * 2)

	_, hit := cache.Get(testName, namespace)

	wantHit := false

	if hit != wantHit {
		t.Errorf("hit, want: %v, got %v", wantHit, hit)
	}
}
