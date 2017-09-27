// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package inttests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/stretchr/testify/assert"
)

func createFunction(request requests.CreateFunctionRequest) (string, int) {
	marshalled, _ := json.Marshal(request)
	return fireRequest("http://localhost:8080/system/functions", http.MethodPost, string(marshalled))
}

func deleteFunction(name string) (string, int) {
	marshalled, _ := json.Marshal(requests.DeleteFunctionRequest{FunctionName: name})
	return fireRequest("http://localhost:8080/system/functions", http.MethodDelete, string(marshalled))
}

func TestCreate_ValidRequest(t *testing.T) {
	request := requests.CreateFunctionRequest{
		Service:    "test_resizer",
		Image:      "functions/resizer",
		Network:    "func_functions",
		EnvProcess: "",
	}

	_, code := createFunction(request)

	assert.Equal(t, http.StatusOK, code)

	deleteFunction("test_resizer")
}

func TestCreate_InvalidImage(t *testing.T) {
	request := requests.CreateFunctionRequest{
		Service:    "test_resizer",
		Image:      "a b c",
		Network:    "func_functions",
		EnvProcess: "",
	}

	body, code := createFunction(request)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Contains(t, body, "is not a valid repository/tag")
}

func TestCreate_InvalidNetwork(t *testing.T) {
	request := requests.CreateFunctionRequest{
		Service:    "test_resizer",
		Image:      "functions/resizer",
		Network:    "non_existent_network",
		EnvProcess: "",
	}

	body, code := createFunction(request)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Contains(t, body, "network non_existent_network not found")
}

func TestCreate_InvalidJson(t *testing.T) {
	reqBody := `not json`
	_, code := fireRequest("http://localhost:8080/system/functions", http.MethodPost, reqBody)

	assert.Equal(t, http.StatusBadRequest, code)
}
