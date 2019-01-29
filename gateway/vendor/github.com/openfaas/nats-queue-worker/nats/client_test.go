package nats

import (
	"testing"
)

func TestGetClientID(t *testing.T) {
	clientID := GetClientID("computer-a")
	want := "computer-a"
	if clientID != want {
		t.Logf("Want clientID: `%s`, but got: `%s`\n", want, clientID)
		t.Fail()
	}
}

func TestGetClientIDWhenHostHasUnsupportedCharacters(t *testing.T) {
	clientID := GetClientID("computer-a.acme.com")
	want := "computer-a_acme_com"
	if clientID != want {
		t.Logf("Want clientID: `%s`, but got: `%s`\n", want, clientID)
		t.Fail()
	}
}
