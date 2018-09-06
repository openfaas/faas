// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package bootstrap

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/types"
)

var r *mux.Router

// Mark this as a Golang "package"
func init() {
	r = mux.NewRouter()
}

// Router gives access to the underlying router for when new routes need to be added.
func Router() *mux.Router {
	return r
}

// Serve load your handlers into the correct OpenFaaS route spec. This function is blocking.
func Serve(handlers *types.FaaSHandlers, config *types.FaaSConfig) {
	r.HandleFunc("/system/functions", handlers.FunctionReader).Methods("GET")
	r.HandleFunc("/system/functions", handlers.DeployHandler).Methods("POST")
	r.HandleFunc("/system/functions", handlers.DeleteHandler).Methods("DELETE")
	r.HandleFunc("/system/functions", handlers.UpdateHandler).Methods("PUT")

	r.HandleFunc("/system/function/{name:[-a-zA-Z_0-9]+}", handlers.ReplicaReader).Methods("GET")
	r.HandleFunc("/system/scale-function/{name:[-a-zA-Z_0-9]+}", handlers.ReplicaUpdater).Methods("POST")

	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}", handlers.FunctionProxy)
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/", handlers.FunctionProxy)

	r.HandleFunc("/system/info", handlers.InfoHandler).Methods("GET")

	if config.EnableHealth {
		r.HandleFunc("/healthz", handlers.Health).Methods("GET")
	}

	readTimeout := config.ReadTimeout
	writeTimeout := config.WriteTimeout

	tcpPort := 8080
	if config.TCPPort != nil {
		tcpPort = *config.TCPPort
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", tcpPort),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes, // 1MB - can be overridden by setting Server.MaxHeaderBytes.
		Handler:        r,
	}

	log.Fatal(s.ListenAndServe())
}
