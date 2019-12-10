// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"fmt"
	"testing"
	"time"
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
	readConfig := ReadConfig{}

	config, _ := readConfig.Read(defaults)

	if config.UseExternalProvider() != false {
		t.Log("Default for UseExternalProvider should be false")
		t.Fail()
	}

	if config.DirectFunctions != false {
		t.Log("Default for DirectFunctions should be false")
		t.Fail()
	}

	if len(config.DirectFunctionsSuffix) > 0 {
		t.Log("Default for DirectFunctionsSuffix should be empty")
		t.Fail()
	}

	if len(config.Namespace) > 0 {
		t.Log("Default for Namespace should be empty")
		t.Fail()
	}
}

func TestRead_NamespaceOverride(t *testing.T) {

	defaults := NewEnvBucket()
	readConfig := ReadConfig{}

	defaults.Setenv("function_namespace", "fn")
	wantSuffix := "fn"

	config, _ := readConfig.Read(defaults)

	if config.Namespace != wantSuffix {
		t.Logf("Namespace want: %s, got: %s", wantSuffix, config.Namespace)
		t.Fail()
	}
}

func TestRead_NamespaceOverrideAgressWithFunctionSuffix_Valid(t *testing.T) {

	defaults := NewEnvBucket()
	readConfig := ReadConfig{}

	defaults.Setenv("direct_functions", "true")
	wantSuffix := "openfaas-fn.cluster.local.svc."

	defaults.Setenv("direct_functions_suffix", wantSuffix)
	defaults.Setenv("function_namespace", "openfaas-fn")

	_, err := readConfig.Read(defaults)

	if err != nil {
		t.Logf("Error found: %s", err)
		t.Fail()
	}
}

func TestRead_NamespaceOverrideAgressWithFunctionSuffix_Invalid(t *testing.T) {

	defaults := NewEnvBucket()
	readConfig := ReadConfig{}

	defaults.Setenv("direct_functions", "true")
	wantSuffix := "openfaas-fn.cluster.local.svc."

	defaults.Setenv("direct_functions_suffix", wantSuffix)
	defaults.Setenv("function_namespace", "fn")

	_, err := readConfig.Read(defaults)

	if err == nil {
		t.Logf("Expected an error because function_namespace should be a sub-string of direct_functions_suffix")
		t.Fail()
		return
	}

	want := "function_namespace must be a sub-string of direct_functions_suffix"

	if want != err.Error() {
		t.Logf("Error want: %s, got: %s", want, err.Error())
		t.Fail()
	}

}

func TestRead_DirectFunctionsOverride(t *testing.T) {
	defaults := NewEnvBucket()
	readConfig := ReadConfig{}
	defaults.Setenv("direct_functions", "true")
	wantSuffix := "openfaas-fn.cluster.local.svc."
	defaults.Setenv("direct_functions_suffix", wantSuffix)

	config, _ := readConfig.Read(defaults)

	if config.DirectFunctions != true {
		t.Logf("DirectFunctions should be true, got: %v", config.DirectFunctions)
		t.Fail()
	}

	if config.DirectFunctionsSuffix != wantSuffix {
		t.Logf("DirectFunctionsSuffix want: %s, got: %s", wantSuffix, config.DirectFunctionsSuffix)
		t.Fail()
	}
}

func TestRead_ScaleZeroDefaultAndOverride(t *testing.T) {
	defaults := NewEnvBucket()
	readConfig := ReadConfig{}
	// defaults.Setenv("scale_from_zero", "true")
	config, _ := readConfig.Read(defaults)

	want := false
	if config.ScaleFromZero != want {
		t.Logf("ScaleFromZero should be %v, got: %v", want, config.ScaleFromZero)
		t.Fail()
	}

	defaults.Setenv("scale_from_zero", "true")
	config, _ = readConfig.Read(defaults)
	want = true

	if config.ScaleFromZero != want {
		t.Logf("ScaleFromZero was overriden - should be %v, got: %v", want, config.ScaleFromZero)
		t.Fail()
	}

}

func TestRead_EmptyTimeoutConfig(t *testing.T) {
	defaults := NewEnvBucket()
	readConfig := ReadConfig{}

	config, _ := readConfig.Read(defaults)

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

	readConfig := ReadConfig{}
	config, _ := readConfig.Read(defaults)

	if (config.ReadTimeout) != time.Duration(10)*time.Second {
		t.Logf("ReadTimeout incorrect, got: %d\n", config.ReadTimeout)
		t.Fail()
	}
	if (config.WriteTimeout) != time.Duration(60)*time.Second {
		t.Logf("WriteTimeout incorrect, got: %d\n", config.WriteTimeout)
		t.Fail()
	}
}

