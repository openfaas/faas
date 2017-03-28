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
)

// OsEnv implements interface to wrap os.Getenv
type OsEnv struct {
}

// Getenv wraps os.Getenv
func (OsEnv) Getenv(key string) string {
	return os.Getenv(key)
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

	res, _ = ioutil.ReadAll(r.Body)
	defer r.Body.Close()

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
	osEnv := OsEnv{}
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
