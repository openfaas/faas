package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	ftypes "github.com/openfaas/faas-provider/types"
)

type NatsClient struct {
	nc        *nats.Conn
	jetStream nats.JetStream

	NatsURL        string
	NatsStream     string
	MaxReconnect   int
	ReconnectDelay time.Duration
}

func NewNatsClient(address string, port int, stream string, maxReconnect int, reconnectDelay time.Duration) *NatsClient {
	natsURL := fmt.Sprintf("nats://%s:%d", address, port)

	return &NatsClient{
		NatsURL:        natsURL,
		NatsStream:     stream,
		MaxReconnect:   maxReconnect,
		ReconnectDelay: reconnectDelay,
	}
}

func (n *NatsClient) Connect() error {
	nc, err := nats.Connect(n.NatsURL, nats.MaxReconnects(n.MaxReconnect), nats.ReconnectWait(n.ReconnectDelay),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("NATS disconnected: %q", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %s", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Panicf("NATS connection closed: %q", nc.LastError())
		}),
	)
	if err != nil {
		return fmt.Errorf("error connecting to NATS on %s: %v", n.NatsURL, err)
	}
	n.nc = nc

	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("error connecting to JetStream: %v", err)
	}

	n.jetStream = js
	return nil
}

func (n *NatsClient) Queue(req *ftypes.QueueRequest) error {
	fmt.Printf("NatsStream - submitting request: %s.\n", req.Function)

	out, err := json.Marshal(req)
	if err != nil {
		log.Println(err)
	}

	streamName := n.NatsStream
	if len(req.QueueName) > 0 {
		streamName = req.QueueName
	}

	_, err = n.jetStream.Publish(streamName, out)

	return err
}
