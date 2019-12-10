// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// OsEnv implements interface to wrap os.Getenv
type OsEnv struct {
}

// Getenv wraps os.Getenv
func (OsEnv) Getenv(key string) string {
	return os.Getenv(key)
}

// HasEnv provides interface for os.Getenv
type HasEnv interface {
	Getenv(key string) string
}

// ReadConfig constitutes config from env variables
type ReadConfig struct {
}

func parseBoolValue(val string) bool {
	if val == "true" {
		return true
	}
	return false
}

func parseIntOrDurationValue(val string, fallback time.Duration) time.Duration {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return time.Duration(parsedVal) * time.Second
		}
	}

	duration, durationErr := time.ParseDuration(val)
	if durationErr != nil {
		return fallback
	}
	return duration
}

// Read fetches gateway server configuration from environmental variables
func (ReadConfig) Read(hasEnv HasEnv) (*GatewayConfig, error) {
	cfg := GatewayConfig{
		PrometheusHost: "prometheus",
		PrometheusPort: 9090,
	}

	defaultDuration := time.Second * 8

	cfg.ReadTimeout = parseIntOrDurationValue(hasEnv.Getenv("read_timeout"), defaultDuration)
	cfg.WriteTimeout = parseIntOrDurationValue(hasEnv.Getenv("write_timeout"), defaultDuration)
	cfg.UpstreamTimeout = parseIntOrDurationValue(hasEnv.Getenv("upstream_timeout"), defaultDuration)

	if len(hasEnv.Getenv("functions_provider_url")) > 0 {
		var err error
		cfg.FunctionsProviderURL, err = url.Parse(hasEnv.Getenv("functions_provider_url"))
		if err != nil {
			return nil, fmt.Errorf("if functions_provider_url is provided, then it should be a valid URL, error: %s", err)
		}
	}

	if len(hasEnv.Getenv("logs_provider_url")) > 0 {
		var err error
		cfg.LogsProviderURL, err = url.Parse(hasEnv.Getenv("logs_provider_url"))
		if err != nil {
			return nil, fmt.Errorf("if logs_provider_url is provided, then it should be a valid URL, error: %s", err)
		}
	} else if cfg.FunctionsProviderURL != nil {
		cfg.LogsProviderURL, _ = url.Parse(cfg.FunctionsProviderURL.String())
	}

	faasNATSAddress := hasEnv.Getenv("faas_nats_address")
	if len(faasNATSAddress) > 0 {
		cfg.NATSAddress = &faasNATSAddress
	}

	faasNATSPort := hasEnv.Getenv("faas_nats_port")
	if len(faasNATSPort) > 0 {
		port, err := strconv.Atoi(faasNATSPort)
		if err == nil {
			cfg.NATSPort = &port
		} else {
			return nil, fmt.Errorf("faas_nats_port invalid number: %s", faasNATSPort)
		}
	}

	faasNATSClusterName := hasEnv.Getenv("faas_nats_cluster_name")
	if len(faasNATSClusterName) > 0 {
		cfg.NATSClusterName = &faasNATSClusterName
	} else {
		v := "faas-cluster"
		cfg.NATSClusterName = &v
	}

	faasNATSChannel := hasEnv.Getenv("faas_nats_channel")
	if len(faasNATSChannel) > 0 {
		cfg.NATSChannel = &faasNATSChannel
	} else {
		v := "faas-request"
		cfg.NATSChannel = &v
	}

	prometheusPort := hasEnv.Getenv("faas_prometheus_port")
	if len(prometheusPort) > 0 {
		prometheusPortVal, err := strconv.Atoi(prometheusPort)
		if err != nil {
			return nil, fmt.Errorf("faas_prometheus_port invalid number: %s", faasNATSPort)
		}
		cfg.PrometheusPort = prometheusPortVal

	}

	prometheusHost := hasEnv.Getenv("faas_prometheus_host")
	if len(prometheusHost) > 0 {
		cfg.PrometheusHost = prometheusHost
	}

	cfg.DirectFunctions = parseBoolValue(hasEnv.Getenv("direct_functions"))
	cfg.DirectFunctionsSuffix = hasEnv.Getenv("direct_functions_suffix")

	cfg.UseBasicAuth = parseBoolValue(hasEnv.Getenv("basic_auth"))

	secretPath := hasEnv.Getenv("secret_mount_path")
	if len(secretPath) == 0 {
		secretPath = "/run/secrets/"
	}
	cfg.SecretMountPath = secretPath
	cfg.ScaleFromZero = parseBoolValue(hasEnv.Getenv("scale_from_zero"))

	cfg.MaxIdleConns = 1024
	cfg.MaxIdleConnsPerHost = 1024

	maxIdleConns := hasEnv.Getenv("max_idle_conns")
	if len(maxIdleConns) > 0 {
		val, err := strconv.Atoi(maxIdleConns)
		if err != nil {
			return nil, fmt.Errorf("invalid value for max_idle_conns: %s", maxIdleConns)
		}
		cfg.MaxIdleConns = val

	}

	maxIdleConnsPerHost := hasEnv.Getenv("max_idle_conns_per_host")
	if len(maxIdleConnsPerHost) > 0 {
		val, err := strconv.Atoi(maxIdleConnsPerHost)
		if err != nil {
			return nil, fmt.Errorf("invalid value for max_idle_conns_per_host: %s", maxIdleConnsPerHost)
		}
		cfg.MaxIdleConnsPerHost = val

	}

	cfg.AuthProxyURL = hasEnv.Getenv("auth_proxy_url")
	cfg.AuthProxyPassBody = parseBoolValue(hasEnv.Getenv("auth_proxy_pass_body"))

	cfg.Namespace = hasEnv.Getenv("function_namespace")

	if len(cfg.DirectFunctionsSuffix) > 0 && len(cfg.Namespace) > 0 {
		if strings.HasPrefix(cfg.DirectFunctionsSuffix, cfg.Namespace) == false {
			return nil, fmt.Errorf("function_namespace must be a sub-string of direct_functions_suffix")
		}
	}

	return &cfg, nil
}

