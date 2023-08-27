// Copyright (c) Alex Ellis 1017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"testing"

	"github.com/openfaas/faas/gateway/scaling"
)

func TestDisabledScale(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(0)
	got := CalculateReplicas("firing", scaling.DefaultMinReplicas, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)
	if got != minReplicas {
		t.Logf("Expected not to scale, but replicas were: %d", got)
		t.Fail()
	}
}

func TestParameterEdge(t *testing.T) {
	minReplicas := uint64(0)
	scalingFactor := uint64(0)
	got := CalculateReplicas("firing", scaling.DefaultMinReplicas, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)
	if got != 0 {
		t.Log("Expected not to scale")
		t.Fail()
	}
}

func TestScaling_SameUpperLowerLimit(t *testing.T) {
	minReplicas := uint64(5)
	maxReplicas := uint64(5)
	scalingFactor := uint64(10)

	got := CalculateReplicas("firing", minReplicas, minReplicas, maxReplicas, scalingFactor)

	want := minReplicas
	if want != got {
		t.Logf("Replicas - want: %d, got: %d", want, got)
		t.Fail()
	}
}

func TestMaxScale(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(100)
	got := CalculateReplicas("firing", scaling.DefaultMinReplicas, scaling.DefaultMaxReplicas*2, minReplicas, scalingFactor)
	if got != scaling.DefaultMaxReplicas {
		t.Fatalf("want ceiling: %d, but got: %d", scaling.DefaultMaxReplicas, got)
	}
}

func TestInitialScale_From1_Factor10(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(10)
	got := CalculateReplicas("firing", scaling.DefaultMinReplicas, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)
	want := uint64(1)

	if got != want {
		t.Fatalf("want: %d, but got: %d", want, got)
	}
}

func TestScale_midrange_factor25(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(25)
	current := uint64(4)
	maxReplicas := uint64(scaling.DefaultMaxReplicas)

	got := CalculateReplicas("firing", current, maxReplicas, minReplicas, scalingFactor)
	want := uint64(5)
	if want != got {
		t.Fatalf("want: %d, but got: %d", want, got)
	}
}

func TestScale_Ceiling_IsDefaultMaxReplicas(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(10)
	current := uint64(scaling.DefaultMaxReplicas)

	got := CalculateReplicas("firing", current, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)
	if got != scaling.DefaultMaxReplicas {
		t.Fatalf("want: %d, but got: %d", scaling.DefaultMaxReplicas, got)
	}
}

func TestScaleCeilingReplicasOver(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(10)
	got := CalculateReplicas("firing", 19, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)

	if got != scaling.DefaultMaxReplicas {
		t.Fatalf("want: %d, but got: %d", scaling.DefaultMaxReplicas, got)
	}
}

func TestBackingOff(t *testing.T) {
	minReplicas := uint64(1)
	scalingFactor := uint64(10)
	got := CalculateReplicas("resolved", 8, scaling.DefaultMaxReplicas, minReplicas, scalingFactor)
	if got != 1 {
		t.Log("Expected backing off to 1 replica")
		t.Fail()
	}
}

func TestScaledUpFrom1(t *testing.T) {
	currentReplicas := uint64(1)
	maxReplicas := uint64(5)
	scalingFactor := uint64(30)
	got := CalculateReplicas("firing", currentReplicas, maxReplicas, scaling.DefaultMinReplicas, scalingFactor)
	if got <= currentReplicas {
		t.Log("Expected got > currentReplica")
		t.Fail()
	}
}

func TestScaledUpWithSmallParam(t *testing.T) {
	currentReplicas := uint64(1)
	maxReplicas := uint64(4)
	scalingFactor := uint64(1)
	got := CalculateReplicas("firing", currentReplicas, maxReplicas, scaling.DefaultMinReplicas, scalingFactor)
	if got <= currentReplicas {
		t.Log("Expected got > currentReplica")
		t.Fail()
	}
}
