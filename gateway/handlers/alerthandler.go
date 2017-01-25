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

func scaleService(req requests.PrometheusAlert, c *client.Client) error {
	var err error
	//Todo: convert to loop / handler.
	serviceName := req.Alerts[0].Labels.FunctionName
	service, _, inspectErr := c.ServiceInspectWithRaw(context.Background(), serviceName)
	if inspectErr == nil {
		var replicas uint64

		if req.Status == "firing" {
			if *service.Spec.Mode.Replicated.Replicas < 20 {
				replicas = *service.Spec.Mode.Replicated.Replicas + uint64(5)
			} else {
				return err
			}
		} else { // Resolved event.
			// Previously decremented by 5, but event only fires once, so set to 1/1.
			if *service.Spec.Mode.Replicated.Replicas > 1 {
				// replicas = *service.Spec.Mode.Replicated.Replicas - uint64(5)
				// if replicas < 1 {
				// replicas = 1
				// }
				// return nil

				replicas = 1
			} else {
				return nil
			}
		}

		log.Printf("Scaling %s to %d replicas.\n", serviceName, replicas)

		service.Spec.Mode.Replicated.Replicas = &replicas
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
			err := scaleService(req, c)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}
	}
}
