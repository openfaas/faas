// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package inttests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

type PostFunctionRequest struct {
	Image      string `json:"image"`
	EnvProcess string `json:"envProcess"`
	Network    string `json:"network"`
	Service    string `json:"service"`
}

type DeleteFunctionRequest struct {
	FunctionName string `json:"functionName"`
}

func createFunction(request PostFunctionRequest) (string, int, error) {
	marshalled, _ := json.Marshal(request)
	return fireRequest("http://localhost:8080/system/functions", http.MethodPost, string(marshalled))
}

func deleteFunction(name string) (string, int, error) {
	marshalled, _ := json.Marshal(DeleteFunctionRequest{name})
	return fireRequest("http://localhost:8080/system/functions", http.MethodDelete, string(marshalled))
}

func TestCreate_ValidRequest(t *testing.T) {
	request := PostFunctionRequest{
		"functions/resizer",
		"",
		"func_functions",
		"test_resizer",
	}

	_, code, err := createFunction(request)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	if code != http.StatusOK {
		t.Errorf("Got HTTP code: %d, want %d\n", code, http.StatusOK)
		return
	}

	deleteFunction("test_resizer")
}

func TestCreate_InvalidImage(t *testing.T) {
	request := PostFunctionRequest{
		"a b c",
		"",
		"func_functions",
		"test_resizer",
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
	request := PostFunctionRequest{
		"functions/resizer",
		"",
		"non_existent_network",
		"test_resizer",
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
