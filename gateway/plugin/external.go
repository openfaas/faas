package plugin

import (
	"net/url"

	"github.com/alexellis/faas/gateway/handlers"
)

func NewExternalServiceQuery(externalURL url.URL) handlers.ServiceQuery {
	return ExternalServiceQuery{
		URL: externalURL,
	}
}

type ExternalServiceQuery struct {
	URL url.URL
}

// GetReplicas replica count for function
func (s ExternalServiceQuery) GetReplicas(serviceName string) (uint64, uint64, error) {
	var err error

	return 0, 0, err
}

// SetReplicas update the replica count
func (s ExternalServiceQuery) SetReplicas(serviceName string, count uint64) error {
	var err error

	return err
}
