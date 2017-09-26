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

	if r.Body != nil {
		defer r.Body.Close()
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

func debugHeaders(source *http.Header, direction string) {
	for k, vv := range *source {
		fmt.Printf("[%s] %s=%s\n", direction, k, vv)
	}
}

type requestInfo struct {
	headerWritten bool
}

func pipeRequest(config *WatchdogConfig, w http.ResponseWriter, r *http.Request, method string, hasBody bool) {
	startTime := time.Now()

	parts := strings.Split(config.faasProcess, " ")

	ri := &requestInfo{}

	if config.debugHeaders {
		debugHeaders(&r.Header, "in")
	}

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
	if hasBody == false {
		wgCount = 1
	}

	if hasBody {
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
	}

	wg.Add(wgCount)

	var timer *time.Timer

	if config.execTimeout > 0*time.Second {
		timer = time.NewTimer(config.execTimeout)

		go func() {
			<-timer.C
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
		}()
	}

	// Only write body if this is appropriate for the method.
	if hasBody {
		// Write to pipe in separate go-routine to prevent blocking
		go func() {
			defer wg.Done()
			writer.Write(requestBody)
			writer.Close()
		}()
	}

	go func() {
		defer wg.Done()
		out, err = targetCmd.CombinedOutput()
	}()

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

	if ri.headerWritten == false {
		execTime := time.Since(startTime).Seconds()
		w.Header().Set("X-Duration-Seconds", fmt.Sprintf("%f", execTime))
		ri.headerWritten = true
		w.WriteHeader(200)
		w.Write(out)
	}

	if config.debugHeaders {
		header := w.Header()
		debugHeaders(&header, "out")
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

		log.Println(r.URL.String())
		if len(r.URL.String()) > 0 {
			envs = append(envs, fmt.Sprintf("Http_Query=%s", r.URL.String()))
		}
	}

	return envs
}

func makeRequestHandler(config *WatchdogConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case
			"POST",
			"PUT",
			"DELETE",
			"UPDATE":
			pipeRequest(config, w, r, r.Method, true)
			break
		case
			"GET":
			pipeRequest(config, w, r, r.Method, false)
			break
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)

		}
	}
}
