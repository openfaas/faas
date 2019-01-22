// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"log"
	"net/url"
	"os"
	"strconv"
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
func (ReadConfig) Read(hasEnv HasEnv) GatewayConfig {
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
			log.Fatal("If functions_provider_url is provided, then it should be a valid URL.", err)
		}
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
			log.Println("faas_nats_port invalid number: " + faasNATSPort)
		}
	}

	prometheusPort := hasEnv.Getenv("faas_prometheus_port")
	if len(prometheusPort) > 0 {
		prometheusPortVal, err := strconv.Atoi(prometheusPort)
		if err != nil {
			log.Println("Invalid port for faas_prometheus_port")
		} else {
			cfg.PrometheusPort = prometheusPortVal
		}
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
			log.Println("Invalid value for max_idle_conns")
		} else {
			cfg.MaxIdleConns = val
		}
	}

	maxIdleConnsPerHost := hasEnv.Getenv("max_idle_conns_per_host")
	if len(maxIdleConnsPerHost) > 0 {
		val, err := strconv.Atoi(maxIdleConnsPerHost)
		if err != nil {
			log.Println("Invalid value for max_idle_conns_per_host")
		} else {
			cfg.MaxIdleConnsPerHost = val
		}
	}

	return cfg
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

	// Address of the NATS service. Required for async mode.
	NATSAddress *string

	// Port of the NATS Service. Required for async mode.
	NATSPort *int

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

	MaxIdleConns int

	MaxIdleConnsPerHost int
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
