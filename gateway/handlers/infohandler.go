package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/openfaas/faas/gateway/types"
	"github.com/openfaas/faas/gateway/version"
)

// MakeInfoHandler is responsible for display component version information
func MakeInfoHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sw := types.NewStringResponseWriter()
		h.ServeHTTP(sw, r)

		log.Printf("Body: %s", sw.Body())
		provider := make(map[string]interface{})
		providerVersion := &types.VersionInfo{}

		err := json.Unmarshal(sw.Body(), &provider)
		if err != nil {
			log.Printf("Error unmarshalling provider json. Got %s. Error %s\n", string(sw.Body()), err.Error())
		}

		versionMap := provider["version"].(map[string]interface{})
		providerVersion.SHA = versionMap["sha"].(string)
		providerVersion.Release = versionMap["release"].(string)

		gatewayInfo := &types.GatewayInfo{
			Version: &types.VersionInfo{
				CommitMessage: version.GitCommitMessage,
				Release:       version.BuildVersion(),
				SHA:           version.GitCommitSHA,
			},
			Provider: &types.ProviderInfo{
				Version:       providerVersion,
				Name:          provider["provider"].(string),
				Orchestration: provider["orchestration"].(string),
			},
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
