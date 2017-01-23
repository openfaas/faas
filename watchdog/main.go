package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	readTimeoutStr := os.Getenv("read_timeout")
	writeTimeoutStr := os.Getenv("write_timeout")
	writeDebugStr := os.Getenv("write_debug")
	process := os.Getenv("fprocess")

	readTimeout := 5 * time.Second
	writeTimeout := 5 * time.Second
	writeDebug := true

	if len(process) == 0 {
		log.Panicln("Provide a valid process via fprocess environmental variable.")
		return
	}

	if len(writeDebugStr) > 0 && writeDebugStr == "false" {
		writeDebug = false
	}

	if len(readTimeoutStr) > 0 {
		parsedVal, parseErr := strconv.Atoi(readTimeoutStr)
		if parseErr == nil && parsedVal > 0 {
			readTimeout = time.Duration(parsedVal) * time.Second
		}
	}

	if len(writeTimeoutStr) > 0 {
		parsedVal, parseErr := strconv.Atoi(writeTimeoutStr)
		if parseErr == nil && parsedVal > 0 {
			writeTimeout = time.Duration(parsedVal) * time.Second
		}
	}

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			parts := strings.Split(process, " ")

			targetCmd := exec.Command(parts[0], parts[1:]...)
			writer, _ := targetCmd.StdinPipe()

			res, _ := ioutil.ReadAll(r.Body)

			writer.Write(res)
			writer.Close()

			out, err := targetCmd.CombinedOutput()

			if err != nil {
				if writeDebug == true {
					log.Println(targetCmd, err)
				}

				w.WriteHeader(500)
				response := bytes.NewBufferString(err.Error())
				w.Write(response.Bytes())
				return
			}
			w.WriteHeader(200)

			if writeDebug == true {
				os.Stdout.Write(out)
			}

			w.Write(out)
		}
	})

	log.Fatal(s.ListenAndServe())
}
