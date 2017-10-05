// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package inttests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/openfaas/faas/gateway/requests"
)

func createFunction(request requests.CreateFunctionRequest) (string, int, error) {
	marshalled, _ := json.Marshal(request)
	return fireRequest("http://localhost:8080/system/functions", http.MethodPost, string(marshalled))
}

func deleteFunction(name string) (string, int, error) {
	marshalled, _ := json.Marshal(requests.DeleteFunctionRequest{name})
	return fireRequest("http://localhost:8080/system/functions", http.MethodDelete, string(marshalled))
}

func TestCreate_ValidRequest(t *testing.T) {
	request := requests.CreateFunctionRequest{
		Service:    "test_resizer",
		Image:      "functions/resizer",
		Network:    "func_functions",
		EnvProcess: "",
	}

	_, code, err := createFunction(request)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	expectedErrorCode := http.StatusOK
	if code != expectedErrorCode {
		t.Errorf("Got HTTP code: %d, want %d\n", code, expectedErrorCode)
		return
	}

	deleteFunction("test_resizer")
}

func TestCreate_InvalidImage(t *testing.T) {
	request := requests.CreateFunctionRequest{
		Service:    "test_resizer",
		Image:      "a b c",
		Network:    "func_functions",
		EnvProcess: "",
	}

	body, code, err := createFunction(request)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	expectedErrorCode := http.StatusBadRequest
	if code != expectedErrorCode {
		t.Errorf("Got HTTP code: %d, want %d\n", code, expectedErrorCode)
		return
	}

	expectedErrorSlice := "is not a valid repository/tag"
	if !strings.Contains(body, expectedErrorSlice) {
		t.Errorf("Error message %s does not contain: %s\n", body, expectedErrorSlice)
		return
	}
}

func TestCreate_InvalidNetwork(t *testing.T) {
	request := requests.CreateFunctionRequest{
		Service:    "test_resizer",
		Image:      "functions/resizer",
		Network:    "non_existent_network",
		EnvProcess: "",
	}

	body, code, err := createFunction(request)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	expectedErrorCode := http.StatusBadRequest
	if code != expectedErrorCode {
		t.Errorf("Got HTTP code: %d, want %d\n", code, expectedErrorCode)
		return
	}

	expectedErrorSlice := "network non_existent_network not found"
	if !strings.Contains(body, expectedErrorSlice) {
		t.Errorf("Error message %s does not contain: %s\n", body, expectedErrorSlice)
		return
	}
}

func TestCreate_InvalidJson(t *testing.T) {
	reqBody := `not json`
	_, code, err := fireRequest("http://localhost:8080/system/functions", http.MethodPost, reqBody)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	if code != http.StatusBadRequest {
		t.Errorf("Got HTTP code: %d, want %d\n", code, http.StatusBadRequest)
	}
}
