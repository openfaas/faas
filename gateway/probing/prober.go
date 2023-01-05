// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package probing

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/openfaas/faas/gateway/pkg/middleware"
	"github.com/openfaas/faas/gateway/scaling"
	"github.com/openfaas/faas/gateway/types"
)

// NewFunctionProber create a new scaler with the specified
// ScalingConfig
func NewFunctionProber(functionQuery scaling.FunctionQuery, resolver middleware.BaseURLResolver) FunctionProber {
	// if directFunctions {
	return &FunctionHTTPProber{
		Query:    functionQuery,
		Resolver: resolver,
	}
}

// FunctionHTTPProber probes a function's health endpoint
type FunctionHTTPProber struct {
	Query           scaling.FunctionQuery
	Resolver        middleware.BaseURLResolver
	DirectFunctions bool
}

type FunctionNonProber struct {
}

func (f *FunctionNonProber) Probe(functionName, namespace string) FunctionProbeResult {
	return FunctionProbeResult{
		Found:     true,
		Available: true,
	}
}

type FunctionProber interface {
	Probe(functionName, namespace string) FunctionProbeResult
}

// FunctionProbeResult holds the result of scaling from zero
type FunctionProbeResult struct {
	Available bool
	Error     error
	Found     bool
	Duration  time.Duration
	Updated   time.Time
}

// Expired find out whether the cache item has expired with
// the given expiry duration from when it was stored.
func (res *FunctionProbeResult) Expired(expiry time.Duration) bool {
	return time.Now().After(res.Updated.Add(expiry))
}

// Scale scales a function from zero replicas to 1 or the value set in
// the minimum replicas metadata
func (f *FunctionHTTPProber) Probe(functionName, namespace string) FunctionProbeResult {
	start := time.Now()

	cachedResponse, _ := f.Query.Get(functionName, namespace)
	probePath := "/_/ready"

	if cachedResponse.Annotations != nil {
		if v, ok := (*cachedResponse.Annotations)["com.openfaas.ready.http.path"]; ok && len(v) > 0 {
			probePath = v
		}
	}

	maxCount := 10
	pollInterval := time.Millisecond * 50

	err := types.Retry(func(attempt int) error {
		u := f.Resolver.BuildURL(functionName, namespace, probePath, true)

		r, _ := http.NewRequest(http.MethodGet, u, nil)
		r.Header.Set("User-Agent", "com.openfaas.gateway/probe")

		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			return err
		}

		log.Printf("[Probe] %s => %d", u, resp.StatusCode)

		if resp.StatusCode == http.StatusOK {
			return nil
		}
		return fmt.Errorf("failed with status: %s", resp.Status)
	}, "Probe", maxCount, pollInterval)

	if err != nil {
		return FunctionProbeResult{
			Error:     err,
			Available: false,
			Found:     true,
			Duration:  time.Since(start),
			Updated:   time.Now(),
		}
	}

	return FunctionProbeResult{
		Error:     nil,
		Available: true,
		Found:     true,
		Duration:  time.Since(start),
		Updated:   time.Now(),
	}
}
