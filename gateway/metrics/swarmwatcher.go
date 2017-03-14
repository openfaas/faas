package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// AttachSwarmWatcher adds a go-route to monitor the amount of service replicas in the swarm
// matching a 'function' label.
func AttachSwarmWatcher(dockerClient *client.Client, metricsOptions MetricOptions) {
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				serviceFilter := filters.NewArgs()

				options := types.ServiceListOptions{
					Filters: serviceFilter,
				}

				services, err := dockerClient.ServiceList(context.Background(), options)
				if err != nil {
					fmt.Println(err)
				}

				for _, service := range services {
					if len(service.Spec.TaskTemplate.ContainerSpec.Labels["function"]) > 0 {
						metricsOptions.ServiceReplicasCounter.
							WithLabelValues(service.Spec.Name).
							Set(float64(*service.Spec.Mode.Replicated.Replicas))
					}
				}
				break
			case <-quit:
				return
			}
		}
	}()

}
