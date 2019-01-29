package main

import (
	"os"
	"testing"
	"time"
)

func Test_ReadConfig(t *testing.T) {

	readConfig := ReadConfig{}

	os.Setenv("faas_nats_address", "test_nats")
	os.Setenv("faas_gateway_address", "test_gatewayaddr")
	os.Setenv("faas_function_suffix", "test_suffix")
	os.Setenv("faas_print_body", "true")
	os.Setenv("write_debug", "true")
	os.Setenv("max_inflight", "10")
	os.Setenv("ack_wait", "10ms")

	config := readConfig.Read()

	expected := "test_nats"
	if config.NatsAddress != expected {
		t.Logf("Expected NatsAddress `%s` actual `%s`\n", expected, config.NatsAddress)
		t.Fail()
	}

	expected = "test_gatewayaddr"
	if config.GatewayAddress != expected {
		t.Logf("Expected GatewayAddress `%s` actual `%s`\n", expected, config.GatewayAddress)
		t.Fail()
	}

	expected = "test_suffix"
	if config.FunctionSuffix != expected {
		t.Logf("Expected FunctionSuffix `%s` actual `%s`\n", expected, config.FunctionSuffix)
		t.Fail()
	}

	if config.DebugPrintBody != true {
		t.Logf("Expected DebugPrintBody `%v` actual `%v`\n", true, config.DebugPrintBody)
		t.Fail()
	}

	if config.WriteDebug != true {
		t.Logf("Expected WriteDebug `%v` actual `%v`\n", true, config.WriteDebug)
		t.Fail()
	}

	expectedMaxInflight := 10
	if config.MaxInflight != expectedMaxInflight {
		t.Logf("Expected maxInflight `%v` actual `%v`\n", expectedMaxInflight, config.MaxInflight)
		t.Fail()
	}

	expectedAckWait := time.Millisecond * 10
	if config.AckWait != expectedAckWait {
		t.Logf("Expected maxInflight `%v` actual `%v`\n", expectedAckWait, config.AckWait)
		t.Fail()
	}

	os.Unsetenv("max_inflight")
	os.Unsetenv("ack_wait")

	config = readConfig.Read()

	expectedMaxInflight = 1
	if config.MaxInflight != expectedMaxInflight {
		t.Logf("Expected maxInflight `%v` actual `%v`\n", expectedMaxInflight, config.MaxInflight)
		t.Fail()
	}

	expectedAckWait = time.Second * 30
	if config.AckWait != expectedAckWait {
		t.Logf("Expected maxInflight `%v` actual `%v`\n", expectedAckWait, config.AckWait)
		t.Fail()
	}

	os.Setenv("max_inflight", "10.00")
	os.Setenv("ack_wait", "10")

	config = readConfig.Read()

	expectedMaxInflight = 1
	if config.MaxInflight != expectedMaxInflight {
		t.Logf("Expected maxInflight `%v` actual `%v`\n", expectedMaxInflight, config.MaxInflight)
		t.Fail()
	}

	expectedAckWait = time.Second * 30
	if config.AckWait != expectedAckWait {
		t.Logf("Expected ackWait `%v` actual `%v`\n", expectedAckWait, config.AckWait)
		t.Fail()
	}
}
