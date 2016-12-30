package main

import (
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
			process := os.Getenv("fprocess")
			parts := strings.Split(process, " ")
			targetCmd := exec.Command(parts[0], parts[1:]...)
			writer, _ := targetCmd.StdinPipe()
			res, _ := ioutil.ReadAll(r.Body)
			writer.Write(res)
			writer.Close()
			out, err := targetCmd.Output()
			if err != nil {
				panic(err)
			}

			os.Stdout.Write(out)
			w.Write(out)
		}
	})

	log.Fatal(s.ListenAndServe())
}
