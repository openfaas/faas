// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package tests

import (
	"testing"

	"github.com/openfaas/faas/gateway/handlers"
)

func TestScale1to5(t *testing.T) {
	newReplicas := handlers.CalculateReplicas("firing", 1, 20)
	if newReplicas != 5 {
		t.Log("Expected increment in blocks of 5 from 1 to 5")
		t.Fail()
	}
}

func TestScale5to10(t *testing.T) {
	newReplicas := handlers.CalculateReplicas("firing", 5, 20)
	if newReplicas != 10 {
		t.Log("Expected increment in blocks of 5 from 5 to 10")
		t.Fail()
	}
}

func TestScaleCeilingOf20Replicas_Noaction(t *testing.T) {
	newReplicas := handlers.CalculateReplicas("firing", 20, 20)
	if newReplicas != 20 {
		t.Log("Expected ceiling of 20 replicas")
		t.Fail()
	}
}

func TestScaleCeilingOf20Replicas(t *testing.T) {
	newReplicas := handlers.CalculateReplicas("firing", 19, 20)
	if newReplicas != 20 {
		t.Log("Expected ceiling of 20 replicas")
		t.Fail()
	}
}

func TestBackingOff10to1(t *testing.T) {
	newReplicas := handlers.CalculateReplicas("resolved", 10, 20)
	if newReplicas != 1 {
		t.Log("Expected backing off to 1 replica")
		t.Fail()
	}
}
