package main

import (
	"strconv"
	"time"
)

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
	return true
}

func parseIntValue(val string) int {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)

		if parseErr == nil && parsedVal >= 0 {

			return parsedVal
		}
	}
	return 0
}

// Read fetches config from environmental variables.
func (ReadConfig) Read(hasEnv HasEnv) WatchdogConfig {
	cfg := WatchdogConfig{
		writeDebug: true,
	}

	cfg.faasProcess = hasEnv.Getenv("fprocess")

	readTimeout := parseIntValue(hasEnv.Getenv("read_timeout"))
	writeTimeout := parseIntValue(hasEnv.Getenv("write_timeout"))

	if readTimeout == 0 {
		readTimeout = 5
	}

	if writeTimeout == 0 {
		writeTimeout = 5
	}

	cfg.readTimeout = time.Duration(readTimeout) * time.Second
	cfg.writeTimeout = time.Duration(writeTimeout) * time.Second

	cfg.writeDebug = parseBoolValue(hasEnv.Getenv("write_debug"))

	return cfg
}

// WatchdogConfig for the process.
type WatchdogConfig struct {
	readTimeout  time.Duration
	writeTimeout time.Duration
	// faasProcess is the process to exec
	faasProcess string

	// writeDebug write console stdout statements to the container
	writeDebug bool
}
