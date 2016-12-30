package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"io/ioutil"

	"encoding/json"

	"github.com/alexellis/faas/gateway/metrics"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

type AlexaSessionApplication struct {
	ApplicationId string `json:"applicationId"`
}

type AlexaSession struct {
	SessionId   string                  `json:"sessionId"`
	Application AlexaSessionApplication `json:"application"`
}

type AlexaIntent struct {
	Name string `json:"name"`
}

type AlexaRequest struct {
	Intent AlexaIntent `json:"intent"`
}

type AlexaRequestBody struct {
	Session AlexaSession `json:"session"`
	Request AlexaRequest `json:"request"`
}

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

func isAlexa(requestBody []byte) AlexaRequestBody {
	body := AlexaRequestBody{}
	buf := bytes.NewBuffer(requestBody)
	fmt.Println(buf)
	str := buf.String()
	parts := strings.Split(str, "sessionId")
	if len(parts) > 0 {
		json.Unmarshal(requestBody, &body)
		fmt.Println("Alexa SDK request found")
		fmt.Printf("Session=%s, Intent=%s, App=%s\n", body.Session.SessionId, body.Request.Intent, body.Session.Application.ApplicationId)
	}
	return body
}

func invokeService(w http.ResponseWriter, r *http.Request, metrics metrics.MetricOptions, service string, requestBody []byte) {
	stamp := strconv.FormatInt(time.Now().Unix(), 10)

	start := time.Now()
	buf := bytes.NewBuffer(requestBody)
	url := "http://" + service + ":" + strconv.Itoa(8080) + "/"
	fmt.Printf("[%s] Forwarding request to: %s\n", stamp, url)
	response, err := http.Post(url, "text/plain", buf)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		buf := bytes.NewBufferString("Can't reach service: " + service)
		w.Write(buf.Bytes())
		return
	}

	responseBody, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		fmt.Println(readErr)
		w.WriteHeader(500)
		buf := bytes.NewBufferString("Error reading response from service: " + service)
		w.Write(buf.Bytes())
		return
	}

	w.Write(responseBody)
	seconds := time.Since(start).Seconds()
	fmt.Printf("[%s] took %f seconds\n", stamp, seconds)
	metrics.GatewayServerlessServedTotal.Inc()
	metrics.GatewayFunctions.Observe(seconds)
}

func makeProxy(metrics metrics.MetricOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.GatewayRequestsTotal.Inc()

		if r.Method == "POST" {
			log.Println(r.Header)
			header := r.Header["X-Function"]
			log.Println(header)

			if len(header) > 0 {
				exists, err := lookupSwarmService(header[0])
				if err != nil {
					log.Fatalln(err)
				}
				if exists == true {
					requestBody, _ := ioutil.ReadAll(r.Body)
					invokeService(w, r, metrics, header[0], requestBody)
				}
			} else {
				requestBody, _ := ioutil.ReadAll(r.Body)
				alexaService := isAlexa(requestBody)
				if len(alexaService.Session.SessionId) > 0 &&
					len(alexaService.Session.Application.ApplicationId) > 0 &&
					len(alexaService.Request.Intent.Name) > 0 {
					fmt.Println("Alexa skill detected")
					invokeService(w, r, metrics, alexaService.Request.Intent.Name, requestBody)
				} else {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Provide an x-function header or a valid Alexa SDK request."))
				}
			}
		}
	}
}

func main() {
	GatewayRequestsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_requests_total",
		Help: "Total amount of HTTP requests to the gateway",
	})
	GatewayServerlessServedTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_serverless_invocation_total",
		Help: "Total amount of serverless function invocations",
	})
	GatewayFunctions := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "gateway_functions",
		Help: "Gateway functions",
	})

	prometheus.Register(GatewayRequestsTotal)
	prometheus.Register(GatewayServerlessServedTotal)
	prometheus.Register(GatewayFunctions)

	r := mux.NewRouter()
	r.HandleFunc("/", makeProxy(metrics.MetricOptions{
		GatewayRequestsTotal:         GatewayRequestsTotal,
		GatewayServerlessServedTotal: GatewayServerlessServedTotal,
		GatewayFunctions:             GatewayFunctions,
	}))

	metricsHandler := metrics.PrometheusHandler()
	r.Handle("/metrics", metricsHandler)

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    8 * time.Second,
		WriteTimeout:   8 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        r,
	}

	log.Fatal(s.ListenAndServe())
}
