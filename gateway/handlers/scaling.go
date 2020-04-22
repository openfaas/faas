// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/openfaas/faas/gateway/scaling"
)

func getNamespace(defaultNamespace, fullName string) (string, string) {
	if index := strings.LastIndex(fullName, "."); index > -1 {
		return fullName[:index], fullName[index+1:]
	}
	return fullName, defaultNamespace
}

// MakeScalingHandler creates handler which can scale a function from
// zero to N replica(s). After scaling the next http.HandlerFunc will
// be called. If the function is not ready after the configured
// amount of attempts / queries then next will not be invoked and a status
// will be returned to the client.
func MakeScalingHandler(next http.HandlerFunc, scaler scaling.FunctionScaler, config scaling.ScalingConfig, defaultNamespace string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		functionName, namespace := getNamespace(defaultNamespace, getServiceName(r.URL.String()))

		res := scaler.Scale(functionName, namespace)

		if !res.Found {
			errStr := fmt.Sprintf("error finding function %s.%s: %s", functionName, namespace, res.Error.Error())
			log.Printf("Scaling: %s\n", errStr)

			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(errStr))
			return
		}

		if res.Error != nil {
			errStr := fmt.Sprintf("error finding function %s.%s: %s", functionName, namespace, res.Error.Error())
			log.Printf("Scaling: %s\n", errStr)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errStr))
			return
		}

		if res.Available {
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("[Scale] function=%s.%s 0=>N timed-out after %fs\n", functionName, namespace, res.Duration.Seconds())
	}
}
