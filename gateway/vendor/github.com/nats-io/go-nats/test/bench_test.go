// Copyright 2012-2018 The NATS Authors
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

package test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/nats-io/go-nats"
)

func BenchmarkPublishSpeed(b *testing.B) {
	b.StopTimer()
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(b)
	defer nc.Close()
	b.StartTimer()

	msg := []byte("Hello World")

	for i := 0; i < b.N; i++ {
		if err := nc.Publish("foo", msg); err != nil {
			b.Fatalf("Error in benchmark during Publish: %v\n", err)
		}
	}
	// Make sure they are all processed.
	nc.Flush()
	b.StopTimer()
}

func BenchmarkPubSubSpeed(b *testing.B) {
	b.StopTimer()
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(b)
	defer nc.Close()

	ch := make(chan bool)

	nc.SetErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) {
		b.Fatalf("Error : %v\n", err)
	})

	received := int32(0)

	nc.Subscribe("foo", func(m *nats.Msg) {
		if nr := atomic.AddInt32(&received, 1); nr >= int32(b.N) {
			ch <- true
		}
	})

	msg := []byte("Hello World")

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		if err := nc.Publish("foo", msg); err != nil {
			b.Fatalf("Error in benchmark during Publish: %v\n", err)
		}
		// Don't overrun ourselves and be a slow consumer, server will cut us off
		if int32(i)-atomic.LoadInt32(&received) > 32768 {
			time.Sleep(100 * time.Nanosecond)
		}
	}

	// Make sure they are all processed.
	err := WaitTime(ch, 10*time.Second)
	if err != nil {
		b.Fatal("Timed out waiting for messages")
	} else if atomic.LoadInt32(&received) != int32(b.N) {
		b.Fatalf("Received: %d, err:%v", received, nc.LastError())
	}
	b.StopTimer()
}

func BenchmarkAsyncSubscriptionCreationSpeed(b *testing.B) {
	b.StopTimer()
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(b)
	defer nc.Close()
	b.StartTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		nc.Subscribe("foo", func(m *nats.Msg) {})
	}
}

func BenchmarkSyncSubscriptionCreationSpeed(b *testing.B) {
	b.StopTimer()
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(b)
	defer nc.Close()
	b.StartTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		nc.SubscribeSync("foo")
	}
}

func BenchmarkInboxCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		nats.NewInbox()
	}
}

func BenchmarkRequest(b *testing.B) {
	b.StopTimer()
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(b)
	defer nc.Close()
	ok := []byte("ok")
	nc.Subscribe("req", func(m *nats.Msg) {
		nc.Publish(m.Reply, ok)
	})
	b.StartTimer()
	b.ReportAllocs()
	q := []byte("q")
	for i := 0; i < b.N; i++ {
		_, err := nc.Request("req", q, 1*time.Second)
		if err != nil {
			b.Fatalf("Err %v\n", err)
		}
	}
}

func BenchmarkOldRequest(b *testing.B) {
	b.StopTimer()
	s := RunDefaultServer()
	defer s.Shutdown()
	nc, err := nats.Connect(nats.DefaultURL, nats.UseOldRequestStyle())
	if err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer nc.Close()
	ok := []byte("ok")
	nc.Subscribe("req", func(m *nats.Msg) {
		nc.Publish(m.Reply, ok)
	})
	b.StartTimer()
	b.ReportAllocs()
	q := []byte("q")
	for i := 0; i < b.N; i++ {
		_, err := nc.Request("req", q, 1*time.Second)
		if err != nil {
			b.Fatalf("Err %v\n", err)
		}
	}
}
