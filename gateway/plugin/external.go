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

const maxReplicas = 40

// GetReplicas replica count for function
func (s ExternalServiceQuery) GetReplicas(serviceName string) (uint64, uint64, error) {
	var err error
	function := requests.Function{}

	urlPath := fmt.Sprintf("%ssystem/function/%s", s.URL.String(), serviceName)
	req, _ := http.NewRequest("GET", urlPath, nil)
	res, err := s.ProxyClient.Do(req)
	if err != nil {
		log.Println(urlPath, err)
	}

	if res.StatusCode == 200 {
		if res.Body != nil {
			defer res.Body.Close()
			bytesOut, _ := ioutil.ReadAll(res.Body)
			err = json.Unmarshal(bytesOut, &function)
			if err != nil {
				log.Println(urlPath, err)
			}
		}
	}

	max := uint64(maxReplicas)

	return function.Replicas, max, err
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
	req, _ := http.NewRequest("POST", urlPath, bytes.NewReader(requestBody))
	defer req.Body.Close()
	res, err := s.ProxyClient.Do(req)

	defer res.Body.Close()

	if err != nil {
		log.Println(urlPath, err)
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("error scaling HTTP code %d, %s", res.StatusCode, urlPath)
	}

	return err
}
