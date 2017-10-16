package handler

import (
	"fmt"
	"log"
	"time"

	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/openfaas/faas/gateway/queue"
)

// Sarama currently cannot support latest kafka protocol version 0_11_x
var (
	SARAMA_KAFKA_PROTO_VER = sarama.V0_10_2_0
)

// KafkaQueue queue for work
type KafkaQueue struct {
	topics []string
	shutdown chan struct{}
	producer sarama.AsyncProducer
}

// CreateKafkaQueue ready for asynchronous processing
func CreateKafkaQueue(kafkaBrokers []string, topics []string) (*KafkaQueue, error) {
	kque := &KafkaQueue{topics:topics,shutdown:make(chan struct{})}
	var err error
	log.Printf("Opening connection to %v\n", kafkaBrokers)

	// OpenFaas gateway and zookeeper & kafka brokers may start at same time
	// wait for that kafka is up and topics are provisioned
	waitforBrokersTopics(kafkaBrokers,topics)

	//setup async producer; should we use sync producer (lower throughput)?
	pConfig := sarama.NewConfig()
	pConfig.Version = SARAMA_KAFKA_PROTO_VER
	pConfig.Producer.Return.Successes = true
	kque.producer, err = sarama.NewAsyncProducer(kafkaBrokers, pConfig)

	if err != nil {
		log.Fatalf("Fail to create KafkaQueue %s\n",err)
	} else {
		fmt.Printf("Created KafkaQueue Producer %v\n", kque.producer)
		// Spawn goroutine to check KafkaQueue delivery status
		go func() {
			// Close producer will close "Successes","Errors" chans
			numClosedChans := 0
			for numClosedChans < 2 {
				select {
				case _ = <-kque.shutdown:
					fmt.Println("KafkaQueue shutdown...")
					// Trigger drain & close the following 2 chan
					kque.producer.AsyncClose()
				case msg,ok := <-kque.producer.Successes():
					if ok {
						fmt.Println("KafkaQueue: succeed sending 1 request: ",msg)
					} else {
						numClosedChans++
					}
				case err, ok := <-kque.producer.Errors():
					if ok {
						fmt.Println("KafkaQueue: producer error: ",err)
					} else {
						numClosedChans++
					}
				}
			}
			kque.shutdown<-struct{}{}
		}()
	}

	return kque, err
}

// Remember call this when gateway shuts down to claim background goroutine
func (kque *KafkaQueue) Shutdown() {
	kque.shutdown<-struct{}{}
	<-kque.shutdown
}

// Queue request for processing
func (kque *KafkaQueue) Queue(req *queue.Request) error {
	var err error

	out, err := json.Marshal(req)
	if err != nil {
		log.Println(err)
	}

	// We only have 1 topic now: faas-request
	kque.producer.Input() <- &sarama.ProducerMessage{Topic: kque.topics[0], Key: nil, Value: sarama.ByteEncoder(out)}

	fmt.Printf("KafkaQueue - submitting request: %s.\n", req.Function)

	return nil
}

// OpenFaas gateway and zookeeper & kafka brokers may start at same time
// wait for that kafka is up and topics are provisioned
func waitforBrokersTopics(brokers []string, topics []string) {
	var client sarama.Client
	var err error
	for {
		client,err = sarama.NewClient(brokers,nil)
		if client!=nil && err==nil { break }
		if client!=nil { client.Close() }
		fmt.Println("Wait for kafka brokers coming up...")
		time.Sleep(2*time.Second)
	}
	fmt.Println("Kafka brokers up")
	count := len(topics)
LOOP_TOPIC:
	for {
		tops,err := client.Topics()
		if tops!=nil && err==nil {
			for _,t1 := range tops {
				for _,t2 := range topics {
					if t1==t2 {
						fmt.Println("Topic ",t2," is ready")
						count--
						if count==0 { // All expected topics ready
							break LOOP_TOPIC
						} else {
							break
						}
					}
				}
			}
		}
		fmt.Println("Wait for topics:",topics)
		client.RefreshMetadata(topics...)
		time.Sleep(2*time.Second)
	}
	client.Close()
}
