package types

import (
	"fmt"
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

// ParseIntValue parses the the int in val or, if there is an error, returns the
// specified default value
func ParseIntValue(val string, fallback int) int {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return parsedVal
		}
	}
	return fallback
}

// ParseIntOrDurationValue parses the the duration in val or, if there is an error, returns the
// specified default value
func ParseIntOrDurationValue(val string, fallback time.Duration) time.Duration {
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

// ParseBoolValue parses the the boolean in val or, if there is an error, returns the
// specified default value
func ParseBoolValue(val string, fallback bool) bool {
	if len(val) > 0 {
		return val == "true"
	}
	return fallback
}

// ParseString verifies the string in val is not empty. When empty, it returns the
// specified default value
func ParseString(val string, fallback string) string {
	if len(val) > 0 {
		return val
	}
	return fallback
}

// Read fetches config from environmental variables.
func (ReadConfig) Read(hasEnv HasEnv) (*FaaSConfig, error) {
	cfg := &FaaSConfig{
		ReadTimeout:     ParseIntOrDurationValue(hasEnv.Getenv("read_timeout"), time.Second*10),
		WriteTimeout:    ParseIntOrDurationValue(hasEnv.Getenv("write_timeout"), time.Second*10),
		EnableBasicAuth: ParseBoolValue(hasEnv.Getenv("basic_auth"), false),
		// default value from Gateway
		SecretMountPath: ParseString(hasEnv.Getenv("secret_mount_path"), "/run/secrets/"),
	}

	port := ParseIntValue(hasEnv.Getenv("port"), 8080)
	cfg.TCPPort = &port

	cfg.MaxIdleConns = 1024
	maxIdleConns := hasEnv.Getenv("max_idle_conns")
	if len(maxIdleConns) > 0 {
		val, err := strconv.Atoi(maxIdleConns)
		if err != nil {
			return nil, fmt.Errorf("invalid value for max_idle_conns: %s", maxIdleConns)
		}
		cfg.MaxIdleConns = val

	}

	cfg.MaxIdleConnsPerHost = 1024
	maxIdleConnsPerHost := hasEnv.Getenv("max_idle_conns_per_host")
	if len(maxIdleConnsPerHost) > 0 {
		val, err := strconv.Atoi(maxIdleConnsPerHost)
		if err != nil {
			return nil, fmt.Errorf("invalid value for max_idle_conns_per_host: %s", maxIdleConnsPerHost)
		}
		cfg.MaxIdleConnsPerHost = val

	}

	return cfg, nil
}
