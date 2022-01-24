// Copyright (c) Alex Ellis 2017
// Copyright (c) 2018 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package metrics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"log"

	"github.com/openfaas/faas-provider/auth"
	types "github.com/openfaas/faas-provider/types"
	"github.com/openfaas/faas/gateway/scaling"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter is a prometheus exporter
type Exporter struct {
	metricOptions     MetricOptions
	services          []types.FunctionStatus
	credentials       *auth.BasicAuthCredentials
	FunctionNamespace string
}

// NewExporter creates a new exporter for the OpenFaaS gateway metrics
func NewExporter(options MetricOptions, credentials *auth.BasicAuthCredentials, namespace string) *Exporter {
	return &Exporter{
		metricOptions:     options,
		services:          []types.FunctionStatus{},
		credentials:       credentials,
		FunctionNamespace: namespace,
	}
}

// Describe is to describe the metrics for Prometheus
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {

	e.metricOptions.GatewayFunctionInvocation.Describe(ch)
	e.metricOptions.GatewayFunctionsHistogram.Describe(ch)
	e.metricOptions.ServiceReplicasGauge.Describe(ch)
	e.metricOptions.GatewayFunctionInvocationStarted.Describe(ch)
	e.metricOptions.ServiceTargetLoadGauge.Describe(ch)

	e.metricOptions.ServiceMetrics.Counter.Describe(ch)
	e.metricOptions.ServiceMetrics.Histogram.Describe(ch)
}

// Collect collects data to be consumed by prometheus
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.metricOptions.GatewayFunctionInvocation.Collect(ch)
	e.metricOptions.GatewayFunctionsHistogram.Collect(ch)

	e.metricOptions.GatewayFunctionInvocationStarted.Collect(ch)

	e.metricOptions.ServiceReplicasGauge.Reset()
	e.metricOptions.ServiceTargetLoadGauge.Reset()

	for _, service := range e.services {
		var serviceName string
		if len(service.Namespace) > 0 {
			serviceName = fmt.Sprintf("%s.%s", service.Name, service.Namespace)
		} else {
			serviceName = service.Name
		}

		// Set current replica count
		e.metricOptions.ServiceReplicasGauge.
			WithLabelValues(serviceName).
			Set(float64(service.Replicas))

		// Set minimum replicas
		minReplicas := scaling.DefaultMinReplicas
		if service.Labels != nil {
			a := *service.Labels
			if v, ok := a[scaling.MinScaleLabel]; ok && len(v) > 0 {
				val, _ := strconv.Atoi(v)
				minReplicas = val
			}
		}

		e.metricOptions.ServiceMinReplicasGauge.
			WithLabelValues(serviceName).
			Set(float64(minReplicas))

		// Set scale type
		scaleType := scaling.DefaultTypeScale
		if service.Labels != nil {
			a := *service.Labels
			if v, ok := a[scaling.ScaleTypeLabel]; ok && len(v) > 0 {
				scaleType = v
			}
		}

		// Set target load
		targetScale := scaling.DefaultTargetLoad
		if service.Labels != nil {
			a := *service.Labels
			if v, ok := a[scaling.TargetLoadLabel]; ok && len(v) > 0 {
				val, _ := strconv.Atoi(v)
				targetScale = val
			}
		}

		e.metricOptions.ServiceTargetLoadGauge.
			WithLabelValues(serviceName, scaleType).
			Set(float64(targetScale))

	}

	e.metricOptions.ServiceReplicasGauge.Collect(ch)
	e.metricOptions.ServiceMinReplicasGauge.Collect(ch)
	e.metricOptions.ServiceTargetLoadGauge.Collect(ch)

	e.metricOptions.ServiceMetrics.Counter.Collect(ch)
	e.metricOptions.ServiceMetrics.Histogram.Collect(ch)
}

// StartServiceWatcher starts a ticker and collects service replica counts to expose to prometheus
func (e *Exporter) StartServiceWatcher(endpointURL url.URL, metricsOptions MetricOptions, label string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:

				namespaces, err := e.getNamespaces(endpointURL)
				if err != nil {
					log.Println(err)
				}

				services := []types.FunctionStatus{}

				// Providers like faasd for instance have no namespaces.
				if len(namespaces) == 0 {
					services, err = e.getFunctions(endpointURL, e.FunctionNamespace)
					if err != nil {
						log.Println(err)
						continue
					}
					e.services = services
				} else {
					for _, namespace := range namespaces {
						nsServices, err := e.getFunctions(endpointURL, namespace)
						if err != nil {
							log.Println(err)
							continue
						}
						services = append(services, nsServices...)
					}
				}

				e.services = services

				break
			case <-quit:
				return
			}
		}
	}()
}

func (e *Exporter) getHTTPClient(timeout time.Duration) http.Client {

	return http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 0,
			}).DialContext,
			MaxIdleConns:          1,
			DisableKeepAlives:     true,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}
}

func (e *Exporter) getFunctions(endpointURL url.URL, namespace string) ([]types.FunctionStatus, error) {
	timeout := 3 * time.Second
	proxyClient := e.getHTTPClient(timeout)

	endpointURL.Path = path.Join(endpointURL.Path, "/system/functions")
	if len(namespace) > 0 {
		q := endpointURL.Query()
		q.Set("namespace", namespace)
		endpointURL.RawQuery = q.Encode()
	}

	get, _ := http.NewRequest(http.MethodGet, endpointURL.String(), nil)
	if e.credentials != nil {
		get.SetBasicAuth(e.credentials.User, e.credentials.Password)
	}

	services := []types.FunctionStatus{}
	res, err := proxyClient.Do(get)
	if err != nil {
		return services, err
	}

	bytesOut, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return services, readErr
	}

	unmarshalErr := json.Unmarshal(bytesOut, &services)
	if unmarshalErr != nil {
		return services, unmarshalErr
	}
	return services, nil
}

func (e *Exporter) getNamespaces(endpointURL url.URL) ([]string, error) {
	namespaces := []string{}
	endpointURL.Path = path.Join(endpointURL.Path, "system/namespaces")

	get, _ := http.NewRequest(http.MethodGet, endpointURL.String(), nil)
	if e.credentials != nil {
		get.SetBasicAuth(e.credentials.User, e.credentials.Password)
	}

	timeout := 3 * time.Second
	proxyClient := e.getHTTPClient(timeout)

	res, err := proxyClient.Do(get)
	if err != nil {
		return namespaces, err
	}

	if res.StatusCode == http.StatusNotFound {
		return namespaces, nil
	}

	bytesOut, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return namespaces, readErr
	}

	unmarshalErr := json.Unmarshal(bytesOut, &namespaces)
	if unmarshalErr != nil {
		return namespaces, unmarshalErr
	}
	return namespaces, nil
}
