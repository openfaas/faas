// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"io/ioutil"
	"net/http/httptest"

	providerTypes "github.com/openfaas/faas-provider/types"
	"github.com/openfaas/faas/gateway/types"
	"github.com/openfaas/faas/gateway/version"
)

// MakeInfoHandler is responsible for display component version information
func MakeInfoHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseRecorder := httptest.NewRecorder()
		h.ServeHTTP(responseRecorder, r)
		upstreamCall := responseRecorder.Result()

		defer upstreamCall.Body.Close()

		var provider *providerTypes.ProviderInfo

		upstreamBody, _ := ioutil.ReadAll(upstreamCall.Body)
		err := json.Unmarshal(upstreamBody, &provider)
		if err != nil {
			log.Printf("Error unmarshalling provider json from body %s. Error %s\n", upstreamBody, err.Error())
		}

		gatewayInfo := &types.GatewayInfo{
			Version: &providerTypes.VersionInfo{
				CommitMessage: version.GitCommitMessage,
				Release:       version.BuildVersion(),
				SHA:           version.GitCommitSHA,
			},
			Provider: provider,
			Arch:     types.Arch,
		}

		jsonOut, marshalErr := json.Marshal(gatewayInfo)
		if marshalErr != nil {
			log.Printf("Error during unmarshal of gateway info request %s\n", marshalErr.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonOut)

	}
}
