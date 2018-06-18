// Copyright 2016-2018 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package stan

////////////////////////////////////////////////////////////////////////////////
// Benchmarks
////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkPublish(b *testing.B) {
	b.StopTimer()

	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()
	sc := NewDefaultConnection(b)
	defer sc.Close()

	hw := []byte("Hello World")

	b.StartTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := sc.Publish("foo", hw); err != nil {
			b.Fatalf("Got error on publish: %v\n", err)
		}
	}
}

func BenchmarkPublishAsync(b *testing.B) {
	b.StopTimer()

	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()
	sc := NewDefaultConnection(b)
	defer sc.Close()

	hw := []byte("Hello World")

	ch := make(chan bool)
	received := int32(0)

	ah := func(guid string, err error) {
		if err != nil {
			b.Fatalf("Received an error in ack callback: %v\n", err)
		}
		if nr := atomic.AddInt32(&received, 1); nr >= int32(b.N) {
			ch <- true
		}
	}
	b.StartTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := sc.PublishAsync("foo", hw, ah); err != nil {
			//fmt.Printf("Client status %v, Server status %v\n", s.nc.Status(), (sc.(*conn)).nc.Status())
			fmt.Printf("len(ackmap) = %d\n", len(sc.(*conn).pubAckMap))

			b.Fatalf("Error from PublishAsync: %v\n", err)
		}
	}

	err := WaitTime(ch, 10*time.Second)
	if err != nil {
		fmt.Printf("sc error is %v\n", sc.(*conn).nc.LastError())
		b.Fatal("Timed out waiting for ack messages")
	} else if atomic.LoadInt32(&received) != int32(b.N) {
		b.Fatalf("Received: %d", received)
	}

	//	msgs, bytes, _ := sc.(*conn).ackSubscription.MaxPending()
	//	fmt.Printf("max pending msgs:%d bytes:%d\n", msgs, bytes)
}

func BenchmarkSubscribe(b *testing.B) {
	b.StopTimer()

	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()
	sc := NewDefaultConnection(b)
	defer sc.Close()

	hw := []byte("Hello World")
	pch := make(chan bool)

	// Queue up all the messages. Keep this outside of the timing.
	for i := 0; i < b.N; i++ {
		if i == b.N-1 {
			// last one
			sc.PublishAsync("foo", hw, func(lguid string, err error) {
				if err != nil {
					b.Fatalf("Got an error from ack handler, %v", err)
				}
				pch <- true
			})
		} else {
			sc.PublishAsync("foo", hw, nil)
		}
	}

	// Wait for published to finish
	if err := WaitTime(pch, 10*time.Second); err != nil {
		b.Fatalf("Error waiting for publish to finish\n")
	}

	ch := make(chan bool)
	received := int32(0)

	b.StartTimer()
	b.ReportAllocs()

	sc.Subscribe("foo", func(m *Msg) {
		if nr := atomic.AddInt32(&received, 1); nr >= int32(b.N) {
			ch <- true
		}
	}, DeliverAllAvailable())

	err := WaitTime(ch, 10*time.Second)
	nr := atomic.LoadInt32(&received)
	if err != nil {
		b.Fatalf("Timed out waiting for messages, received only %d of %d\n", nr, b.N)
	} else if nr != int32(b.N) {
		b.Fatalf("Only Received: %d of %d", received, b.N)
	}
}

func BenchmarkQueueSubscribe(b *testing.B) {
	b.StopTimer()

	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()
	sc := NewDefaultConnection(b)
	defer sc.Close()

	hw := []byte("Hello World")
	pch := make(chan bool)

	// Queue up all the messages. Keep this outside of the timing.
	for i := 0; i < b.N; i++ {
		if i == b.N-1 {
			// last one
			sc.PublishAsync("foo", hw, func(lguid string, err error) {
				if err != nil {
					b.Fatalf("Got an error from ack handler, %v", err)
				}
				pch <- true
			})
		} else {
			sc.PublishAsync("foo", hw, nil)
		}
	}

	// Wait for published to finish
	if err := WaitTime(pch, 10*time.Second); err != nil {
		b.Fatalf("Error waiting for publish to finish\n")
	}

	ch := make(chan bool)
	received := int32(0)

	b.StartTimer()
	b.ReportAllocs()

	mcb := func(m *Msg) {
		if nr := atomic.AddInt32(&received, 1); nr >= int32(b.N) {
			ch <- true
		}
	}

	sc.QueueSubscribe("foo", "bar", mcb, DeliverAllAvailable())
	sc.QueueSubscribe("foo", "bar", mcb, DeliverAllAvailable())
	sc.QueueSubscribe("foo", "bar", mcb, DeliverAllAvailable())
	sc.QueueSubscribe("foo", "bar", mcb, DeliverAllAvailable())

	err := WaitTime(ch, 20*time.Second)
	nr := atomic.LoadInt32(&received)
	if err != nil {
		b.Fatalf("Timed out waiting for messages, received only %d of %d\n", nr, b.N)
	} else if nr != int32(b.N) {
		b.Fatalf("Only Received: %d of %d", received, b.N)
	}
}

func BenchmarkPublishSubscribe(b *testing.B) {
	b.StopTimer()

	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()
	sc := NewDefaultConnection(b)
	defer sc.Close()

	hw := []byte("Hello World")

	ch := make(chan bool)
	received := int32(0)

	// Subscribe callback, counts msgs received.
	_, err := sc.Subscribe("foo", func(m *Msg) {
		if nr := atomic.AddInt32(&received, 1); nr >= int32(b.N) {
			ch <- true
		}
	}, DeliverAllAvailable())

	if err != nil {
		b.Fatalf("Error subscribing, %v", err)
	}

	b.StartTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := sc.PublishAsync("foo", hw, func(guid string, err error) {
			if err != nil {
				b.Fatalf("Received an error in publish ack callback: %v\n", err)
			}
		})
		if err != nil {
			b.Fatalf("Error publishing %v\n", err)
		}
	}

	err = WaitTime(ch, 30*time.Second)
	nr := atomic.LoadInt32(&received)
	if err != nil {
		b.Fatalf("Timed out waiting for messages, received only %d of %d\n", nr, b.N)
	} else if nr != int32(b.N) {
		b.Fatalf("Only Received: %d of %d", received, b.N)
	}
}

func BenchmarkTimeNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		now := time.Now()
		now.Add(10 * time.Nanosecond)
	}
}
