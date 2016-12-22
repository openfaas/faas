package main

import (
	"bytes"
	"context"
	"log"
	"net/http"

	"io/ioutil"

	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
)

func lookupSwarmService(serviceName string) (bool, error) {
	var c *client.Client
	var err error
	c, err = client.NewEnvClient()
	if err != nil {
		log.Fatal("Error with Docker client.")
	}
	serviceFilter := filters.NewArgs()
	serviceFilter.Add("name", serviceName)
	services, err := c.ServiceList(context.Background(), types.ServiceListOptions{Filters: serviceFilter})

	return len(services) > 0, err
}

func proxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Println(r.Header)
		header := r.Header["X-Function"]
		log.Println(header)

		exists, err := lookupSwarmService(header[0])
		if err != nil {
			log.Fatalln(err)
		}

		if exists == true {
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
