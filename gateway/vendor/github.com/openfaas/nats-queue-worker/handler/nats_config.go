package handler

import (
	"os"
	"time"

	"github.com/openfaas/nats-queue-worker/nats"
)

type NATSConfig interface {
	GetClientID() string
	GetMaxReconnect() int
	GetReconnectDelay() time.Duration
}

type DefaultNATSConfig struct {
	maxReconnect   int
	reconnectDelay time.Duration
}

func NewDefaultNATSConfig(maxReconnect int, reconnectDelay time.Duration) DefaultNATSConfig {
	return DefaultNATSConfig{maxReconnect, reconnectDelay}
}

// GetClientID returns the ClientID assigned to this producer/consumer.
func (DefaultNATSConfig) GetClientID() string {
	val, _ := os.Hostname()
	return getClientID(val)
}

func (c DefaultNATSConfig) GetMaxReconnect() int {
	return c.maxReconnect
}

func (c DefaultNATSConfig) GetReconnectDelay() time.Duration {
	return c.reconnectDelay
}

func getClientID(hostname string) string {
	return "faas-publisher-" + nats.GetClientID(hostname)
}
