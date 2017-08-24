// Copyright (c) Alex Ellis 2017. All rights reserved.
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

	"github.com/alexellis/faas/gateway/requests"
)

func AttachExternalWatcher(endpointURL url.URL, metricsOptions MetricOptions, label string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	quit := make(chan struct{})
	proxyClient := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
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

				for _, service := range services {
					metricsOptions.ServiceReplicasCounter.
						WithLabelValues(service.Name).
						Set(float64(service.Replicas))
				}

				break
			case <-quit:
				return
			}
		}
	}()
}
