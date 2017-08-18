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

func parseIntValue(val string, fallback int) int {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return parsedVal
		}
	}
	return fallback
}

// Read fetches config from environmental variables.
func (ReadConfig) Read(hasEnv HasEnv) GatewayConfig {
	cfg := GatewayConfig{}

	readTimeout := parseIntValue(hasEnv.Getenv("read_timeout"), 8)
	writeTimeout := parseIntValue(hasEnv.Getenv("write_timeout"), 8)

	cfg.ReadTimeout = time.Duration(readTimeout) * time.Second
	cfg.WriteTimeout = time.Duration(writeTimeout) * time.Second

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

	return cfg
}

// GatewayConfig for the process.
type GatewayConfig struct {
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	FunctionsProviderURL *url.URL
	NATSAddress          *string
	NATSPort             *int
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
