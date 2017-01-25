package tests

import (
	"testing"

	"io/ioutil"

	"github.com/alexellis/faas/gateway/handlers"
	"github.com/alexellis/faas/gateway/requests"
)

func TestIsAlexa(t *testing.T) {
	requestBody, _ := ioutil.ReadFile("./alexhostname_request.json")
	var result requests.AlexaRequestBody

	result = handlers.IsAlexa(requestBody)

	if len(result.Session.Application.ApplicationId) == 0 {
		t.Fail()
	}
	if len(result.Session.SessionId) == 0 {
		t.Fail()
	}
	if len(result.Request.Intent.Name) == 0 {
		t.Fail()
	}
}
