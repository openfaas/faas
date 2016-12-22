package main

import (
	"bytes"
	"log"
	"net/http"

	"io/ioutil"

	"strconv"

	"github.com/gorilla/mux"
)

func proxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Println(r.Header)
		header := r.Header["X-Function"]
		log.Println(header)
		if header[0] == "catservice" {
			// client := http.Client{Timeout: time.Second * 2}
			requestBody, _ := ioutil.ReadAll(r.Body)
			buf := bytes.NewBuffer(requestBody)

			response, err := http.Post("http://"+header[0]+":"+strconv.Itoa(8080)+"/", "text/plain", buf)
			if err != nil {
				log.Fatalln(err)
			}
			responseBody, _ := ioutil.ReadAll(response.Body)
			w.Write(responseBody)

		}
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", proxy)
	log.Fatal(http.ListenAndServe(":8080", r))
}
