package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
)

func makeAfterburnRequestHandler(config *WatchdogConfig) func(http.ResponseWriter, *http.Request) {

	var process *exec.Cmd
	var writePipe *io.WriteCloser
	var readPipe *io.ReadCloser
	var mutex *sync.Mutex

	parts := strings.Split(config.faasProcess, " ")

	process = exec.Command(parts[0], parts[1:]...)

	writePiper, writeErr := process.StdinPipe()
	writePipe = &writePiper
	if writeErr != nil {
		log.Fatalln(writeErr)
	}

	readPiper, readPipeErr := process.StdoutPipe()
	if readPipeErr != nil {
		log.Fatalln(readPipeErr)
	}

	readPipe = &readPiper
	mutex = &sync.Mutex{}

	go func() {
		log.Println("Run")
		err := process.Run()
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Process completed running.")
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		var bodyBytes []byte
		var wg sync.WaitGroup

		wgCount := 1

		if config.readDebug {
			fmt.Printf("Header: %s\n", r.Header)
			reqBody, _ := ioutil.ReadAll(r.Body)
			defer r.Body.Close()

			r.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

			fmt.Printf("Body: %s\n", reqBody)
		}

		log.Println(">> Lock mutex")
		mutex.Lock()

		wg.Add(wgCount)

		go func(p *exec.Cmd) {
			log.Println("Writing to process pipe")

			r.Write(*writePipe)
			log.Println("Written to process pipe")

			defer wg.Done()
		}(process)

		wg.Wait()
		log.Println("Synchronizing")

		wg.Add(wgCount)

		go func() {
			log.Println("Read pipe")

			buffReader := bufio.NewReader(*readPipe)
			processRes, err := http.ReadResponse(buffReader, r)

			log.Println("Read pipe 2")
			if err != nil {
				log.Println("read pipe error", err)
				w.WriteHeader(http.StatusInternalServerError)

				wg.Done()
				return
			}

			if processRes.Body != nil {
				defer processRes.Body.Close()
			}

			log.Printf("r.len=[%d] processRes.len=[%d]\n", r.ContentLength, processRes.ContentLength)

			var bodyErr error
			bodyBytes, bodyErr = ioutil.ReadAll(processRes.Body)
			if bodyErr != nil {
				log.Println("read body err", bodyErr)
			}

			w.WriteHeader(processRes.StatusCode)
			log.Printf("r.len=[%d] processRes.len=[%d] bodyBytes.len=[%d]\n", r.ContentLength, processRes.ContentLength, len(bodyBytes))

			// log.Println("bodyBytes:", string(bodyBytes), " len [", len(bodyBytes), "]")

			_, writeErr := w.Write(bodyBytes)
			if writeErr != nil {
				log.Println(writeErr)
			}
			if config.writeDebug {
				fmt.Printf("Response: %s", string(bodyBytes))
			}

			// defer processRes.Body.Close()
			defer wg.Done()
		}()

		log.Println("Waiting again")
		wg.Wait()

		log.Println("<< Unlock mutex")
		mutex.Unlock()
		log.Println(">> Mutex unlocked")

		// w.Write(bodyBytes)
	}
}
