package handler

import (
	"fmt"
	"log"
	"sync"
)

// CreateNATSQueue ready for asynchronous processing
func CreateNATSQueue(address string, port int, clientConfig NATSConfig) (*NATSQueue, error) {
	var err error
	natsURL := fmt.Sprintf("nats://%s:%d", address, port)
	log.Printf("Opening connection to %s\n", natsURL)

	clientID := clientConfig.GetClientID()
	clusterID := "faas-cluster"

	queue1 := NATSQueue{
		ClientID:       clientID,
		ClusterID:      clusterID,
		NATSURL:        natsURL,
		Topic:          "faas-request",
		maxReconnect:   clientConfig.GetMaxReconnect(),
		reconnectDelay: clientConfig.GetReconnectDelay(),
		ncMutex:        &sync.RWMutex{},
	}

	err = queue1.connect()

	return &queue1, err
}
