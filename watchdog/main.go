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

func main() {
	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// get the fprocess to execute and it's args
			process := os.Getenv("fprocess")
			//

			// split the fprocess and it's args
			parts := strings.Split(process, " ")
			//

			// prepare the Command object for the fprocess
			fprocessCmd := exec.Command(parts[0], parts[1:]...)
			//

			// get a reference on all outputs for the fprocess (as byte[])
			var stdout bytes.Buffer
			fprocessCmd.Stdout = &stdout

			var stderr bytes.Buffer
			fprocessCmd.Stderr = &stderr
			//

			// write the http.Request's POST body to the stdin of the fprocess
			stdin, _ := fprocessCmd.StdinPipe()
			res, _ := ioutil.ReadAll(r.Body)
			stdin.Write(res)
			stdin.Close()
			//
			
			// execute the fprocess
			err := fprocessCmd.Run()
			//

			if err != nil {
				// get the fprocess's stderr content
				errorStack := stderr.String()
				//

				// print the error's details to the logger
				log.Println(err.Error())
				log.Println(errorStack)
				//

				// send the http response with the error that occured
				w.WriteHeader(500)
				response := bytes.NewBufferString(err.Error() + "\n" + errorStack)
				w.Write(response.Bytes())
				//

				return
			}

			// TODO: consider stdout to container as configurable via env-variable.
			// write the fprocess's stdout to the fwatchdog stdout
			os.Stdout.Write(stdout.Bytes())
			//

			// send the http response with the fprocess's stdout
			w.WriteHeader(200)
			w.Write(stdout.Bytes())
			//
		}
	})

	log.Fatal(s.ListenAndServe())
}
