package scaling

import (
	"fmt"
	"log"
	"time"

	"github.com/openfaas/faas/gateway/types"
	"golang.org/x/sync/singleflight"
)

// NewFunctionScaler create a new scaler with the specified
// ScalingConfig
func NewFunctionScaler(config ScalingConfig, functionCacher FunctionCacher) FunctionScaler {
	return FunctionScaler{
		Cache:        functionCacher,
		Config:       config,
		SingleFlight: &singleflight.Group{},
	}
}

// FunctionScaler scales from zero
type FunctionScaler struct {
	Cache        FunctionCacher
	Config       ScalingConfig
	SingleFlight *singleflight.Group
}

// FunctionScaleResult holds the result of scaling from zero
type FunctionScaleResult struct {
	Available bool
	Error     error
	Found     bool
	Duration  time.Duration
}

// Scale scales a function from zero replicas to 1 or the value set in
// the minimum replicas metadata
func (f *FunctionScaler) Scale(functionName, namespace string) FunctionScaleResult {
	start := time.Now()

	// First check the cache, if there are available replicas, then the
	// request can be served.
	if cachedResponse, hit := f.Cache.Get(functionName, namespace); hit &&
		cachedResponse.AvailableReplicas > 0 {
		return FunctionScaleResult{
			Error:     nil,
			Available: true,
			Found:     true,
			Duration:  time.Since(start),
		}
	}

	// The wasn't a hit, or there were no available replicas found
	// so query the live endpoint
	getKey := fmt.Sprintf("GetReplicas-%s.%s", functionName, namespace)
	res, err, _ := f.SingleFlight.Do(getKey, func() (interface{}, error) {
		return f.Config.ServiceQuery.GetReplicas(functionName, namespace)
	})

	if err != nil {
		return FunctionScaleResult{
			Error:     err,
			Available: false,
			Found:     false,
			Duration:  time.Since(start),
		}
	}
	if res == nil {
		return FunctionScaleResult{
			Error:     fmt.Errorf("empty response from server"),
			Available: false,
			Found:     false,
			Duration:  time.Since(start),
		}
	}

	// Check if there are available replicas in the live data
	if res.(ServiceQueryResponse).AvailableReplicas > 0 {
		return FunctionScaleResult{
			Error:     nil,
			Available: true,
			Found:     true,
			Duration:  time.Since(start),
		}
	}

	// Store the result of GetReplicas in the cache
	queryResponse := res.(ServiceQueryResponse)
	f.Cache.Set(functionName, namespace, queryResponse)

	// If the desired replica count is 0, then a scale up event
	// is required.
	if queryResponse.Replicas == 0 {
		minReplicas := uint64(1)
		if queryResponse.MinReplicas > 0 {
			minReplicas = queryResponse.MinReplicas
		}

		// In a retry-loop, first query desired replicas, then
		// set them if the value is still at 0.
		scaleResult := types.Retry(func(attempt int) error {

			res, err, _ := f.SingleFlight.Do(getKey, func() (interface{}, error) {
				return f.Config.ServiceQuery.GetReplicas(functionName, namespace)
			})

			if err != nil {
				return err
			}

			// Cache the response
			queryResponse = res.(ServiceQueryResponse)
			f.Cache.Set(functionName, namespace, queryResponse)

			// The scale up is complete because the desired replica count
			// has been set to 1 or more.
			if queryResponse.Replicas > 0 {
				return nil
			}

			// Request a scale up to the minimum amount of replicas
			setKey := fmt.Sprintf("SetReplicas-%s.%s", functionName, namespace)

			if _, err, _ := f.SingleFlight.Do(setKey, func() (interface{}, error) {

				log.Printf("[Scale %d/%d] function=%s 0 => %d requested",
					attempt, int(f.Config.SetScaleRetries), functionName, minReplicas)

				if err := f.Config.ServiceQuery.SetReplicas(functionName, namespace, minReplicas); err != nil {
					return nil, fmt.Errorf("unable to scale function [%s], err: %s", functionName, err)
				}
				return nil, nil
			}); err != nil {
				return err
			}

			return nil

		}, "Scale", int(f.Config.SetScaleRetries), f.Config.FunctionPollInterval)

		if scaleResult != nil {
			return FunctionScaleResult{
				Error:     scaleResult,
				Available: false,
				Found:     true,
				Duration:  time.Since(start),
			}
		}

	}

	// Holding pattern for at least one function replica to be available
	for i := 0; i < int(f.Config.MaxPollCount); i++ {

		res, err, _ := f.SingleFlight.Do(getKey, func() (interface{}, error) {
			return f.Config.ServiceQuery.GetReplicas(functionName, namespace)
		})
		queryResponse := res.(ServiceQueryResponse)

		if err == nil {
			f.Cache.Set(functionName, namespace, queryResponse)
		}

		totalTime := time.Since(start)

		if err != nil {
			return FunctionScaleResult{
				Error:     err,
				Available: false,
				Found:     true,
				Duration:  totalTime,
			}
		}

		if queryResponse.AvailableReplicas > 0 {

			log.Printf("[Ready] function=%s waited for - %.4fs", functionName, totalTime.Seconds())

			return FunctionScaleResult{
				Error:     nil,
				Available: true,
				Found:     true,
				Duration:  totalTime,
			}
		}

		time.Sleep(f.Config.FunctionPollInterval)
	}

	return FunctionScaleResult{
		Error:     nil,
		Available: true,
		Found:     true,
		Duration:  time.Since(start),
	}
}
