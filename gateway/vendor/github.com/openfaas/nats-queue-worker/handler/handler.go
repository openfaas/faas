package handler

import (
	"fmt"
	"log"
	"sync"
)

// CreateNATSQueue ready for asynchronous processing
func CreateNATSQueue(address string, port int, clusterName, channel string, clientConfig NATSConfig) (*NATSQueue, error) {
	var err error
	natsURL := fmt.Sprintf("nats://%s:%d", address, port)
	log.Printf("Opening connection to %s\n", natsURL)

	clientID := clientConfig.GetClientID()

	// If 'channel' is empty, use the previous default.
	if channel == "" {
		channel = "faas-request"
	}

	queue1 := NATSQueue{
		ClientID:       clientID,
		ClusterID:      clusterName,
		NATSURL:        natsURL,
		Topic:          channel,
		maxReconnect:   clientConfig.GetMaxReconnect(),
		reconnectDelay: clientConfig.GetReconnectDelay(),
		ncMutex:        &sync.RWMutex{},
	}

	err = queue1.connect()

	return &queue1, err
}
