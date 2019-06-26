package inttests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/openfaas/faas/gateway/types"
)

func Test_InfoEndpoint_Returns_200(t *testing.T) {
	_, code, err := fireRequest("http://localhost:8080/system/info", http.MethodGet, "")

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	wantCode := http.StatusOK
	if code != wantCode {
		t.Errorf("status code, want: %d, got: %d", wantCode, code)
		t.Fail()
	}
}

func Test_InfoEndpoint_Returns_Gateway_Version_SHA_And_Message(t *testing.T) {
	body, _, err := fireRequest("http://localhost:8080/system/info", http.MethodGet, "")

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	gatewayInfo := &types.GatewayInfo{}
	err = json.Unmarshal([]byte(body), gatewayInfo)
	if err != nil {
		t.Errorf("Could not unmarshal gateway info, response body:%s, error:%s", body, err.Error())
		t.Fail()
	}

	if len(gatewayInfo.Version.SHA) != 40 {
		t.Errorf("length of SHA incorrect, want: %d, got: %d. Json body was %s", 40, len(gatewayInfo.Version.SHA), body)
	}

	if len(gatewayInfo.Version.CommitMessage) == 0 {
		t.Errorf("length of commit message should be greater than 0. Json body was %s", body)
	}
}

func Test_InfoEndpoint_Returns_Arch(t *testing.T) {
	body, _, err := fireRequest("http://localhost:8080/system/info", http.MethodGet, "")

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	gatewayInfo := &types.GatewayInfo{}
	err = json.Unmarshal([]byte(body), gatewayInfo)
	if err != nil {
		t.Errorf("Could not unmarshal gateway info, response body:%s, error:%s", body, err.Error())
		t.Fail()
	}

	if len(gatewayInfo.Arch) == 0 {
		t.Errorf("value of arch should be non-empty")
	}
}
