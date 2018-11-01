package scaling

import (
	"fmt"
	"log"
	"time"
)

// NewFunctionScaler create a new scaler with the specified
// ScalingConfig
func NewFunctionScaler(config ScalingConfig) FunctionScaler {
	cache := FunctionCache{
		Cache:  make(map[string]*FunctionMeta),
		Expiry: config.CacheExpiry,
	}

	return FunctionScaler{
		Cache:  &cache,
		Config: config,
	}
}

// FunctionScaler scales from zero
type FunctionScaler struct {
	Cache  *FunctionCache
	Config ScalingConfig
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
func (f *FunctionScaler) Scale(functionName string) FunctionScaleResult {
	start := time.Now()
	queryResponse, err := f.Config.ServiceQuery.GetReplicas(functionName)

	if err != nil {
		return FunctionScaleResult{
			Error:     err,
			Available: false,
			Found:     false,
			Duration:  time.Since(start),
		}
	}

	f.Cache.Set(functionName, queryResponse)

	if queryResponse.AvailableReplicas == 0 {
		minReplicas := uint64(1)
		if queryResponse.MinReplicas > 0 {
			minReplicas = queryResponse.MinReplicas
		}

		log.Printf("[Scale] function=%s 0 => %d requested", functionName, minReplicas)

		setScaleErr := f.Config.ServiceQuery.SetReplicas(functionName, minReplicas)
		if setScaleErr != nil {
			return FunctionScaleResult{
				Error:     fmt.Errorf("unable to scale function [%s], err: %s", functionName, err),
				Available: false,
				Found:     true,
				Duration:  time.Since(start),
			}
		}

		for i := 0; i < int(f.Config.MaxPollCount); i++ {
			queryResponse, err := f.Config.ServiceQuery.GetReplicas(functionName)
			f.Cache.Set(functionName, queryResponse)
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

				log.Printf("[Scale] function=%s 0 => %d successful - %f seconds", functionName, queryResponse.AvailableReplicas, totalTime.Seconds())

				return FunctionScaleResult{
					Error:     nil,
					Available: true,
					Found:     true,
					Duration:  totalTime,
				}
			}

			time.Sleep(f.Config.FunctionPollInterval)
		}
	}

	return FunctionScaleResult{
		Error:     nil,
		Available: true,
		Found:     true,
		Duration:  time.Since(start),
	}
}
