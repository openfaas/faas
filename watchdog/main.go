// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/openfaas/faas/watchdog/types"
)

func main() {
	osEnv := types.OsEnv{}
	readConfig := ReadConfig{}
	config := readConfig.Read(osEnv)

	if len(config.faasProcess) == 0 {
		log.Panicln("Provide a valid process via fprocess environmental variable.")
		return
	}

	readTimeout := config.readTimeout
	writeTimeout := config.writeTimeout

	// Move to readconfig.go
	var tcpPort = 8080
	if value, exists := os.LookupEnv("port"); exists {
		tcpPort, _ = strconv.Atoi(value)
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", tcpPort),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	var handler http.HandlerFunc

	// Move to readconfig.go
	if len(os.Getenv("afterburn")) > 0 {
		handler = makeAfterburnRequestHandler(&config)
	} else {
		handler = makeRequestHandler(&config)
	}

	http.HandleFunc("/", handler)

	path := filepath.Join(os.TempDir(), ".lock")
	log.Printf("Writing lock-file to: %s\n", path)
	writeErr := ioutil.WriteFile(path, []byte{}, 0660)
	if writeErr != nil {
		log.Panicf("Cannot write %s. To disable lock-file set env suppress_lock=true.\n Error: %s.\n", path, writeErr.Error())
	}
	log.Fatal(s.ListenAndServe())
}
