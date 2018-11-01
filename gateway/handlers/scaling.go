// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/openfaas/faas/gateway/scaling"
)

// MakeScalingHandler creates handler which can scale a function from
// zero to N replica(s). After scaling the next http.HandlerFunc will
// be called. If the function is not ready after the configured
// amount of attempts / queries then next will not be invoked and a status
// will be returned to the client.
func MakeScalingHandler(next http.HandlerFunc, config scaling.ScalingConfig) http.HandlerFunc {

	scaler := scaling.NewFunctionScaler(config)

	return func(w http.ResponseWriter, r *http.Request) {

		functionName := getServiceName(r.URL.String())
		res := scaler.Scale(functionName)

		if !res.Found {
			errStr := fmt.Sprintf("error finding function %s: %s", functionName, res.Error.Error())
			log.Printf("Scaling: %s", errStr)

			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(errStr))
			return
		}

		if res.Error != nil {
			errStr := fmt.Sprintf("error finding function %s: %s", functionName, res.Error.Error())
			log.Printf("Scaling: %s", errStr)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errStr))
			return
		}

		if res.Available {
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("[Scale] function=%s 0=>N timed-out after %f seconds", functionName, res.Duration.Seconds())
	}
}
