// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package inttests

import (
	"net/http"
	"testing"
)

func TestCreate_ValidJson(t *testing.T) {
	reqBody := `{}`
	_, code, err := fireRequest("http://localhost:8080/system/functions", http.MethodPost, reqBody)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	if code != http.StatusOK {
		t.Errorf("Got HTTP code: %d, want %d\n", code, http.StatusBadRequest)
	}
}

func TestCreateBadFunctionNotJson(t *testing.T) {
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
