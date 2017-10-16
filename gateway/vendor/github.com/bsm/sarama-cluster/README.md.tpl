# Sarama Cluster

[![GoDoc](https://godoc.org/github.com/bsm/sarama-cluster?status.svg)](https://godoc.org/github.com/bsm/sarama-cluster)
[![Build Status](https://travis-ci.org/bsm/sarama-cluster.svg?branch=master)](https://travis-ci.org/bsm/sarama-cluster)
[![Go Report Card](https://goreportcard.com/badge/github.com/bsm/sarama-cluster)](https://goreportcard.com/report/github.com/bsm/sarama-cluster)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Cluster extensions for [Sarama](https://github.com/Shopify/sarama), the Go client library for Apache Kafka 0.9 (and later).

## Documentation

Documentation and example are available via godoc at http://godoc.org/github.com/bsm/sarama-cluster

## Examples

Consumers have two modes of operation. In the default multiplexed mode messages (and errors) of multiple
topics and partitions are all passed to the single channel:

```go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	cluster "github.com/bsm/sarama-cluster"
)

func main() {{ "ExampleConsumer" | code }}
```

Users who require access to individual partitions can use the partitioned mode which exposes access to partition-level
consumers:

```go
package main

import (
  "fmt"
  "log"
  "os"
  "os/signal"

  cluster "github.com/bsm/sarama-cluster"
)

func main() {{ "ExampleConsumer_Partitions" | code }}
```

## Running tests

You need to install Ginkgo & Gomega to run tests. Please see
http://onsi.github.io/ginkgo for more details.

To run tests, call:

	$ make test

## Troubleshooting

### Consumer not receiving any messages?

By default, sarama's `Config.Consumer.Offsets.Initial` is set to `sarama.OffsetNewest`. This means that in the event that a brand new consumer is created, and it has never committed any offsets to kafka, it will only receive messages starting from the message after the current one that was written.

If you wish to receive all messages (from the start of all messages in the topic) in the event that a consumer does not have any offsets committed to kafka, you need to set `Config.Consumer.Offsets.Initial` to `sarama.OffsetOldest`.
