// Copyright (c) Alex Ellis, Eric Stoekl 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package inttests

import (
	"encoding/json"
	"github.com/openfaas/faas/gateway/requests"
	"net/http"
	"testing"
)

func listFunction(verbose bool) (string, int, error) {
	reqUrl := "http://localhost:8080/system/functions"
	if verbose {
		reqUrl += "?v=true"
	}
	return fireRequest(reqUrl, http.MethodGet, "")
}

func TestList(t *testing.T) {
	var results []requests.Function

	body, code, err := listFunction(false)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	expectedErrorCode := http.StatusOK
	if code != expectedErrorCode {
		t.Errorf("Got HTTP code: %d, want %d\n", code, expectedErrorCode)
		return
	}

	jsonErr := json.Unmarshal([]byte(body), &results)
	if jsonErr != nil {
		t.Errorf("Error parsing json: %s", jsonErr.Error())
	}

	for _, result := range results {
		if result.ReplicaCount != 0 {
			t.Errorf("Verbose mode should get replicaCount, instead got %v\n", result)
		}
	}
}

func TestList_verbose(t *testing.T) {
	var results []requests.Function

	body, code, err := listFunction(true)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	expectedErrorCode := http.StatusOK
	if code != expectedErrorCode {
		t.Errorf("Got HTTP code: %d, want %d\n", code, expectedErrorCode)
		return
	}

	jsonErr := json.Unmarshal([]byte(body), &results)
	if jsonErr != nil {
		t.Errorf("Error parsing json: %s", jsonErr.Error())
	}

	for _, result := range results {
		if result.ReplicaCount != 1 {
			t.Errorf("Verbose mode should get replicaCount, instead got %v\n", result)
		}
	}
}