func TestRead_ReadAndWriteTimeoutDurationConfig(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("read_timeout", "20s")
	defaults.Setenv("write_timeout", "1m30s")

	readConfig := ReadConfig{}
	config, _ := readConfig.Read(defaults)

	if (config.ReadTimeout) != time.Duration(20)*time.Second {
		t.Logf("ReadTimeout incorrect, got: %d\n", config.ReadTimeout)
		t.Fail()
	}
	if (config.WriteTimeout) != time.Duration(90)*time.Second {
		t.Logf("WriteTimeout incorrect, got: %d\n", config.WriteTimeout)
		t.Fail()
	}
}

func TestRead_UseNATSDefaultsToOff(t *testing.T) {
	defaults := NewEnvBucket()
	readConfig := ReadConfig{}

	config, _ := readConfig.Read(defaults)

	if config.UseNATS() == true {
		t.Log("NATS is supposed to be off by default")
		t.Fail()
	}
}

func TestRead_UseNATS(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("faas_nats_address", "nats")
	defaults.Setenv("faas_nats_port", "6222")
	readConfig := ReadConfig{}

	config, _ := readConfig.Read(defaults)

	if config.UseNATS() == false {
		t.Log("NATS was requested in config, but not enabled.")
		t.Fail()
	}

	wantNATSAddress := "nats"
	if *config.NATSAddress != wantNATSAddress {
		t.Logf("faas_nats_address: want %s, got %s", wantNATSAddress, *config.NATSAddress)
		t.Fail()
	}

	wantNATSPort := 6222
	if *config.NATSPort != wantNATSPort {
		t.Logf("faas_nats_port: want %d, got %d", wantNATSPort, *config.NATSPort)
		t.Fail()
	}

	wantNATSClusterName := "faas-cluster"
	if *config.NATSClusterName != wantNATSClusterName {
		t.Logf("faas_nats_cluster_name: want %s, got %s", wantNATSClusterName, *config.NATSClusterName)
		t.Fail()
	}

	wantNATSChannel := "faas-request"
	if *config.NATSChannel != wantNATSChannel {
		t.Logf("faas_nats_channel: want %s, got %s", wantNATSChannel, *config.NATSChannel)
		t.Fail()
	}

	defaults.Setenv("faas_nats_cluster_name", "example-nats-cluster")
	config, _ = readConfig.Read(defaults)

	wantNATSClusterName = "example-nats-cluster"
	if *config.NATSClusterName != wantNATSClusterName {
		t.Logf("faas_nats_cluster_name: want %s, got %s", wantNATSClusterName, *config.NATSClusterName)
		t.Fail()
	}

	defaults.Setenv("faas_nats_channel", "foo")
	config, _ = readConfig.Read(defaults)

	wantNATSChannel = "foo"
	if *config.NATSChannel != wantNATSChannel {
		t.Logf("faas_nats_channel: want %s, got %s", wantNATSChannel, *config.NATSChannel)
		t.Fail()
	}
}

func TestRead_UseNATSBadPort(t *testing.T) {

	defaults := NewEnvBucket()
	defaults.Setenv("faas_nats_address", "nats")
	defaults.Setenv("faas_nats_port", "6fff")
	defaults.Setenv("faas_nats_cluster_name", "example-nats-cluster")
	readConfig := ReadConfig{}

	_, err := readConfig.Read(defaults)

	if err != nil {
		want := "faas_nats_port invalid number: 6fff"

		if want != err.Error() {
			t.Errorf("want error: %q, got: %q", want, err.Error())
			t.Fail()
		}
	}

}

func TestRead_PrometheusNonDefaults(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("faas_prometheus_host", "prom1")
	defaults.Setenv("faas_prometheus_port", "9999")
	readConfig := ReadConfig{}

	config, _ := readConfig.Read(defaults)

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

	readConfig := ReadConfig{}

	config, _ := readConfig.Read(defaults)

	if config.PrometheusHost != "prometheus" {
		t.Logf("config.PrometheusHost, want: %s, got: %s\n", "prometheus", config.PrometheusHost)
		t.Fail()
	}

	if config.PrometheusPort != 9090 {
		t.Logf("config.PrometheusHost, want: %d, got: %d\n", 9090, config.PrometheusPort)
		t.Fail()
	}
}

func TestRead_BasicAuthDefaults(t *testing.T) {
	defaults := NewEnvBucket()

	readConfig := ReadConfig{}

	config, _ := readConfig.Read(defaults)

	if config.UseBasicAuth != false {
		t.Logf("config.UseBasicAuth, want: %t, got: %t\n", false, config.UseBasicAuth)
		t.Fail()
	}

	wantSecretsMount := "/run/secrets/"
	if config.SecretMountPath != wantSecretsMount {
		t.Logf("config.SecretMountPath, want: %s, got: %s\n", wantSecretsMount, config.SecretMountPath)
		t.Fail()
	}
}

