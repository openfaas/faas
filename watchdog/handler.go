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
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/openfaas/faas/watchdog/types"
)

type requestInfo struct {
	headerWritten bool
}

// buildFunctionInput for a GET method this is an empty byte array.
func buildFunctionInput(config *WatchdogConfig, r *http.Request) ([]byte, error) {
	var res []byte
	var requestBytes []byte
	var err error

	if r.Body == nil {
		return res, nil
	}
	defer r.Body.Close()

	if err != nil {
		log.Println(err)
		return res, err
	}

	requestBytes, err = ioutil.ReadAll(r.Body)
	if config.marshalRequest {
		marshalRes, marshalErr := types.MarshalRequest(requestBytes, &r.Header)
		err = marshalErr
		res = marshalRes
	} else {
		res = requestBytes
	}

	return res, err
}

// debugHeaders prints HTTP headers as key/value pairs
func debugHeaders(source *http.Header, direction string) {
	for k, vv := range *source {
		fmt.Printf("[%s] %s=%s\n", direction, k, vv)
	}
}

func pipeRequest(config *WatchdogConfig, w http.ResponseWriter, r *http.Request, method string) {
	startTime := time.Now()

	parts := strings.Split(config.faasProcess, " ")

	ri := &requestInfo{}

	if config.debugHeaders {
		debugHeaders(&r.Header, "in")
	}

	log.Println("Forking fprocess.")

	targetCmd := exec.Command(parts[0], parts[1:]...)

	envs := getAdditionalEnvs(config, r, method)
	if len(envs) > 0 {
		targetCmd.Env = envs
	}

	writer, _ := targetCmd.StdinPipe()

	var out []byte
	var err error
	var requestBody []byte

	var wg sync.WaitGroup

	wgCount := 2

	var buildInputErr error
	requestBody, buildInputErr = buildFunctionInput(config, r)
	if buildInputErr != nil {
		ri.headerWritten = true
		w.WriteHeader(http.StatusBadRequest)
		// I.e. "exit code 1"
		w.Write([]byte(buildInputErr.Error()))

		// Verbose message - i.e. stack trace
		w.Write([]byte("\n"))
		w.Write(out)

		return
	}

	wg.Add(wgCount)

	var timer *time.Timer

	if config.execTimeout > 0*time.Second {
		timer = time.AfterFunc(config.execTimeout, func() {
			log.Printf("Killing process: %s\n", config.faasProcess)
			if targetCmd != nil && targetCmd.Process != nil {
				ri.headerWritten = true
				w.WriteHeader(http.StatusRequestTimeout)

				w.Write([]byte("Killed process.\n"))

				val := targetCmd.Process.Kill()
				if val != nil {
					log.Printf("Killed process: %s - error %s\n", config.faasProcess, val.Error())
				}
			}
		})
	}

	// Write to pipe in separate go-routine to prevent blocking
	go func() {
		defer wg.Done()
		writer.Write(requestBody)
		writer.Close()
	}()

	if config.combineOutput {
		// Read the output from stdout/stderr and combine into one variable for output.
		go func() {
			defer wg.Done()

			out, err = targetCmd.CombinedOutput()
		}()
	} else {
		go func() {
			var b bytes.Buffer
			targetCmd.Stderr = &b

			defer wg.Done()

			out, err = targetCmd.Output()
			if b.Len() > 0 {
				log.Printf("stderr: %s", b.Bytes())
			}
			b.Reset()
		}()
	}

	wg.Wait()
	if timer != nil {
		timer.Stop()
	}

	if err != nil {
		if config.writeDebug == true {
			log.Printf("Success=%t, Error=%s\n", targetCmd.ProcessState.Success(), err.Error())
			log.Printf("Out=%s\n", out)
		}

		if ri.headerWritten == false {
			w.WriteHeader(http.StatusInternalServerError)
			response := bytes.NewBufferString(err.Error())
			w.Write(response.Bytes())
			w.Write([]byte("\n"))
			if len(out) > 0 {
				w.Write(out)
			}
			ri.headerWritten = true
		}
		return
	}

	var bytesWritten string
	if config.writeDebug == true {
		os.Stdout.Write(out)
	} else {
		bytesWritten = fmt.Sprintf("Wrote %d Bytes", len(out))
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

	execDuration := time.Since(startTime).Seconds()
	if ri.headerWritten == false {
		w.Header().Set("X-Duration-Seconds", fmt.Sprintf("%f", execDuration))
		ri.headerWritten = true
		w.WriteHeader(200)
		w.Write(out)
	}

	if config.debugHeaders {
		header := w.Header()
		debugHeaders(&header, "out")
	}

	if len(bytesWritten) > 0 {
		log.Printf("%s - Duration: %f seconds", bytesWritten, execDuration)
	} else {
		log.Printf("Duration: %f seconds", execDuration)
	}
}

func getAdditionalEnvs(config *WatchdogConfig, r *http.Request, method string) []string {
	var envs []string

	if config.cgiHeaders {
		envs = os.Environ()

		for k, v := range r.Header {
			kv := fmt.Sprintf("Http_%s=%s", strings.Replace(k, "-", "_", -1), v[0])
			envs = append(envs, kv)
		}

		envs = append(envs, fmt.Sprintf("Http_Method=%s", method))
		envs = append(envs, fmt.Sprintf("Http_ContentLength=%d", r.ContentLength))

		if config.writeDebug {
			log.Println("Query ", r.URL.RawQuery)
		}

		if len(r.URL.RawQuery) > 0 {
			envs = append(envs, fmt.Sprintf("Http_Query=%s", r.URL.RawQuery))
		}

		if config.writeDebug {
			log.Println("Path ", r.URL.Path)
		}

		if len(r.URL.Path) > 0 {
			envs = append(envs, fmt.Sprintf("Http_Path=%s", r.URL.Path))
		}

		if len(r.Host) > 0 {
			envs = append(envs, fmt.Sprintf("Http_Host=%s", r.Host))
		}

	}

	return envs
}

func lockFilePresent() bool {
	path := filepath.Join(os.TempDir(), ".lock")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func createLockFile() (string, error) {
	path := filepath.Join(os.TempDir(), ".lock")
	log.Printf("Writing lock-file to: %s\n", path)
	writeErr := ioutil.WriteFile(path, []byte{}, 0660)

	atomic.StoreInt32(&acceptingConnections, 1)

	return path, writeErr
}

func makeHealthHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if atomic.LoadInt32(&acceptingConnections) == 0 || lockFilePresent() == false {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))

			break
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func makeRequestHandler(config *WatchdogConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodGet:
			pipeRequest(config, w, r, r.Method)
			break
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)

		}
	}
}
