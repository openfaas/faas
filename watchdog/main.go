// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openfaas/faas/watchdog/types"
)

var version bool

func main() {
	flag.BoolVar(&version, "version", false, "Print the version and exit")

	flag.Parse()
	printVersion()

	if version == true {
		return
	}

	acceptingConnections = false

	osEnv := types.OsEnv{}
	readConfig := ReadConfig{}
	config := readConfig.Read(osEnv)

	if len(config.faasProcess) == 0 {
		log.Panicln("Provide a valid process via fprocess environmental variable.")
		return
	}

	readTimeout := config.readTimeout
	writeTimeout := config.writeTimeout

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.port),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	http.HandleFunc("/_/health", makeHealthHandler())
	http.HandleFunc("/", makeRequestHandler(&config))

	if config.suppressLock == false {
		path, writeErr := createLockFile()

		if writeErr != nil {
			log.Panicf("Cannot write %s. To disable lock-file set env suppress_lock=true.\n Error: %s.\n", path, writeErr.Error())
		}
	} else {
		log.Println("Warning: \"suppress_lock\" is enabled. No automated health-checks will be in place for your function.")
		acceptingConnections = true
	}

	listenUntilShutdown(config.writeTimeout, s)
}

func listenUntilShutdown(shutdownTimeout time.Duration, s *http.Server) {

	idleConnsClosed := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM)

		<-sig

		log.Printf("SIGTERM received.. shutting down server")

		acceptingConnections = false

		if err := s.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("Error in Shutdown: %v", err)
		}

		<-time.Tick(shutdownTimeout)

		close(idleConnsClosed)
	}()

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("Error ListenAndServe: %v", err)
		close(idleConnsClosed)
	}

	<-idleConnsClosed
}

func printVersion() {
	sha := "unknown"
	if len(GitCommit) > 0 {
		sha = GitCommit
	}

	log.Printf("Version: %v\tSHA: %v\n", BuildVersion(), sha)
}
