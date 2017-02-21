package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// CalculateReplicas decides what replica count to set depending on a Prometheus alert
func CalculateReplicas(status string, currentReplicas uint64) uint64 {
	newReplicas := currentReplicas

	if status == "firing" {
		if currentReplicas == 1 {
			newReplicas = 5
		} else {
			if currentReplicas+5 > 20 {
				newReplicas = 20
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
	//Todo: convert to loop / handler.
	serviceName := req.Alerts[0].Labels.FunctionName
	service, _, inspectErr := c.ServiceInspectWithRaw(context.Background(), serviceName)
	if inspectErr == nil {

		currentReplicas := *service.Spec.Mode.Replicated.Replicas
		status := req.Status
		newReplicas := CalculateReplicas(status, currentReplicas)

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

	return err
}

func MakeAlertHandler(c *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Alert received.")
		body, readErr := ioutil.ReadAll(r.Body)
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
