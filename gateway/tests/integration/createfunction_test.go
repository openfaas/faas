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
