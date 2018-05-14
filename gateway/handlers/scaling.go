// Copyright (c) OpenFaaS Project. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/openfaas/faas/gateway/requests"
)

// ScalingConfig for scaling behaviours
type ScalingConfig struct {
	MaxPollCount         uint
	FunctionPollInterval time.Duration
	CacheExpiry          time.Duration
}

// MakeScalingHandler creates handler which can scale a function from
// zero to 1 replica(s).
func MakeScalingHandler(next http.HandlerFunc, upstream http.HandlerFunc, config ScalingConfig) http.HandlerFunc {
	cache := FunctionCache{
		Cache:  make(map[string]*FunctionMeta),
		Expiry: config.CacheExpiry,
	}

	return func(w http.ResponseWriter, r *http.Request) {

		functionName := getServiceName(r.URL.String())

		if replicas, hit := cache.Get(functionName); hit && replicas > 0 {
			next.ServeHTTP(w, r)
			return
		}

		replicas, code, err := getReplicas(functionName, upstream)
		cache.Set(functionName, replicas)

		if err != nil {
			var errStr string
			if code == http.StatusNotFound {
				errStr = fmt.Sprintf("unable to find function: %s", functionName)

			} else {
				errStr = fmt.Sprintf("error finding function %s: %s", functionName, err.Error())
			}

			log.Printf(errStr)
			w.WriteHeader(code)
			w.Write([]byte(errStr))
			return
		}

		if replicas == 0 {
			minReplicas := uint64(1)

			err := scaleFunction(functionName, minReplicas, upstream)
			if err != nil {
				errStr := fmt.Errorf("unable to scale function [%s], err: %s", functionName, err)
				log.Printf(errStr.Error())

				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(errStr.Error()))
				return
			}

			for i := 0; i < int(config.MaxPollCount); i++ {
				replicas, _, err := getReplicas(functionName, upstream)
				cache.Set(functionName, replicas)

				if err != nil {
					errStr := fmt.Sprintf("error: %s", err.Error())
					log.Printf(errStr)

					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(errStr))
					return
				}

				if replicas > 0 {
					break
				}

				time.Sleep(config.FunctionPollInterval)
			}
		}

		next.ServeHTTP(w, r)
	}
}

func getReplicas(functionName string, upstream http.HandlerFunc) (uint64, int, error) {

	replicasQuery, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/system/function/%s", functionName), nil)
	rr := httptest.NewRecorder()

	upstream.ServeHTTP(rr, replicasQuery)
	if rr.Code != 200 {
		log.Printf("error, query replicas status: %d", rr.Code)

		var errBody string
		if rr.Body != nil {
			errBody = string(rr.Body.String())
		}

		return 0, rr.Code, fmt.Errorf("unable to query function: %s", string(errBody))
	}

	replicaBytes, _ := ioutil.ReadAll(rr.Body)
	replicaResult := requests.Function{}
	json.Unmarshal(replicaBytes, &replicaResult)

	return replicaResult.AvailableReplicas, rr.Code, nil
}

func scaleFunction(functionName string, minReplicas uint64, upstream http.HandlerFunc) error {
	scaleReq := ScaleServiceRequest{
		Replicas:    minReplicas,
		ServiceName: functionName,
	}

	scaleBytesOut, _ := json.Marshal(scaleReq)
	scaleBytesOutBody := bytes.NewBuffer(scaleBytesOut)
	setReplicasReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/system/scale-function/%s", functionName), scaleBytesOutBody)

	rr := httptest.NewRecorder()
	upstream.ServeHTTP(rr, setReplicasReq)

	if rr.Code != 200 {
		return fmt.Errorf("scale to 1 replica status: %d", rr.Code)
	}

	return nil
}

// ScaleServiceRequest request to scale a function
type ScaleServiceRequest struct {
	ServiceName string `json:"serviceName"`
	Replicas    uint64 `json:"replicas"`
}
