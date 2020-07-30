// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

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

func isBoolValueSet(val string) bool {
	return len(val) > 0
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
func (ReadConfig) Read(hasEnv HasEnv) WatchdogConfig {
	cfg := WatchdogConfig{
		writeDebug:    false,
		cgiHeaders:    true,
		combineOutput: true,
	}

	cfg.faasProcess = hasEnv.Getenv("fprocess")

	cfg.readTimeout = parseIntOrDurationValue(hasEnv.Getenv("read_timeout"), time.Second*5)
	cfg.writeTimeout = parseIntOrDurationValue(hasEnv.Getenv("write_timeout"), time.Second*5)
	cfg.shutdownTimeout = parseIntOrDurationValue(hasEnv.Getenv("shutdown_timeout"), cfg.writeTimeout)

	cfg.execTimeout = parseIntOrDurationValue(hasEnv.Getenv("exec_timeout"), time.Second*0)
	cfg.port = parseIntValue(hasEnv.Getenv("port"), 8080)

	writeDebugEnv := hasEnv.Getenv("write_debug")
	if isBoolValueSet(writeDebugEnv) {
		cfg.writeDebug = parseBoolValue(writeDebugEnv)
	}

	cgiHeadersEnv := hasEnv.Getenv("cgi_headers")
	if isBoolValueSet(cgiHeadersEnv) {
		cfg.cgiHeaders = parseBoolValue(cgiHeadersEnv)
	}

	cfg.marshalRequest = parseBoolValue(hasEnv.Getenv("marshal_request"))
	cfg.debugHeaders = parseBoolValue(hasEnv.Getenv("debug_headers"))

	cfg.suppressLock = parseBoolValue(hasEnv.Getenv("suppress_lock"))

	cfg.contentType = hasEnv.Getenv("content_type")

	if isBoolValueSet(hasEnv.Getenv("combine_output")) {
		cfg.combineOutput = parseBoolValue(hasEnv.Getenv("combine_output"))
	}

	cfg.metricsPort = 8081
	cfg.maxInflight = parseIntValue(hasEnv.Getenv("max_inflight"), 0)

	return cfg
}

// WatchdogConfig for the process.
type WatchdogConfig struct {

	// HTTP read timeout
	readTimeout time.Duration

	// HTTP write timeout
	writeTimeout time.Duration

	// faasProcess is the process to exec
	faasProcess string

	// duration until the faasProcess will be killed
	execTimeout time.Duration

	// when SIGTERM is sent the server will wait `shutdownTimeout` before
	// closing off connections and a futher `shutdownTimeout` before exiting
	// defaults to `writeTimeout` if not specified
	shutdownTimeout time.Duration

	// writeDebug write console stdout statements to the container
	writeDebug bool

	// marshal header and body via JSON
	marshalRequest bool

	// cgiHeaders will make environmental variables available with all the HTTP headers.
	cgiHeaders bool

	// prints out all incoming and out-going HTTP headers
	debugHeaders bool

	// Don't write a lock file to /tmp/
	suppressLock bool

	// contentType forces a specific pre-defined value for all responses
	contentType string

	// port for HTTP server
	port int

	// combineOutput combines stderr and stdout in response
	combineOutput bool

	// metricsPort is the HTTP port to serve metrics on
	metricsPort int

	// maxInflight limits the number of simultaneous
	// requests that the watchdog allows concurrently.
	// Any request which exceeds this limit will
	// have an immediate response of 429.
	maxInflight int
}
