// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package plugin

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"fmt"

	"io/ioutil"

	"github.com/openfaas/faas/gateway/handlers"
	"github.com/openfaas/faas/gateway/requests"
)

// NewExternalServiceQuery proxies service queries to external plugin via HTTP
func NewExternalServiceQuery(externalURL url.URL) handlers.ServiceQuery {
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

	return ExternalServiceQuery{
		URL:         externalURL,
		ProxyClient: proxyClient,
	}
}

// ExternalServiceQuery proxies service queries to external plugin via HTTP
type ExternalServiceQuery struct {
	URL         url.URL
	ProxyClient http.Client
}

// GetReplicas replica count for function
func (s ExternalServiceQuery) GetReplicas(serviceName string) (uint64, uint64, uint64, error) {
	var err error
	function := requests.Function{}

	urlPath := fmt.Sprintf("%ssystem/function/%s", s.URL.String(), serviceName)

	req, _ := http.NewRequest(http.MethodGet, urlPath, nil)

	res, err := s.ProxyClient.Do(req)

	if err != nil {

		log.Println(urlPath, err)
	} else {

		if res.Body != nil {
			defer res.Body.Close()
		}

		if res.StatusCode == http.StatusOK {
			bytesOut, _ := ioutil.ReadAll(res.Body)
			err = json.Unmarshal(bytesOut, &function)
			if err != nil {
				log.Println(urlPath, err)
			}
		}
	}

	maxReplicas := uint64(handlers.DefaultMaxReplicas)
	minReplicas := uint64(1)

	if function.Labels != nil {
		labels := *function.Labels
		minScale := labels[handlers.MinScaleLabel]
		maxScale := labels[handlers.MaxScaleLabel]

		if len(minScale) > 0 {
			labelValue, err := strconv.Atoi(minScale)
			if err != nil {
				log.Printf("Bad replica count: %s, should be uint", minScale)
			} else {
				minReplicas = uint64(labelValue)
			}
		}

		if len(maxScale) > 0 {
			labelValue, err := strconv.Atoi(maxScale)
			if err != nil {
				log.Printf("Bad replica count: %s, should be uint", maxScale)
			} else {
				maxReplicas = uint64(labelValue)
			}
		}
	}

	return function.Replicas, maxReplicas, minReplicas, err
}

// ScaleServiceRequest request scaling of replica
type ScaleServiceRequest struct {
	ServiceName string `json:"serviceName"`
	Replicas    uint64 `json:"replicas"`
}

// SetReplicas update the replica count
func (s ExternalServiceQuery) SetReplicas(serviceName string, count uint64) error {
	var err error

	scaleReq := ScaleServiceRequest{
		ServiceName: serviceName,
		Replicas:    count,
	}

	requestBody, err := json.Marshal(scaleReq)
	if err != nil {
		return err
	}

	urlPath := fmt.Sprintf("%ssystem/scale-function/%s", s.URL.String(), serviceName)
	req, _ := http.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBody))
	defer req.Body.Close()
	res, err := s.ProxyClient.Do(req)

	if err != nil {
		log.Println(urlPath, err)
	} else {
		if res.Body != nil {
			defer res.Body.Close()
		}
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("error scaling HTTP code %d, %s", res.StatusCode, urlPath)
	}

	return err
}
