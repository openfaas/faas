// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package inttests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelete_EmptyFunctionGivenFails(t *testing.T) {
	reqBody := `{"functionName":""}`
	_, code := fireRequest("http://localhost:8080/system/functions", http.MethodDelete, reqBody)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestDelete_NonExistingFunctionGives404(t *testing.T) {
	reqBody := `{"functionName":"does_not_exist"}`
	_, code := fireRequest("http://localhost:8080/system/functions", http.MethodDelete, reqBody)

	assert.Equal(t, http.StatusNotFound, code)
}
