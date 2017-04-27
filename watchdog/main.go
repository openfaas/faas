// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/alexellis/faas/watchdog/types"
)

func buildFunctionInput(config *WatchdogConfig, r *http.Request) ([]byte, error) {
	var res []byte
	var requestBytes []byte
	var err error

	defer r.Body.Close()
	requestBytes, _ = ioutil.ReadAll(r.Body)
	if config.marshalRequest {
		marshalRes, marshalErr := types.MarshalRequest(requestBytes, &r.Header)
		err = marshalErr
		res = marshalRes
	} else {
		res = requestBytes
	}
	return res, err
}

func debugHeaders(source *http.Header, direction string) {
	for k, vv := range *source {
		fmt.Printf("[%s] %s=%s\n", direction, k, vv)
	}
}

func pipeRequest(config *WatchdogConfig, w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(config.faasProcess, " ")

	if config.debugHeaders {
		debugHeaders(&r.Header, "in")
	}

	targetCmd := exec.Command(parts[0], parts[1:]...)
	writer, _ := targetCmd.StdinPipe()

	var out []byte
	var err error
	var res []byte

	var wg sync.WaitGroup
	wg.Add(2)

	res, buildInputErr := buildFunctionInput(config, r)
	if buildInputErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(buildInputErr.Error()))
		return
	}

	// Write to pipe in separate go-routine to prevent blocking
	go func() {
		defer wg.Done()
		writer.Write(res)
		writer.Close()
	}()

	go func() {
		defer wg.Done()
		out, err = targetCmd.CombinedOutput()
	}()

	wg.Wait()

	if err != nil {
		if config.writeDebug == true {
			log.Println(targetCmd, err)
		}
		w.WriteHeader(500)
		response := bytes.NewBufferString(err.Error())
		w.Write(response.Bytes())
		return
	}
	if config.writeDebug == true {
		os.Stdout.Write(out)
	}

	if len(config.contentType) > 0 {
		w.Header().Set("Content-Type", config.contentType)
	} else {

		// Match content-type of caller if no override specified.
		clientContentType := r.Header.Get("Content-Type")
		if len(clientContentType) > 0 {
			w.Header().Set("Content-Type", clientContentType)
		}
	}

	w.WriteHeader(200)
	w.Write(out)

	if config.debugHeaders {
		header := w.Header()
		debugHeaders(&header, "out")
	}
}

func makeRequestHandler(config *WatchdogConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			pipeRequest(config, w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func main() {
	osEnv := types.OsEnv{}
	readConfig := ReadConfig{}
	config := readConfig.Read(osEnv)

	if len(config.faasProcess) == 0 {
		log.Panicln("Provide a valid process via fprocess environmental variable.")
		return
	}

	readTimeout := time.Duration(config.readTimeout) * time.Second
	writeTimeout := time.Duration(config.writeTimeout) * time.Second

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	http.HandleFunc("/", makeRequestHandler(&config))

	if config.suppressLock == false {
		path := "/tmp/.lock"
		log.Printf("Writing lock-file to: %s\n", path)
		writeErr := ioutil.WriteFile(path, []byte{}, 0660)
		if writeErr != nil {
			log.Panicf("Cannot write %s. Error: %s\n", path, writeErr.Error())
		}
	}

	log.Fatal(s.ListenAndServe())
}
