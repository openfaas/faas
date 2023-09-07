package scaling

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/openfaas/faas-provider/types"
)

const (
	// DefaultMinReplicas is the minimal amount of replicas for a service.
	DefaultMinReplicas = 1

	// DefaultMaxReplicas is the amount of replicas a service will auto-scale up to.
	DefaultMaxReplicas = 5

	// DefaultScalingFactor is the defining proportion for the scaling increments.
	DefaultScalingFactor = 10

	DefaultTypeScale = "rps"

	// MinScaleLabel label indicating min scale for a function
	MinScaleLabel = "com.openfaas.scale.min"

	// MaxScaleLabel label indicating max scale for a function
	MaxScaleLabel = "com.openfaas.scale.max"

	// ScalingFactorLabel label indicates the scaling factor for a function
	ScalingFactorLabel = "com.openfaas.scale.factor"
)

func MakeHorizontalScalingHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.Body == nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		scaleRequest := types.ScaleServiceRequest{}
		if err := json.Unmarshal(body, &scaleRequest); err != nil {
			http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
			return
		}

		if scaleRequest.Replicas < 1 {
			scaleRequest.Replicas = 1
		}

		if scaleRequest.Replicas > DefaultMaxReplicas {
			scaleRequest.Replicas = DefaultMaxReplicas
		}

		upstreamReq, _ := json.Marshal(scaleRequest)
		// Restore the io.ReadCloser to its original state
		r.Body = io.NopCloser(bytes.NewBuffer(upstreamReq))

		next.ServeHTTP(w, r)
	}
}
