package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/alexellis/faas/gateway/metrics"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

// MakeProxy creates a proxy for HTTP web requests which can be routed to a function.
func MakeProxy(metrics metrics.MetricOptions, wildcard bool, c *client.Client, logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "POST" {
			logger.Infoln(r.Header)
			header := r.Header["X-Function"]
			logger.Infoln(header)

			if wildcard == true {
				vars := mux.Vars(r)
				name := vars["name"]
				fmt.Println("invoke by name")
				lookupInvoke(w, r, metrics, name, c, logger)
				defer r.Body.Close()

			} else if len(header) > 0 {
				lookupInvoke(w, r, metrics, header[0], c, logger)
				defer r.Body.Close()
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Provide a named /function URL or an x-function header."))
				defer r.Body.Close()
			}
		}
	}
}

func writeHead(service string, metrics metrics.MetricOptions, code int, w http.ResponseWriter) {
	w.WriteHeader(code)

	metrics.GatewayFunctionInvocation.With(prometheus.Labels{"function_name": service, "code": strconv.Itoa(code)}).Inc()

	// metrics.GatewayFunctionInvocation.WithLabelValues(service).Add(1)
}

func trackTime(then time.Time, metrics metrics.MetricOptions, name string) {
	since := time.Since(then)
	metrics.GatewayFunctionsHistogram.WithLabelValues(name).Observe(since.Seconds())
}

func lookupInvoke(w http.ResponseWriter, r *http.Request, metrics metrics.MetricOptions, name string, c *client.Client, logger *logrus.Logger) {
	exists, err := lookupSwarmService(name, c)

	if err != nil || exists == false {
		if err != nil {
			logger.Fatalln(err)
		}
		writeHead(name, metrics, http.StatusInternalServerError, w)
		w.Write([]byte("Error resolving service."))
		defer r.Body.Close()
		return
	}

	if exists {
		defer trackTime(time.Now(), metrics, name)
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
	stamp := strconv.FormatInt(time.Now().Unix(), 10)

	defer func(when time.Time) {
		seconds := time.Since(when).Seconds()

		fmt.Printf("[%s] took %f seconds\n", stamp, seconds)
		metrics.GatewayFunctionsHistogram.WithLabelValues(service).Observe(seconds)
	}(time.Now())

	// start := time.Now()
	buf := bytes.NewBuffer(requestBody)
	url := "http://" + service + ":" + strconv.Itoa(8080) + "/"
	contentType := r.Header.Get("Content-Type")
	if len(contentType) == 0 {
		contentType = "text/plain"
	}

	fmt.Printf("[%s] Forwarding request [%s] to: %s\n", stamp, contentType, url)

	response, err := http.Post(url, r.Header.Get("Content-Type"), buf)
	if err != nil {
		logger.Infoln(err)
		writeHead(service, metrics, http.StatusInternalServerError, w)
		buf := bytes.NewBufferString("Can't reach service: " + service)
		w.Write(buf.Bytes())
		return
	}

	responseBody, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		fmt.Println(readErr)

		writeHead(service, metrics, http.StatusInternalServerError, w)
		buf := bytes.NewBufferString("Error reading response from service: " + service)
		w.Write(buf.Bytes())
		return
	}

	// Match header for strict services
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

	writeHead(service, metrics, http.StatusOK, w)
	w.Write(responseBody)
}
