// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"testing"

	"github.com/openfaas/faas/gateway/scaling"
)

func TestDisabledScale(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(0)
	newReplicas := CalculateReplicas("firing", scaling.DefaultMinReplicas, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)
	if newReplicas != minReplicas {
		t.Logf("Expected not to scale, but replicas were: %d", newReplicas)
		t.Fail()
	}
}

func TestParameterEdge(t *testing.T) {
	minReplicas := uint64(0)
	scalingFactor := uint64(0)
	newReplicas := CalculateReplicas("firing", scaling.DefaultMinReplicas, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)
	if newReplicas != 0 {
		t.Log("Expected not to scale")
		t.Fail()
	}
}

func TestScalingWithSameUpperLowerLimit(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(20)
	//	status string, currentReplicas uint64, maxReplicas uint64, minReplicas uint64, scalingFactor uint64)
	newReplicas := CalculateReplicas("firing", minReplicas, minReplicas, minReplicas, scalingFactor)
	if newReplicas != 1 {
		t.Logf("Replicas - want: %d, got: %d", minReplicas, newReplicas)
		t.Fail()
	}
}

func TestMaxScale(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(100)
	newReplicas := CalculateReplicas("firing", scaling.DefaultMinReplicas, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)
	if newReplicas != 20 {
		t.Log("Expected ceiling of 20 replicas")
		t.Fail()
	}
}

func TestInitialScale(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(20)
	newReplicas := CalculateReplicas("firing", scaling.DefaultMinReplicas, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)
	if newReplicas != 4 {
		t.Log("Expected the increment to equal 4")
		t.Fail()
	}
}

func TestScale(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(20)
	newReplicas := CalculateReplicas("firing", 4, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)
	if newReplicas != 8 {
		t.Log("Expected newReplicas to equal 8")
		t.Fail()
	}
}

func TestScaleCeiling(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(20)
	newReplicas := CalculateReplicas("firing", 20, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)
	if newReplicas != 20 {
		t.Log("Expected ceiling of 20 replicas")
		t.Fail()
	}
}

func TestScaleCeilingEdge(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(20)
	newReplicas := CalculateReplicas("firing", 19, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)
	if newReplicas != 20 {
		t.Log("Expected ceiling of 20 replicas")
		t.Fail()
	}
}

func TestBackingOff(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(20)
	newReplicas := CalculateReplicas("resolved", 8, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)
	if newReplicas != 1 {
		t.Log("Expected backing off to 1 replica")
		t.Fail()
	}
}