// GatewayConfig provides config for the API Gateway server process
type GatewayConfig struct {

	// HTTP timeout for reading a request from clients.
	ReadTimeout time.Duration

	// HTTP timeout for writing a response from functions.
	WriteTimeout time.Duration

	// UpstreamTimeout maximum duration of HTTP call to upstream URL
	UpstreamTimeout time.Duration

	// URL for alternate functions provider.
	FunctionsProviderURL *url.URL

	// URL for alternate function logs provider.
	LogsProviderURL *url.URL

	// Address of the NATS service. Required for async mode.
	NATSAddress *string

	// Port of the NATS Service. Required for async mode.
	NATSPort *int

	// The name of the NATS Streaming cluster. Required for async mode.
	NATSClusterName *string

	// NATSChannel is the name of the NATS Streaming channel used for asynchronous function invocations.
	NATSChannel *string

	// Host to connect to Prometheus.
	PrometheusHost string

	// Port to connect to Prometheus.
	PrometheusPort int

	// If set to true we will access upstream functions directly rather than through the upstream provider
	DirectFunctions bool

	// If set this will be used to resolve functions directly
	DirectFunctionsSuffix string

	// If set, reads secrets from file-system for enabling basic auth.
	UseBasicAuth bool

	// SecretMountPath specifies where to read secrets from for embedded basic auth
	SecretMountPath string

	// Enable the gateway to scale any service from 0 replicas to its configured "min replicas"
	ScaleFromZero bool

	// MaxIdleConns with a default value of 1024, can be used for tuning HTTP proxy performance
	MaxIdleConns int

	// MaxIdleConnsPerHost with a default value of 1024, can be used for tuning HTTP proxy performance
	MaxIdleConnsPerHost int

	// AuthProxyURL specifies URL for an authenticating proxy, disabled when blank, enabled when valid URL i.e. http://basic-auth.openfaas:8080/validate
	AuthProxyURL string

	// AuthProxyPassBody pass body to validation proxy
	AuthProxyPassBody bool

	// Namespace for endpoints
	Namespace string
}

// UseNATS Use NATSor not
func (g *GatewayConfig) UseNATS() bool {
	return g.NATSPort != nil &&
		g.NATSAddress != nil
}

// UseExternalProvider decide whether to bypass built-in Docker Swarm engine
func (g *GatewayConfig) UseExternalProvider() bool {
	return g.FunctionsProviderURL != nil
}
