package handler

import (
	"os"
	"strings"
	"testing"
)

func Test_GetClientID_ContainsHostname(t *testing.T) {
	c := DefaultNatsConfig{}

	val := c.GetClientID()

	hostname, _ := os.Hostname()
	if !strings.HasSuffix(val, hostname) {
		t.Errorf("GetClientID should contain hostname as suffix, got: %s", val)
		t.Fail()
	}
}

func TestCreateClientId(t *testing.T) {
	clientId := getClientId("computer-a")
	expected := "faas-publisher-computer-a"
	if clientId != expected {
		t.Logf("Expected client id `%s` actual `%s`\n", expected, clientId)
		t.Fail()
	}
}

func TestCreateClientIdWhenHostHasUnsupportedCharacters(t *testing.T) {
	clientId := getClientId("computer-a.acme.com")
	expected := "faas-publisher-computer-a_acme_com"
	if clientId != expected {
		t.Logf("Expected client id `%s` actual `%s`\n", expected, clientId)
		t.Fail()
	}
}


