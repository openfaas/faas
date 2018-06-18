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
