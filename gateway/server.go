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

func lookupSwarmService(serviceName string, c *client.Client) (bool, error) {
	fmt.Printf("Resolving: '%s'\n", serviceName)
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
	if len(parts) > 1 {
		json.Unmarshal(requestBody, &body)
	}
	return body
}

func invokeService(w http.ResponseWriter, r *http.Request, metrics metrics.MetricOptions, service string, requestBody []byte) {
	metrics.GatewayFunctionInvocation.WithLabelValues(service).Add(1)

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

	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
	seconds := time.Since(start).Seconds()
	fmt.Printf("[%s] took %f seconds\n", stamp, seconds)
	metrics.GatewayServerlessServedTotal.Inc()
	metrics.GatewayFunctions.Observe(seconds)
}

func lookupInvoke(w http.ResponseWriter, r *http.Request, metrics metrics.MetricOptions, name string, c *client.Client) {
	exists, err := lookupSwarmService(name, c)
	if err != nil || exists == false {
		if err != nil {
			log.Fatalln(err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error resolving service."))
	}
	if exists == true {
		requestBody, _ := ioutil.ReadAll(r.Body)
		invokeService(w, r, metrics, name, requestBody)
	}
}

func makeProxy(metrics metrics.MetricOptions, wildcard bool, c *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.GatewayRequestsTotal.Inc()

		if r.Method == "POST" {
			log.Println(r.Header)
			header := r.Header["X-Function"]
			log.Println(header)
			fmt.Println(wildcard)

			if wildcard == true {
				vars := mux.Vars(r)
				name := vars["name"]
				fmt.Println("invoke by name")
				lookupInvoke(w, r, metrics, name, c)
			} else if len(header) > 0 {
				lookupInvoke(w, r, metrics, header[0], c)
			} else {
				requestBody, _ := ioutil.ReadAll(r.Body)
				alexaService := isAlexa(requestBody)
				fmt.Println(alexaService)

				if len(alexaService.Session.SessionId) > 0 &&
					len(alexaService.Session.Application.ApplicationId) > 0 &&
					len(alexaService.Request.Intent.Name) > 0 {

					fmt.Println("Alexa SDK request found")
					fmt.Printf("SessionId=%s, Intent=%s, AppId=%s\n", alexaService.Session.SessionId, alexaService.Request.Intent.Name, alexaService.Session.Application.ApplicationId)

					invokeService(w, r, metrics, alexaService.Request.Intent.Name, requestBody)
				} else {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Provide an x-function header or a valid Alexa SDK request."))
				}
			}
		}
	}
}

func makeAlertHandler(c *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(c)
		// Todo: parse alert, validate alert and scale up or down function

		fmt.Println("Alert received.")
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
	GatewayFunctionInvocation := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_function_invocation_total",
			Help: "Individual function metrics",
		},
		[]string{"function_name"},
	)

	prometheus.Register(GatewayRequestsTotal)
	prometheus.Register(GatewayServerlessServedTotal)
	prometheus.Register(GatewayFunctions)
	prometheus.Register(GatewayFunctionInvocation)

	metricsOptions := metrics.MetricOptions{
		GatewayRequestsTotal:         GatewayRequestsTotal,
		GatewayServerlessServedTotal: GatewayServerlessServedTotal,
		GatewayFunctions:             GatewayFunctions,
		GatewayFunctionInvocation:    GatewayFunctionInvocation,
	}

	var c *client.Client
	var err error
	c, err = client.NewEnvClient()
	if err != nil {
		log.Fatal("Error with Docker client.")
	}

	r := mux.NewRouter()
	r.HandleFunc("/function/{name:[a-zA-Z_]+}", makeProxy(metricsOptions, true, c))

	r.HandleFunc("/system/alert", makeAlertHandler(c))

	r.HandleFunc("/", makeProxy(metricsOptions, false, c))

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
