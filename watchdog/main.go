package main

import (
	"bytes"
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
	if config.marshallRequest {
		marshalRes, marshallErr := types.MarshalRequest(requestBytes, &r.Header)
		err = marshallErr
		res = marshalRes
	} else {
		res = requestBytes
	}
	return res, err
}

func pipeRequest(config *WatchdogConfig, w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(config.faasProcess, " ")

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

	// Match header for strict services
	if r.Header.Get("Content-Type") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(200)
	w.Write(out)
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

	log.Fatal(s.ListenAndServe())
}
