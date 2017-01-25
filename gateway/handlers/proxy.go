package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/alexellis/faas/gateway/metrics"
	"github.com/alexellis/faas/gateway/requests"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
)

// makeProxy creates a proxy for HTTP web requests which can be routed to a function.
func MakeProxy(metrics metrics.MetricOptions, wildcard bool, c *client.Client, logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.GatewayRequestsTotal.Inc()

		if r.Method == "POST" {
			logger.Infoln(r.Header)
			header := r.Header["X-Function"]
			logger.Infoln(header)

			if wildcard == true {
				vars := mux.Vars(r)
				name := vars["name"]
				fmt.Println("invoke by name")
				lookupInvoke(w, r, metrics, name, c, logger)
			} else if len(header) > 0 {
				lookupInvoke(w, r, metrics, header[0], c, logger)
			} else {
				requestBody, _ := ioutil.ReadAll(r.Body)
				alexaService := IsAlexa(requestBody)
				fmt.Println(alexaService)

				if len(alexaService.Session.SessionId) > 0 &&
					len(alexaService.Session.Application.ApplicationId) > 0 &&
					len(alexaService.Request.Intent.Name) > 0 {

					fmt.Println("Alexa SDK request found")
					fmt.Printf("SessionId=%s, Intent=%s, AppId=%s\n", alexaService.Session.SessionId, alexaService.Request.Intent.Name, alexaService.Session.Application.ApplicationId)

					invokeService(w, r, metrics, alexaService.Request.Intent.Name, requestBody, logger)
				} else {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Provide an x-function header or a valid Alexa SDK request."))
				}
			}
		}
	}
}

func IsAlexa(requestBody []byte) requests.AlexaRequestBody {
	body := requests.AlexaRequestBody{}
	buf := bytes.NewBuffer(requestBody)
	// fmt.Println(buf)
	str := buf.String()
	parts := strings.Split(str, "sessionId")
	if len(parts) > 1 {
		json.Unmarshal(requestBody, &body)
	}
	return body
}

func lookupInvoke(w http.ResponseWriter, r *http.Request, metrics metrics.MetricOptions, name string, c *client.Client, logger *logrus.Logger) {
	exists, err := lookupSwarmService(name, c)
	if err != nil || exists == false {
		if err != nil {
			logger.Fatalln(err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error resolving service."))
	}
	if exists == true {
		requestBody, _ := ioutil.ReadAll(r.Body)
		invokeService(w, r, metrics, name, requestBody, logger)
	}
}

func lookupSwarmService(serviceName string, c *client.Client) (bool, error) {
	fmt.Printf("Resolving: '%s'\n", serviceName)
	serviceFilter := filters.NewArgs()
	serviceFilter.Add("name", serviceName)
	services, err := c.ServiceList(context.Background(), types.ServiceListOptions{Filters: serviceFilter})

	return len(services) > 0, err
}

func invokeService(w http.ResponseWriter, r *http.Request, metrics metrics.MetricOptions, service string, requestBody []byte, logger *logrus.Logger) {
	metrics.GatewayFunctionInvocation.WithLabelValues(service).Add(1)

	stamp := strconv.FormatInt(time.Now().Unix(), 10)

	start := time.Now()
	buf := bytes.NewBuffer(requestBody)
	url := "http://" + service + ":" + strconv.Itoa(8080) + "/"
	fmt.Printf("[%s] Forwarding request to: %s\n", stamp, url)

	response, err := http.Post(url, "text/plain", buf)
	if err != nil {
		logger.Infoln(err)
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
