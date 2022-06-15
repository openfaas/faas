package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	ftypes "github.com/openfaas/faas-provider/types"
)

type NatsClient struct {
	nc        *nats.Conn
	jetStream nats.JetStream

	NatsURL    string
	NatsStream string
}

func NewNatsClient(address string, port int, stream string) *NatsClient {
	natsURL := fmt.Sprintf("nats://%s:%d", address, port)

	return &NatsClient{
		NatsURL:    natsURL,
		NatsStream: stream,
	}
}

func (n *NatsClient) Connect() error {
	nc, err := nats.Connect(n.NatsURL)
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
