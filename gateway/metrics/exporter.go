// Copyright (c) Alex Ellis 2017
// Copyright (c) 2018 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package metrics

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"log"

	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas/gateway/requests"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter is a prometheus exporter
type Exporter struct {
	metricOptions MetricOptions
	services      []requests.Function
	credentials   *auth.BasicAuthCredentials
}

// NewExporter creates a new exporter for the OpenFaaS gateway metrics
func NewExporter(options MetricOptions, credentials *auth.BasicAuthCredentials) *Exporter {
	return &Exporter{
		metricOptions: options,
		services:      []requests.Function{},
		credentials:   credentials,
	}
}

// Describe is to describe the metrics for Prometheus
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.metricOptions.GatewayFunctionInvocation.Describe(ch)
	e.metricOptions.GatewayFunctionsHistogram.Describe(ch)
	e.metricOptions.ServiceReplicasGauge.Describe(ch)

	e.metricOptions.ServiceMetrics.Counter.Describe(ch)
	e.metricOptions.ServiceMetrics.Histogram.Describe(ch)
}

// Collect collects data to be consumed by prometheus
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.metricOptions.GatewayFunctionInvocation.Collect(ch)
	e.metricOptions.GatewayFunctionsHistogram.Collect(ch)

	e.metricOptions.ServiceReplicasGauge.Reset()
	for _, service := range e.services {
		e.metricOptions.ServiceReplicasGauge.
			WithLabelValues(service.Name).
			Set(float64(service.Replicas))
	}
	e.metricOptions.ServiceReplicasGauge.Collect(ch)

	e.metricOptions.ServiceMetrics.Counter.Collect(ch)
	e.metricOptions.ServiceMetrics.Histogram.Collect(ch)
}

// StartServiceWatcher starts a ticker and collects service replica counts to expose to prometheus
func (e *Exporter) StartServiceWatcher(endpointURL url.URL, metricsOptions MetricOptions, label string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	quit := make(chan struct{})

	timeout := 3 * time.Second

	proxyClient := http.Client{
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

	go func() {
		for {
			select {
			case <-ticker.C:

				get, _ := http.NewRequest(http.MethodGet, endpointURL.String()+"system/functions", nil)
				if e.credentials != nil {
					get.SetBasicAuth(e.credentials.User, e.credentials.Password)
				}

				services := []requests.Function{}
				res, err := proxyClient.Do(get)
				if err != nil {
					log.Println(err)
					continue
				}
				bytesOut, readErr := ioutil.ReadAll(res.Body)
				if readErr != nil {
					log.Println(err)
					continue
				}
				unmarshalErr := json.Unmarshal(bytesOut, &services)
				if unmarshalErr != nil {
					log.Println(err)
					continue
				}

				e.services = services

				break
			case <-quit:
				return
			}
		}
	}()
}
