// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/openfaas/faas/gateway/requests"
)

// DefaultMaxReplicas is the amount of replicas a service will auto-scale up to.
const DefaultMaxReplicas = 20

// MinScaleLabel label indicating min scale for a function
const MinScaleLabel = "com.openfaas.scale.min"

// MaxScaleLabel label indicating max scale for a function
const MaxScaleLabel = "com.openfaas.scale.max"

// ServiceQuery provides interface for replica querying/setting
type ServiceQuery interface {
	GetReplicas(service string) (currentReplicas uint64, maxReplicas uint64, minReplicas uint64, err error)
	SetReplicas(service string, count uint64) error
}

// NewSwarmServiceQuery create new Docker Swarm implementation
func NewSwarmServiceQuery(c *client.Client) ServiceQuery {
	return SwarmServiceQuery{
		c: c,
	}
}

// SwarmServiceQuery implementation for Docker Swarm
type SwarmServiceQuery struct {
	c *client.Client
}

// GetReplicas replica count for function
func (s SwarmServiceQuery) GetReplicas(serviceName string) (uint64, uint64, uint64, error) {
	var err error
	var currentReplicas uint64

	maxReplicas := uint64(DefaultMaxReplicas)
	minReplicas := uint64(1)

	opts := types.ServiceInspectOptions{
		InsertDefaults: true,
	}

	service, _, err := s.c.ServiceInspectWithRaw(context.Background(), serviceName, opts)

	if err == nil {
		currentReplicas = *service.Spec.Mode.Replicated.Replicas

		minScale := service.Spec.Annotations.Labels[MinScaleLabel]
		maxScale := service.Spec.Annotations.Labels[MaxScaleLabel]

		if len(maxScale) > 0 {
			labelValue, err := strconv.Atoi(maxScale)
			if err != nil {
				log.Printf("Bad replica count: %s, should be uint", maxScale)
			} else {
				maxReplicas = uint64(labelValue)
			}
		}

		if len(minScale) > 0 {
			labelValue, err := strconv.Atoi(maxScale)
			if err != nil {
				log.Printf("Bad replica count: %s, should be uint", minScale)
			} else {
				minReplicas = uint64(labelValue)
			}
		}
	}

	return currentReplicas, maxReplicas, minReplicas, err
}

// SetReplicas update the replica count
func (s SwarmServiceQuery) SetReplicas(serviceName string, count uint64) error {
	opts := types.ServiceInspectOptions{
		InsertDefaults: true,
	}

	service, _, err := s.c.ServiceInspectWithRaw(context.Background(), serviceName, opts)
	if err == nil {

		service.Spec.Mode.Replicated.Replicas = &count
		updateOpts := types.ServiceUpdateOptions{}
		updateOpts.RegistryAuthFrom = types.RegistryAuthFromSpec

		_, updateErr := s.c.ServiceUpdate(context.Background(), service.ID, service.Version, service.Spec, updateOpts)
		if updateErr != nil {
			err = updateErr
		}
	}

	return err
}

// MakeAlertHandler handles alerts from Prometheus Alertmanager
func MakeAlertHandler(service ServiceQuery) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Println("Alert received.")

		body, readErr := ioutil.ReadAll(r.Body)

		log.Println(string(body))

		if readErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Unable to read alert."))

			log.Println(readErr)
			return
		}

		var req requests.PrometheusAlert
		err := json.Unmarshal(body, &req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Unable to parse alert, bad format."))
			log.Println(err)
			return
		}

		errors := handleAlerts(&req, service)
		if len(errors) > 0 {
			log.Println(errors)
			var errorOutput string
			for d, err := range errors {
				errorOutput += fmt.Sprintf("[%d] %s\n", d, err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errorOutput))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func handleAlerts(req *requests.PrometheusAlert, service ServiceQuery) []error {
	var errors []error
	for _, alert := range req.Alerts {
		if err := scaleService(alert, service); err != nil {
			log.Println(err)
			errors = append(errors, err)
		}
	}

	return errors
}

func scaleService(alert requests.PrometheusInnerAlert, service ServiceQuery) error {
	var err error
	serviceName := alert.Labels.FunctionName

	if len(serviceName) > 0 {
		currentReplicas, maxReplicas, minReplicas, getErr := service.GetReplicas(serviceName)
		if getErr == nil {
			status := alert.Status

			newReplicas := CalculateReplicas(status, currentReplicas, uint64(maxReplicas), minReplicas)

			log.Printf("[Scale] function=%s %d => %d.\n", serviceName, currentReplicas, newReplicas)
			if newReplicas == currentReplicas {
				return nil
			}

			updateErr := service.SetReplicas(serviceName, newReplicas)
			if updateErr != nil {
				err = updateErr
			}
		}
	}
	return err
}

// CalculateReplicas decides what replica count to set depending on current/desired amount
func CalculateReplicas(status string, currentReplicas uint64, maxReplicas uint64, minReplicas uint64) uint64 {
	newReplicas := currentReplicas
	const step = 5

	if status == "firing" {
		if currentReplicas == 1 {
			// First jump is from 1 to "step" i.e. 1->5
			newReplicas = step
		} else {
			if currentReplicas+step > maxReplicas {
				newReplicas = maxReplicas
			} else {
				newReplicas = currentReplicas + step
			}
		}
	} else { // Resolved event.
		newReplicas = minReplicas
	}
	return newReplicas
}
