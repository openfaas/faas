// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package inttests

import (
	"net/http"
	"testing"
)

func TestDelete_EmptyFunctionGivenFails(t *testing.T) {
	reqBody := `{"functionName":""}`
	_, code, err := fireRequest("http://localhost:8080/system/functions", http.MethodDelete, reqBody)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	if code != http.StatusBadRequest {
		t.Errorf("Got HTTP code: %d, want %d\n", code, http.StatusBadRequest)
	}
}

func TestDelete_NonExistingFunctionGives404(t *testing.T) {
	reqBody := `{"functionName":"does_not_exist"}`
	_, code, err := fireRequest("http://localhost:8080/system/functions", http.MethodDelete, reqBody)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	if code != http.StatusNotFound {
		t.Errorf("Got HTTP code: %d, want %d\n", code, http.StatusNotFound)
	}
}
