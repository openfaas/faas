package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"strconv"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// MakeAlertHandler handles alerts from Prometheus Alertmanager
func MakeAlertHandler(c *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Alert received.")
		body, readErr := ioutil.ReadAll(r.Body)

		log.Println(string(body))

		if readErr != nil {
			log.Println(readErr)
			return
		}

		var req requests.PrometheusAlert
		err := json.Unmarshal(body, &req)
		if err != nil {
			log.Println(err)
			return
		}

		if len(req.Alerts) > 0 {
			if err := scaleService(req, c); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}
	}
}

// CalculateReplicas decides what replica count to set depending on a Prometheus alert
func CalculateReplicas(status string, currentReplicas uint64, maxReplicas uint64) uint64 {
	newReplicas := currentReplicas

	if status == "firing" {
		if currentReplicas == 1 {
			newReplicas = 5
		} else {
			if currentReplicas+5 > maxReplicas {
				newReplicas = maxReplicas
			} else {
				newReplicas = currentReplicas + 5
			}
		}
	} else { // Resolved event.
		newReplicas = 1
	}
	return newReplicas
}

func scaleService(req requests.PrometheusAlert, c *client.Client) error {
	var err error
	serviceName := req.Alerts[0].Labels.FunctionName

	if len(serviceName) > 0 {

		service, _, inspectErr := c.ServiceInspectWithRaw(context.Background(), serviceName)
		if inspectErr == nil {

			currentReplicas := *service.Spec.Mode.Replicated.Replicas
			status := req.Status

			replicaLabel := service.Spec.TaskTemplate.ContainerSpec.Labels["com.faas.max_replicas"]
			maxReplicas := 20
			if len(replicaLabel) > 0 {
				maxReplicas, err = strconv.Atoi(replicaLabel)
				if err != nil {
					log.Printf("Bad replica count: %s, should be uint.\n", replicaLabel)
				}
			}
			newReplicas := CalculateReplicas(status, currentReplicas, uint64(maxReplicas))

			if newReplicas == currentReplicas {
				return nil
			}

			log.Printf("Scaling %s to %d replicas.\n", serviceName, newReplicas)
			service.Spec.Mode.Replicated.Replicas = &newReplicas
			updateOpts := types.ServiceUpdateOptions{}
			updateOpts.RegistryAuthFrom = types.RegistryAuthFromSpec

			response, updateErr := c.ServiceUpdate(context.Background(), service.ID, service.Version, service.Spec, updateOpts)
			if updateErr != nil {
				err = updateErr
			}
			log.Println(response)

		} else {
			err = inspectErr
		}
	}
	return err
}
