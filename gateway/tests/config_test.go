// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package tests

import (
	"testing"
	"time"

	"github.com/openfaas/faas/gateway/types"
)

type EnvBucket struct {
	Items map[string]string
}

func NewEnvBucket() EnvBucket {
	return EnvBucket{
		Items: make(map[string]string),
	}
}

func (e EnvBucket) Getenv(key string) string {
	return e.Items[key]
}

func (e EnvBucket) Setenv(key string, value string) {
	e.Items[key] = value
}

func TestRead_UseExternalProvider_Defaults(t *testing.T) {
	defaults := NewEnvBucket()
	readConfig := types.ReadConfig{}

	config := readConfig.Read(defaults)

	if config.UseExternalProvider() != false {
		t.Log("Default for UseExternalProvider should be false")
		t.Fail()
	}
}

func TestRead_EmptyTimeoutConfig(t *testing.T) {
	defaults := NewEnvBucket()
	readConfig := types.ReadConfig{}

	config := readConfig.Read(defaults)

	if (config.ReadTimeout) != time.Duration(8)*time.Second {
		t.Log("ReadTimeout incorrect")
		t.Fail()
	}
	if (config.WriteTimeout) != time.Duration(8)*time.Second {
		t.Log("WriteTimeout incorrect")
		t.Fail()
	}
}

func TestRead_ReadAndWriteTimeoutConfig(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("read_timeout", "10")
	defaults.Setenv("write_timeout", "60")

	readConfig := types.ReadConfig{}
	config := readConfig.Read(defaults)

	if (config.ReadTimeout) != time.Duration(10)*time.Second {
		t.Logf("ReadTimeout incorrect, got: %d\n", config.ReadTimeout)
		t.Fail()
	}
	if (config.WriteTimeout) != time.Duration(60)*time.Second {
		t.Logf("WriteTimeout incorrect, got: %d\n", config.WriteTimeout)
		t.Fail()
	}
}

func TestRead_UseNATSDefaultsToOff(t *testing.T) {
	defaults := NewEnvBucket()
	readConfig := types.ReadConfig{}

	config := readConfig.Read(defaults)

	if config.UseNATS() == true {
		t.Log("NATS is supposed to be off by default")
		t.Fail()
	}
}

func TestRead_UseNATS(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("faas_nats_address", "nats")
	defaults.Setenv("faas_nats_port", "6222")
	readConfig := types.ReadConfig{}

	config := readConfig.Read(defaults)

	if config.UseNATS() == false {
		t.Log("NATS was requested in config, but not enabled.")
		t.Fail()
	}
}

func TestRead_UseNATSBadPort(t *testing.T) {

	defaults := NewEnvBucket()
	defaults.Setenv("faas_nats_address", "nats")
	defaults.Setenv("faas_nats_port", "6fff")
	readConfig := types.ReadConfig{}

	config := readConfig.Read(defaults)

	if config.UseNATS() == true {
		t.Log("NATS had bad config, should not be enabled.")
		t.Fail()
	}
}

func TestRead_PrometheusNonDefaults(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("faas_prometheus_host", "prom1")
	defaults.Setenv("faas_prometheus_port", "9999")
	readConfig := types.ReadConfig{}

	config := readConfig.Read(defaults)

	if config.PrometheusHost != "prom1" {
		t.Logf("config.PrometheusHost, want: %s, got: %s\n", "prom1", config.PrometheusHost)
		t.Fail()
	}

	if config.PrometheusPort != 9999 {
		t.Logf("config.PrometheusHost, want: %d, got: %d\n", 9999, config.PrometheusPort)
		t.Fail()
	}
}

func TestRead_PrometheusDefaults(t *testing.T) {
	defaults := NewEnvBucket()

	readConfig := types.ReadConfig{}

	config := readConfig.Read(defaults)

	if config.PrometheusHost != "prometheus" {
		t.Logf("config.PrometheusHost, want: %s, got: %s\n", "prometheus", config.PrometheusHost)
		t.Fail()
	}

	if config.PrometheusPort != 9090 {
		t.Logf("config.PrometheusHost, want: %d, got: %d\n", 9090, config.PrometheusPort)
		t.Fail()
	}
}
