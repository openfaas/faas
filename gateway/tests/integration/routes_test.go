// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package inttests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Before running these tests do a Docker stack deploy.

func TestGet_Rejected(t *testing.T) {
	var reqBody string

	_, code := fireRequest("http://localhost:8080/function/func_echoit", http.MethodGet, reqBody)

	assert.Equal(t, http.StatusMethodNotAllowed, code)
}

func TestEchoIt_Post_Route_Handler_ForwardsClientHeaders(t *testing.T) {
	reqBody := "test message"
	headers := make(map[string]string, 0)
	headers["X-Api-Key"] = "123"

	body, code := fireRequestWithHeaders("http://localhost:8080/function/func_echoit", http.MethodPost, reqBody, headers)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, reqBody, body)
}

func TestEchoIt_Post_Route_Handler(t *testing.T) {
	reqBody := "test message"

	body, code := fireRequest("http://localhost:8080/function/func_echoit", http.MethodPost, reqBody)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, reqBody, body)
}

func TestEchoIt_Post_X_Header_Routing_Handler(t *testing.T) {
	reqBody := "test message"
	headers := make(map[string]string, 0)
	headers["X-Function"] = "func_echoit"

	body, code := fireRequestWithHeaders("http://localhost:8080/", http.MethodPost, reqBody, headers)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, reqBody, body)
}