func TestRead_BasicAuth_SetTrue(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("basic_auth", "true")
	defaults.Setenv("secret_mount_path", "/etc/openfaas/")

	readConfig := ReadConfig{}

	config, _ := readConfig.Read(defaults)

	if config.UseBasicAuth != true {
		t.Logf("config.UseBasicAuth, want: %t, got: %t\n", true, config.UseBasicAuth)
		t.Fail()
	}

	wantSecretsMount := "/etc/openfaas/"
	if config.SecretMountPath != wantSecretsMount {
		t.Logf("config.SecretMountPath, want: %s, got: %s\n", wantSecretsMount, config.SecretMountPath)
		t.Fail()
	}
}

func TestRead_MaxIdleConnsDefaults(t *testing.T) {
	defaults := NewEnvBucket()

	readConfig := ReadConfig{}

	config, _ := readConfig.Read(defaults)

	if config.MaxIdleConns != 1024 {
		t.Logf("config.MaxIdleConns, want: %d, got: %d\n", 1024, config.MaxIdleConns)
		t.Fail()
	}

	if config.MaxIdleConnsPerHost != 1024 {
		t.Logf("config.MaxIdleConnsPerHost, want: %d, got: %d\n", 1024, config.MaxIdleConnsPerHost)
		t.Fail()
	}
}

func TestRead_MaxIdleConns_Override(t *testing.T) {
	defaults := NewEnvBucket()

	readConfig := ReadConfig{}
	defaults.Setenv("max_idle_conns", fmt.Sprintf("%d", 100))
	defaults.Setenv("max_idle_conns_per_host", fmt.Sprintf("%d", 2))

	config, _ := readConfig.Read(defaults)

	if config.MaxIdleConns != 100 {
		t.Logf("config.MaxIdleConns, want: %d, got: %d\n", 100, config.MaxIdleConns)
		t.Fail()
	}

	if config.MaxIdleConnsPerHost != 2 {
		t.Logf("config.MaxIdleConnsPerHost, want: %d, got: %d\n", 2, config.MaxIdleConnsPerHost)
		t.Fail()
	}
}

func TestRead_AuthProxy_Defaults(t *testing.T) {
	defaults := NewEnvBucket()

	readConfig := ReadConfig{}
	wantURL := ""
	wantBody := false

	config, _ := readConfig.Read(defaults)

	if config.AuthProxyPassBody != wantBody {
		t.Logf("config.AuthProxyPassBody, want: %t, got: %t\n", wantBody, config.AuthProxyPassBody)
		t.Fail()
	}

	if config.AuthProxyURL != wantURL {
		t.Logf("config.AuthProxyURL, want: %s, got: %s\n", wantURL, config.AuthProxyURL)
		t.Fail()
	}
}

func TestRead_AuthProxy_DefaultsOverrides(t *testing.T) {
	defaults := NewEnvBucket()

	readConfig := ReadConfig{}
	wantURL := "http://auth.openfaas:8080/validate"
	wantBody := true
	defaults.Setenv("auth_proxy_url", wantURL)
	defaults.Setenv("auth_proxy_pass_body", fmt.Sprintf("%t", wantBody))

	config, _ := readConfig.Read(defaults)

	if config.AuthProxyPassBody != wantBody {
		t.Logf("config.AuthProxyPassBody, want: %t, got: %t\n", wantBody, config.AuthProxyPassBody)
		t.Fail()
	}

	if config.AuthProxyURL != wantURL {
		t.Logf("config.AuthProxyURL, want: %s, got: %s\n", wantURL, config.AuthProxyURL)
		t.Fail()
	}
}

func TestRead_LogsProviderURL(t *testing.T) {
	defaults := NewEnvBucket()

	t.Run("default value is nil when functions_provider_url is empty", func(t *testing.T) {
		readConfig := ReadConfig{}
		config, _ := readConfig.Read(defaults)
		if config.LogsProviderURL != nil {
			t.Fatalf("config.LogsProviderURL, want: %s, got: %s\n", "", config.LogsProviderURL)
		}
	})

	t.Run("default value is equal to functions_provider_url", func(t *testing.T) {
		expected := "functions.example.com"
		defaults.Setenv("functions_provider_url", expected)

		readConfig := ReadConfig{}
		config, _ := readConfig.Read(defaults)
		if config.LogsProviderURL.String() != expected {
			t.Fatalf("config.LogsProviderURL, want: %s, got: %s\n", expected, config.LogsProviderURL)
		}
	})

	t.Run("override by logs_provider_url", func(t *testing.T) {
		expected := "logs.example.com"
		defaults.Setenv("logs_provider_url", expected)

		readConfig := ReadConfig{}
		config, _ := readConfig.Read(defaults)
		if config.LogsProviderURL.String() != expected {
			t.Fatalf("config.LogsProviderURL, want: %s, got: %s\n", expected, config.LogsProviderURL)
		}
	})
}
