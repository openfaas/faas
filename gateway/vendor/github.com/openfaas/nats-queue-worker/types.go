package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	stan "github.com/nats-io/go-nats-streaming"
)

// AsyncReport is the report from a function executed on a queue worker.
type AsyncReport struct {
	FunctionName string  `json:"name"`
	StatusCode   int     `json:"statusCode"`
	TimeTaken    float64 `json:"timeTaken"`
}

// NATSQueue represents a subscription to NATS Streaming
type NATSQueue struct {
	clusterID string
	clientID  string
	natsURL   string

	maxReconnect   int
	reconnectDelay time.Duration
	conn           stan.Conn
	connMutex      *sync.RWMutex
	quitCh         chan struct{}

	subject        string
	qgroup         string
	durable        string
	ackWait        time.Duration
	messageHandler func(*stan.Msg)
	startOption    stan.SubscriptionOption
	maxInFlight    stan.SubscriptionOption
	subscription   stan.Subscription
}

// connect creates a subscription to NATS Streaming
func (q *NATSQueue) connect() error {
	log.Printf("Connect: %s\n", q.natsURL)

	nc, err := stan.Connect(
		q.clusterID,
		q.clientID,
		stan.NatsURL(q.natsURL),
		stan.SetConnectionLostHandler(func(conn stan.Conn, err error) {
			log.Printf("Disconnected from %s\n", q.natsURL)

			q.reconnect()
		}),
	)
	if err != nil {
		return fmt.Errorf("can't connect to %s: %v", q.natsURL, err)
	}

	q.connMutex.Lock()
	defer q.connMutex.Unlock()

	q.conn = nc

	log.Printf("Subscribing to: %s at %s\n", q.subject, q.natsURL)
	log.Println("Wait for ", q.ackWait)

	subscription, err := q.conn.QueueSubscribe(
		q.subject,
		q.qgroup,
		q.messageHandler,
		stan.DurableName(q.durable),
		stan.AckWait(q.ackWait),
		q.startOption,
		q.maxInFlight,
	)

	if err != nil {
		return fmt.Errorf("couldn't subscribe to %s at %s. Error: %v", q.subject, q.natsURL, err)
	}

	log.Printf(
		"Listening on [%s], clientID=[%s], qgroup=[%s] durable=[%s]\n",
		q.subject,
		q.clientID,
		q.qgroup,
		q.durable,
	)

	q.subscription = subscription

	return nil
}

func (q *NATSQueue) reconnect() {
	log.Printf("Reconnect\n")

	for i := 0; i < q.maxReconnect; i++ {
		select {
		case <-time.After(time.Duration(i) * q.reconnectDelay):
			if err := q.connect(); err == nil {
				log.Printf("Reconnecting (%d/%d) to %s succeeded\n", i+1, q.maxReconnect, q.natsURL)

				return
			}

			nextTryIn := (time.Duration(i+1) * q.reconnectDelay).String()

			log.Printf("Reconnecting (%d/%d) to %s failed\n", i+1, q.maxReconnect, q.natsURL)
			log.Printf("Waiting %s before next try", nextTryIn)
		case <-q.quitCh:
			log.Println("Received signal to stop reconnecting...")

			return
		}
	}

	log.Printf("Reconnecting limit (%d) reached\n", q.maxReconnect)
}

func (q *NATSQueue) unsubscribe() error {
	q.connMutex.Lock()
	defer q.connMutex.Unlock()

	if q.subscription != nil {
		return fmt.Errorf("q.subscription is nil")
	}

	return q.subscription.Unsubscribe()
}

func (q *NATSQueue) closeConnection() error {
	q.connMutex.Lock()
	defer q.connMutex.Unlock()

	if q.conn == nil {
		return fmt.Errorf("q.conn is nil")
	}

	close(q.quitCh)

	return q.conn.Close()
}
