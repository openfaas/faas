package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// OsEnv implements interface to wrap os.Getenv
type OsEnv struct {
}

// Getenv wraps os.Getenv
func (OsEnv) Getenv(key string) string {
	return os.Getenv(key)
}

func makeRequestHandler(config *WatchdogConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			parts := strings.Split(config.faasProcess, " ")

			targetCmd := exec.Command(parts[0], parts[1:]...)
			writer, _ := targetCmd.StdinPipe()

			res, _ := ioutil.ReadAll(r.Body)

			writer.Write(res)
			writer.Close()

			out, err := targetCmd.CombinedOutput()

			if err != nil {
				if config.writeDebug == true {
					log.Println(targetCmd, err)
				}

				w.WriteHeader(500)
				response := bytes.NewBufferString(err.Error())
				w.Write(response.Bytes())
				return
			}
			w.WriteHeader(200)

			if config.writeDebug == true {
				os.Stdout.Write(out)
			}
			w.Write(out)
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
