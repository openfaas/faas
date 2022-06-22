package scaling

import (
	"fmt"
	"log"
	"time"
)

// NewFunctionScaler create a new scaler with the specified
// ScalingConfig
func NewFunctionScaler(config ScalingConfig, functionCacher FunctionCacher) FunctionScaler {
	return FunctionScaler{
		Cache:        functionCacher,
		Config:       config,
		SingleFlight: NewSingleFlight(),
	}
}

// FunctionScaler scales from zero
type FunctionScaler struct {
	Cache        FunctionCacher
	Config       ScalingConfig
	SingleFlight *SingleFlight
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

	if cachedResponse, hit := f.Cache.Get(functionName, namespace); hit &&
		cachedResponse.AvailableReplicas > 0 {
		return FunctionScaleResult{
			Error:     nil,
			Available: true,
			Found:     true,
			Duration:  time.Since(start),
		}
	}
	getKey := fmt.Sprintf("GetReplicas-%s.%s", functionName, namespace)

	res, err := f.SingleFlight.Do(getKey, func() (interface{}, error) {
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

	queryResponse := res.(ServiceQueryResponse)

	f.Cache.Set(functionName, namespace, queryResponse)

	if queryResponse.AvailableReplicas == 0 {
		minReplicas := uint64(1)
		if queryResponse.MinReplicas > 0 {
			minReplicas = queryResponse.MinReplicas
		}

		scaleResult := backoff(func(attempt int) error {

			res, err := f.SingleFlight.Do(getKey, func() (interface{}, error) {
				return f.Config.ServiceQuery.GetReplicas(functionName, namespace)
			})

			if err != nil {
				return err
			}

			queryResponse = res.(ServiceQueryResponse)

			f.Cache.Set(functionName, namespace, queryResponse)

			if queryResponse.Replicas > 0 {
				return nil
			}

			setKey := fmt.Sprintf("SetReplicas-%s.%s", functionName, namespace)

			if _, err := f.SingleFlight.Do(setKey, func() (interface{}, error) {

				log.Printf("[Scale %d] function=%s 0 => %d requested", attempt, functionName, minReplicas)

				if err := f.Config.ServiceQuery.SetReplicas(functionName, namespace, minReplicas); err != nil {
					return nil, fmt.Errorf("unable to scale function [%s], err: %s", functionName, err)
				}
				return nil, nil
			}); err != nil {
				return err
			}

			return nil

		}, int(f.Config.SetScaleRetries), f.Config.FunctionPollInterval)

		if scaleResult != nil {
			return FunctionScaleResult{
				Error:     scaleResult,
				Available: false,
				Found:     true,
				Duration:  time.Since(start),
			}
		}

		for i := 0; i < int(f.Config.MaxPollCount); i++ {

			res, err := f.SingleFlight.Do(getKey, func() (interface{}, error) {
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

				log.Printf("[Scale] function=%s 0 => %d successful - %fs",
					functionName, queryResponse.AvailableReplicas, totalTime.Seconds())

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

type routine func(attempt int) error

func backoff(r routine, attempts int, interval time.Duration) error {
	var err error

	for i := 0; i < attempts; i++ {
		res := r(i)
		if res != nil {
			err = res

			log.Printf("Attempt: %d, had error: %s\n", i, res)
		} else {
			err = nil
			break
		}
		time.Sleep(interval)
	}
	return err
}
