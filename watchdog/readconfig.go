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
		cgiHeaders: true,
	}

	cfg.faasProcess = hasEnv.Getenv("fprocess")

	readTimeout := parseIntValue(hasEnv.Getenv("read_timeout"))
	writeTimeout := parseIntValue(hasEnv.Getenv("write_timeout"))

	cfg.execTimeout = time.Duration(parseIntValue(hasEnv.Getenv("exec_timeout"))) * time.Second

	if readTimeout == 0 {
		readTimeout = 5
	}

	if writeTimeout == 0 {
		writeTimeout = 5
	}

	cfg.readTimeout = time.Duration(readTimeout) * time.Second
	cfg.writeTimeout = time.Duration(writeTimeout) * time.Second

	readDebugEnv := hasEnv.Getenv("read_debug")
	if isBoolValueSet(readDebugEnv) {
		cfg.readDebug = parseBoolValue(readDebugEnv)
	}

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

	// writeDebug write console stdout statements to the container
	writeDebug bool

	// readDebug print out request body
	readDebug bool

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
}
