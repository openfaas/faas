package tests

import (
	"testing"

	"github.com/openfaas/faas/gateway/handlers"
)

func Test_ParseMemory(t *testing.T) {
	value := "512 m"

	val, err := handlers.ParseMemory(value)
	if err != nil {
		t.Error(err)
	}

	if val != 1024*1024*512 {
		t.Errorf("want: %d got: %d", 1024, val)
	}
}
