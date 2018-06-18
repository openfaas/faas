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
// Package scoped specific tests here..
////////////////////////////////////////////////////////////////////////////////

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	natsd "github.com/nats-io/gnatsd/test"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming/pb"
	"github.com/nats-io/nats-streaming-server/server"
)

func RunServer(ID string) *server.StanServer {
	s, err := server.RunServer(ID)
	if err != nil {
		panic(err)
	}
	return s
}

func runServerWithOpts(sOpts *server.Options) *server.StanServer {
	s, err := server.RunServerWithOpts(sOpts, nil)
	if err != nil {
		panic(err)
	}
	return s
}

// Dumb wait program to sync on callbacks, etc... Will timeout
func Wait(ch chan bool) error {
	return WaitTime(ch, 5*time.Second)
}

func WaitTime(ch chan bool, timeout time.Duration) error {
	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
	}
	return errors.New("timeout")
}

func TestNoNats(t *testing.T) {
	if _, err := Connect("someNonExistentServerID", "myTestClient"); err != nats.ErrNoServers {
		t.Fatalf("Expected NATS: No Servers err, got %v\n", err)
	}
}

func TestUnreachable(t *testing.T) {
	s := natsd.RunDefaultServer()
	defer s.Shutdown()

	// Non-Existent or Unreachable
	connectTime := 25 * time.Millisecond
	start := time.Now()
	if _, err := Connect("someNonExistentServerID", "myTestClient", ConnectWait(connectTime)); err != ErrConnectReqTimeout {
		t.Fatalf("Expected Unreachable err, got %v\n", err)
	}
	if delta := time.Since(start); delta < connectTime {
		t.Fatalf("Expected to wait at least %v, but only waited %v\n", connectTime, delta)
	}
}

const (
	clusterName = "my_test_cluster"
	clientName  = "me"
)

// So that we can pass tests and benchmarks...
type tLogger interface {
	Fatalf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

func stackFatalf(t tLogger, f string, args ...interface{}) {
	lines := make([]string, 0, 32)
	msg := fmt.Sprintf(f, args...)
	lines = append(lines, msg)

	// Generate the Stack of callers:
	for i := 1; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		msg := fmt.Sprintf("%d - %s:%d", i, file, line)
		lines = append(lines, msg)
	}

	t.Fatalf("%s", strings.Join(lines, "\n"))
}

func NewDefaultConnection(t tLogger) Conn {
	sc, err := Connect(clusterName, clientName)
	if err != nil {
		stackFatalf(t, "Expected to connect correctly, got err %v", err)
	}
	return sc
}

func TestConnClosedOnConnectFailure(t *testing.T) {
	s := natsd.RunDefaultServer()
	defer s.Shutdown()

	// Non-Existent or Unreachable
	connectTime := 25 * time.Millisecond
	if _, err := Connect("someNonExistentServerID", "myTestClient", ConnectWait(connectTime)); err != ErrConnectReqTimeout {
		t.Fatalf("Expected Unreachable err, got %v\n", err)
	}

	// Check that the underlying NATS connection has been closed.
	// We will first stop the server. If we have left the NATS connection
	// opened, it should be trying to reconnect.
	s.Shutdown()

	// Wait a bit
	time.Sleep(500 * time.Millisecond)

	// Inspecting go routines in search for a doReconnect
	buf := make([]byte, 10000)
	n := runtime.Stack(buf, true)
	if strings.Contains(string(buf[:n]), "doReconnect") {
		t.Fatal("NATS Connection suspected to not have been closed")
	}
}

func TestNatsConnNotClosedOnClose(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	// Create a NATS connection
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("Unexpected error on Connect: %v", err)
	}
	defer nc.Close()

	// Pass this NATS connection to NATS Streaming
	sc, err := Connect(clusterName, clientName, NatsConn(nc))
	if err != nil {
		t.Fatalf("Unexpected error on connect: %v", err)
	}
	// Now close the NATS Streaming connection
	sc.Close()

	// Verify that NATS connection is not closed
	if nc.IsClosed() {
		t.Fatal("NATS connection should NOT have been closed in Connect")
	}
}

func TestBasicConnect(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()
	sc := NewDefaultConnection(t)
	defer sc.Close()
}

func TestBasicPublish(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()
	sc := NewDefaultConnection(t)
	defer sc.Close()
	if err := sc.Publish("foo", []byte("Hello World!")); err != nil {
		t.Fatalf("Expected no errors on publish, got %v\n", err)
	}
}

func TestBasicPublishAsync(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()
	sc := NewDefaultConnection(t)
	defer sc.Close()
	ch := make(chan bool)
	var glock sync.Mutex
	var guid string
	acb := func(lguid string, err error) {
		glock.Lock()
		defer glock.Unlock()
		if lguid != guid {
			t.Fatalf("Expected a matching guid in ack callback, got %s vs %s\n", lguid, guid)
		}
		ch <- true
	}
	glock.Lock()
	guid, _ = sc.PublishAsync("foo", []byte("Hello World!"), acb)
	glock.Unlock()
	if guid == "" {
		t.Fatalf("Expected non-empty guid to be returned.")
	}
	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive our ack callback")
	}
}

func TestTimeoutPublish(t *testing.T) {
	ns := natsd.RunDefaultServer()
	defer ns.Shutdown()

	opts := server.GetDefaultOptions()
	opts.NATSServerURL = nats.DefaultURL
	opts.ID = clusterName
	s := runServerWithOpts(opts)
	defer s.Shutdown()
	sc, err := Connect(clusterName, clientName, PubAckWait(50*time.Millisecond))

	if err != nil {
		t.Fatalf("Expected to connect correctly, got err %v\n", err)
	}
	// Do not defer the connection close because we are going to
	// shutdown the server before the client connection is closed,
	// which would cause a 2 seconds delay on test exit.

	ch := make(chan bool)
	var glock sync.Mutex
	var guid string
	acb := func(lguid string, err error) {
		glock.Lock()
		defer glock.Unlock()
		if lguid != guid {
			t.Fatalf("Expected a matching guid in ack callback, got %s vs %s\n", lguid, guid)
		}
		if err != ErrTimeout {
			t.Fatalf("Expected a timeout error, got %v", err)
		}
		ch <- true
	}
	// Kill the NATS Streaming server so we timeout.
	s.Shutdown()
	glock.Lock()
	guid, _ = sc.PublishAsync("foo", []byte("Hello World!"), acb)
	glock.Unlock()
	if guid == "" {
		t.Fatalf("Expected non-empty guid to be returned.")
	}
	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive our ack callback with a timeout err")
	}
	// Publish synchronously
	if err := sc.Publish("foo", []byte("hello")); err == nil || err != ErrTimeout {
		t.Fatalf("Expected Timeout error on publish, got %v", err)
	}
}

func TestPublishWithClosedNATSConn(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("Unexpected error on connect: %v", err)
	}
	defer nc.Close()
	sc, err := Connect(clusterName, clientName, NatsConn(nc))
	if err != nil {
		t.Fatalf("Unexpected error on connect: %v", err)
	}
	defer sc.Close()
	// Close the NATS Connection
	nc.Close()
	msg := []byte("hello")
	// Publish should fail
	if err := sc.Publish("foo", msg); err == nil {
		t.Fatal("Expected error on publish")
	}
	// Even PublishAsync should fail right away
	if _, err := sc.PublishAsync("foo", msg, nil); err == nil {
		t.Fatal("Expected error on publish")
	}
}

func TestBasicSubscription(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	sub, err := sc.Subscribe("foo", func(m *Msg) {})
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer sub.Unsubscribe()

	// Close connection
	sc.Close()

	// Expect ErrConnectionClosed on subscribe
	if _, err := sc.Subscribe("foo", func(m *Msg) {}); err == nil || err != ErrConnectionClosed {
		t.Fatalf("Expected ErrConnectionClosed on subscribe, got %v", err)
	}
	if _, err := sc.QueueSubscribe("foo", "bar", func(m *Msg) {}); err == nil || err != ErrConnectionClosed {
		t.Fatalf("Expected ErrConnectionClosed on subscribe, got %v", err)
	}
}

func TestBasicQueueSubscription(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	ch := make(chan bool)
	count := uint32(0)
	cb := func(m *Msg) {
		if m.Sequence == 1 {
			if atomic.AddUint32(&count, 1) == 2 {
				ch <- true
			}
		}
	}

	sub, err := sc.QueueSubscribe("foo", "bar", cb)
	if err != nil {
		t.Fatalf("Expected no error on Subscribe, got %v\n", err)
	}
	defer sub.Unsubscribe()

	// Test that durable and non durable queue subscribers with
	// same name can coexist and they both receive the same message.
	if _, err = sc.QueueSubscribe("foo", "bar", cb, DurableName("durable-queue-sub")); err != nil {
		t.Fatalf("Unexpected error on QueueSubscribe with DurableName: %v", err)
	}

	// Publish a message
	if err := sc.Publish("foo", []byte("msg")); err != nil {
		t.Fatalf("Unexpected error on publish: %v", err)
	}
	// Wait for both messages to be received.
	if err := Wait(ch); err != nil {
		t.Fatal("Did not get our message")
	}

	// Check that one cannot use ':' for the queue durable name.
	if _, err := sc.QueueSubscribe("foo", "bar", cb, DurableName("my:dur")); err == nil {
		t.Fatal("Expected to get an error regarding durable name")
	}
}

func TestDurableQueueSubscriber(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	total := 5
	for i := 0; i < total; i++ {
		if err := sc.Publish("foo", []byte("msg")); err != nil {
			t.Fatalf("Unexpected error on publish: %v", err)
		}
	}
	ch := make(chan bool)
	firstBatch := uint64(total)
	secondBatch := uint64(2 * total)
	cb := func(m *Msg) {
		if !m.Redelivered &&
			(m.Sequence == uint64(firstBatch) || m.Sequence == uint64(secondBatch)) {
			ch <- true
		}
	}
	if _, err := sc.QueueSubscribe("foo", "bar", cb,
		DeliverAllAvailable(),
		DurableName("durable-queue-sub")); err != nil {
		t.Fatalf("Unexpected error on QueueSubscribe with DurableName: %v", err)
	}
	if err := Wait(ch); err != nil {
		t.Fatal("Did not get our message")
	}
	// Close connection
	sc.Close()

	// Create new connection
	sc = NewDefaultConnection(t)
	defer sc.Close()
	// Send more messages
	for i := 0; i < total; i++ {
		if err := sc.Publish("foo", []byte("msg")); err != nil {
			t.Fatalf("Unexpected error on publish: %v", err)
		}
	}
	// Create durable queue sub, it should receive from where it left of,
	// and ignore the start position
	if _, err := sc.QueueSubscribe("foo", "bar", cb,
		StartAtSequence(uint64(10*total)),
		DurableName("durable-queue-sub")); err != nil {
		t.Fatalf("Unexpected error on QueueSubscribe with DurableName: %v", err)
	}
	if err := Wait(ch); err != nil {
		t.Fatal("Did not get our message")
	}
}

func TestBasicPubSub(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	ch := make(chan bool)
	received := int32(0)
	toSend := int32(500)
	hw := []byte("Hello World")
	msgMap := make(map[uint64]struct{})

	sub, err := sc.Subscribe("foo", func(m *Msg) {
		if m.Subject != "foo" {
			t.Fatalf("Expected subject of 'foo', got '%s'\n", m.Subject)
		}
		if !bytes.Equal(m.Data, hw) {
			t.Fatalf("Wrong payload, got %q\n", m.Data)
		}
		// Make sure Seq and Timestamp are set
		if m.Sequence == 0 {
			t.Fatalf("Expected Sequence to be set\n")
		}
		if m.Timestamp == 0 {
			t.Fatalf("Expected timestamp to be set\n")
		}

		if _, ok := msgMap[m.Sequence]; ok {
			t.Fatalf("Detected duplicate for sequence: %d\n", m.Sequence)
		}
		msgMap[m.Sequence] = struct{}{}

		if nr := atomic.AddInt32(&received, 1); nr >= int32(toSend) {
			ch <- true
		}
	})
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer sub.Unsubscribe()

	for i := int32(0); i < toSend; i++ {
		if err := sc.Publish("foo", hw); err != nil {
			t.Fatalf("Received error on publish: %v\n", err)
		}
	}
	if err := WaitTime(ch, 1*time.Second); err != nil {
		t.Fatal("Did not receive our messages")
	}
}

func TestBasicPubQueueSub(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	ch := make(chan bool)
	received := int32(0)
	toSend := int32(100)
	hw := []byte("Hello World")

	sub, err := sc.QueueSubscribe("foo", "bar", func(m *Msg) {
		if m.Subject != "foo" {
			t.Fatalf("Expected subject of 'foo', got '%s'\n", m.Subject)
		}
		if !bytes.Equal(m.Data, hw) {
			t.Fatalf("Wrong payload, got %q\n", m.Data)
		}
		// Make sure Seq and Timestamp are set
		if m.Sequence == 0 {
			t.Fatalf("Expected Sequence to be set\n")
		}
		if m.Timestamp == 0 {
			t.Fatalf("Expected timestamp to be set\n")
		}
		if nr := atomic.AddInt32(&received, 1); nr >= int32(toSend) {
			ch <- true
		}
	})
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer sub.Unsubscribe()

	for i := int32(0); i < toSend; i++ {
		sc.Publish("foo", hw)
	}
	if err := WaitTime(ch, 1*time.Second); err != nil {
		t.Fatal("Did not receive our messages")
	}
}

func TestSubscriptionStartPositionLast(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	// Publish ten messages
	for i := 0; i < 10; i++ {
		data := []byte(fmt.Sprintf("%d", i))
		sc.Publish("foo", data)
	}

	ch := make(chan bool)
	received := int32(0)

	mcb := func(m *Msg) {
		atomic.AddInt32(&received, 1)
		if m.Sequence != 10 {
			t.Fatalf("Wrong sequence received: got %d vs. %d\n", m.Sequence, 10)
		}
		ch <- true
	}
	// Now subscribe and set start position to last received.
	sub, err := sc.Subscribe("foo", mcb, StartWithLastReceived())
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer sub.Unsubscribe()

	// Check for sub setup
	rsub := sub.(*subscription)
	if rsub.opts.StartAt != pb.StartPosition_LastReceived {
		t.Fatalf("Incorrect StartAt state: %s\n", rsub.opts.StartAt)
	}

	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive our message")
	}

	if received > int32(1) {
		t.Fatalf("Should have received only 1 message, but got %d\n", received)
	}
}

func TestSubscriptionStartAtSequence(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	// Publish ten messages
	for i := 1; i <= 10; i++ {
		data := []byte(fmt.Sprintf("%d", i))
		sc.Publish("foo", data)
	}

	ch := make(chan bool)
	received := int32(0)
	shouldReceive := int32(5)

	// Capture the messages that are delivered.
	savedMsgs := make([]*Msg, 0, 5)

	mcb := func(m *Msg) {
		savedMsgs = append(savedMsgs, m)
		if nr := atomic.AddInt32(&received, 1); nr >= int32(shouldReceive) {
			ch <- true
		}
	}

	// Now subscribe and set start position to #6, so should received 6-10.
	sub, err := sc.Subscribe("foo", mcb, StartAtSequence(6))
	if err != nil {
		t.Fatalf("Expected no error on Subscribe, got %v\n", err)
	}
	defer sub.Unsubscribe()

	// Check for sub setup
	rsub := sub.(*subscription)
	if rsub.opts.StartAt != pb.StartPosition_SequenceStart {
		t.Fatalf("Incorrect StartAt state: %s\n", rsub.opts.StartAt)
	}

	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive our messages")
	}

	// Check we received them in order.
	for i, seq := 0, uint64(6); i < 5; i++ {
		m := savedMsgs[i]
		// Check Sequence
		if m.Sequence != seq {
			t.Fatalf("Expected seq: %d, got %d\n", seq, m.Sequence)
		}
		// Check payload
		dseq, _ := strconv.Atoi(string(m.Data))
		if dseq != int(seq) {
			t.Fatalf("Expected payload: %d, got %d\n", seq, dseq)
		}
		seq++
	}
}

func TestSubscriptionStartAtTime(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	// Publish first 5
	for i := 1; i <= 5; i++ {
		data := []byte(fmt.Sprintf("%d", i))
		sc.Publish("foo", data)
	}

	// Buffer each side so slow tests still work.
	time.Sleep(250 * time.Millisecond)
	startTime := time.Now()
	time.Sleep(250 * time.Millisecond)

	// Publish last 5
	for i := 6; i <= 10; i++ {
		data := []byte(fmt.Sprintf("%d", i))
		sc.Publish("foo", data)
	}

	ch := make(chan bool)
	received := int32(0)
	shouldReceive := int32(5)

	// Capture the messages that are delivered.
	savedMsgs := make([]*Msg, 0, 5)

	mcb := func(m *Msg) {
		savedMsgs = append(savedMsgs, m)
		if nr := atomic.AddInt32(&received, 1); nr >= int32(shouldReceive) {
			ch <- true
		}
	}

	// Now subscribe and set start position to #6, so should received 6-10.
	sub, err := sc.Subscribe("foo", mcb, StartAtTime(startTime))
	if err != nil {
		t.Fatalf("Expected no error on Subscribe, got %v\n", err)
	}
	defer sub.Unsubscribe()

	// Check for sub setup
	rsub := sub.(*subscription)
	if rsub.opts.StartAt != pb.StartPosition_TimeDeltaStart {
		t.Fatalf("Incorrect StartAt state: %s\n", rsub.opts.StartAt)
	}

	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive our messages")
	}

	// Check we received them in order.
	for i, seq := 0, uint64(6); i < 5; i++ {
		m := savedMsgs[i]
		// Check time is always greater than startTime
		if m.Timestamp < startTime.UnixNano() {
			t.Fatalf("Expected all messages to have timestamp > startTime.")
		}
		// Check Sequence
		if m.Sequence != seq {
			t.Fatalf("Expected seq: %d, got %d\n", seq, m.Sequence)
		}
		// Check payload
		dseq, _ := strconv.Atoi(string(m.Data))
		if dseq != int(seq) {
			t.Fatalf("Expected payload: %d, got %d\n", seq, dseq)
		}
		seq++
	}

	// Now test Ago helper
	delta := time.Since(startTime)

	sub, err = sc.Subscribe("foo", mcb, StartAtTimeDelta(delta))
	if err != nil {
		t.Fatalf("Expected no error on Subscribe, got %v\n", err)
	}
	defer sub.Unsubscribe()

	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive our messages")
	}
}

func TestSubscriptionStartAt(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	// Publish ten messages
	for i := 1; i <= 10; i++ {
		sc.Publish("foo", []byte("hello"))
	}

	ch := make(chan bool)
	received := 0
	mcb := func(m *Msg) {
		received++
		if received == 10 {
			ch <- true
		}
	}

	// Now subscribe and set start position to sequence. It should be
	// sequence 0
	sub, err := sc.Subscribe("foo", mcb, StartAt(pb.StartPosition_SequenceStart))
	if err != nil {
		t.Fatalf("Expected no error on Subscribe, got %v", err)
	}
	defer sub.Unsubscribe()

	// Check for sub setup
	rsub := sub.(*subscription)
	if rsub.opts.StartAt != pb.StartPosition_SequenceStart {
		t.Fatalf("Incorrect StartAt state: %s\n", rsub.opts.StartAt)
	}

	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive our messages")
	}
}

func TestSubscriptionStartAtFirst(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	// Publish ten messages
	for i := 1; i <= 10; i++ {
		data := []byte(fmt.Sprintf("%d", i))
		sc.Publish("foo", data)
	}

	ch := make(chan bool)
	received := int32(0)
	shouldReceive := int32(10)

	// Capture the messages that are delivered.
	savedMsgs := make([]*Msg, 0, 10)

	mcb := func(m *Msg) {
		savedMsgs = append(savedMsgs, m)
		if nr := atomic.AddInt32(&received, 1); nr >= int32(shouldReceive) {
			ch <- true
		}
	}

	// Now subscribe and set start position to #6, so should received 6-10.
	sub, err := sc.Subscribe("foo", mcb, DeliverAllAvailable())
	if err != nil {
		t.Fatalf("Expected no error on Subscribe, got %v\n", err)
	}
	defer sub.Unsubscribe()

	// Check for sub setup
	rsub := sub.(*subscription)
	if rsub.opts.StartAt != pb.StartPosition_First {
		t.Fatalf("Incorrect StartAt state: %s\n", rsub.opts.StartAt)
	}

	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive our messages")
	}

	if received != shouldReceive {
		t.Fatalf("Expected %d msgs but received %d\n", shouldReceive, received)
	}

	// Check we received them in order.
	for i, seq := 0, uint64(1); i < 10; i++ {
		m := savedMsgs[i]
		// Check Sequence
		if m.Sequence != seq {
			t.Fatalf("Expected seq: %d, got %d\n", seq, m.Sequence)
		}
		// Check payload
		dseq, _ := strconv.Atoi(string(m.Data))
		if dseq != int(seq) {
			t.Fatalf("Expected payload: %d, got %d\n", seq, dseq)
		}
		seq++
	}
}

func TestUnsubscribe(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	// Create a valid one
	sc.Subscribe("foo", nil)

	// Now subscribe, but we will unsubscribe before sending any messages.
	sub, err := sc.Subscribe("foo", func(m *Msg) {
		t.Fatalf("Did not expect to receive any messages\n")
	})
	if err != nil {
		t.Fatalf("Expected no error on Subscribe, got %v\n", err)
	}
	// Create another valid one
	sc.Subscribe("foo", nil)

	// Unsubscribe middle one.
	err = sub.Unsubscribe()
	if err != nil {
		t.Fatalf("Expected no errors from unsubscribe: got %v\n", err)
	}
	// Do it again, should not dump, but should get error.
	err = sub.Unsubscribe()
	if err == nil || err != ErrBadSubscription {
		t.Fatalf("Expected a bad subscription err, got %v\n", err)
	}

	// Publish ten messages
	for i := 1; i <= 10; i++ {
		data := []byte(fmt.Sprintf("%d", i))
		sc.Publish("foo", data)
	}

	sc.Close()
	sc = NewDefaultConnection(t)
	defer sc.Close()
	sub1, err := sc.Subscribe("foo", func(_ *Msg) {})
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	sub2, err := sc.Subscribe("foo", func(_ *Msg) {})
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	// Override clientID to get an error on Subscription.Close() and Unsubscribe()
	sc.(*conn).Lock()
	sc.(*conn).clientID = "foobar"
	sc.(*conn).Unlock()
	if err := sub1.Close(); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("Expected error about unknown clientID, got %v", err)
	}
	if err := sub2.Close(); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("Expected error about unknown clientID, got %v", err)
	}
}

func TestUnsubscribeWhileConnClosing(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc, err := Connect(clusterName, clientName, PubAckWait(50*time.Millisecond))
	if err != nil {
		t.Fatalf("Expected to connect correctly, got err %v\n", err)
	}
	defer sc.Close()

	sub, err := sc.Subscribe("foo", nil)
	if err != nil {
		t.Fatalf("Expected no error on Subscribe, got %v\n", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		sc.Close()
		wg.Done()
	}()

	// Unsubscribe
	sub.Unsubscribe()

	wg.Wait()
}

func TestDupClientID(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	if _, err := Connect(clusterName, clientName); err == nil {
		t.Fatal("Expected to get an error for duplicate clientID")
	}
}

func TestClose(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	sub, err := sc.Subscribe("foo", func(m *Msg) {
		t.Fatalf("Did not expect to receive any messages\n")
	})
	if err != nil {
		t.Fatalf("Expected no errors when subscribing, got %v\n", err)
	}

	err = sc.Close()
	if err != nil {
		t.Fatalf("Did not expect error on Close(), got %v\n", err)
	}

	if _, err := sc.PublishAsync("foo", []byte("Hello World!"), nil); err == nil || err != ErrConnectionClosed {
		t.Fatalf("Expected an ErrConnectionClosed on publish async to a closed connection, got %v", err)
	}

	if err := sc.Publish("foo", []byte("Hello World!")); err == nil || err != ErrConnectionClosed {
		t.Fatalf("Expected an ErrConnectionClosed error on publish to a closed connection, got %v", err)
	}

	if err := sub.Unsubscribe(); err == nil || err != ErrConnectionClosed {
		t.Fatalf("Expected an ErrConnectionClosed error on unsubscribe to a closed connection, got %v", err)
	}

	sc = NewDefaultConnection(t)
	// Override the clientID so that we get an error on close
	sc.(*conn).Lock()
	sc.(*conn).clientID = "foobar"
	sc.(*conn).Unlock()
	if err := sc.Close(); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("Expected error about unknown clientID, got %v", err)
	}
}

func TestDoubleClose(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	if err := sc.Close(); err != nil {
		t.Fatalf("Did not expect an error on first Close, got %v\n", err)
	}

	if err := sc.Close(); err != nil {
		t.Fatalf("Did not expect an error on second Close, got %v\n", err)
	}
}

func TestManualAck(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	toSend := int32(100)
	hw := []byte("Hello World")

	for i := int32(0); i < toSend; i++ {
		sc.PublishAsync("foo", hw, nil)
	}
	sc.Publish("foo", hw)

	fch := make(chan bool)

	// Test that we can't Ack if not in manual mode.
	sub, err := sc.Subscribe("foo", func(m *Msg) {
		if err := m.Ack(); err != ErrManualAck {
			t.Fatalf("Expected an error trying to ack an auto-ack subscription")
		}
		fch <- true
	}, DeliverAllAvailable())
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}

	if err := Wait(fch); err != nil {
		t.Fatal("Did not receive our first message")
	}
	sub.Unsubscribe()

	ch := make(chan bool)
	sch := make(chan bool)
	received := int32(0)

	msgs := make([]*Msg, 0, 101)

	// Test we only receive MaxInflight if we do not ack
	sub, err = sc.Subscribe("foo", func(m *Msg) {
		msgs = append(msgs, m)
		if nr := atomic.AddInt32(&received, 1); nr == int32(10) {
			ch <- true
		} else if nr > 10 {
			m.Ack()
			if nr >= toSend+1 { // sync Publish +1
				sch <- true
			}
		}
	}, DeliverAllAvailable(), MaxInflight(10), SetManualAckMode())
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer sub.Unsubscribe()

	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive at least 10 messages")
	}
	// Wait a bit longer for other messages which would be an error.
	time.Sleep(50 * time.Millisecond)

	if nr := atomic.LoadInt32(&received); nr != 10 {
		t.Fatalf("Only expected to get 10 messages to match MaxInflight without Acks, got %d\n", nr)
	}

	// Now make sure we get the rest of them. So ack the ones we have so far.
	for _, m := range msgs {
		if err := m.Ack(); err != nil {
			t.Fatalf("Unexpected error on Ack: %v\n", err)
		}
	}
	if err := Wait(sch); err != nil {
		t.Fatal("Did not receive all our messages")
	}

	if nr := atomic.LoadInt32(&received); nr != toSend+1 {
		t.Fatalf("Did not receive correct number of messages: %d vs %d\n", nr, toSend+1)
	}

	// Close connection
	sc.Close()
	if err := msgs[0].Ack(); err != ErrBadConnection {
		t.Fatalf("Expected ErrBadConnection, got %v", err)
	}

	// Close the subscription
	sub.Unsubscribe()
	if err := msgs[0].Ack(); err != ErrBadSubscription {
		t.Fatalf("Expected ErrBadSubscription, got %v", err)
	}

	// Test nil msg Ack
	var m *Msg
	if err := m.Ack(); err != ErrNilMsg {
		t.Fatalf("Expected ErrNilMsg, got %v", err)
	}
}

func TestRedelivery(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	toSend := int32(100)
	hw := []byte("Hello World")

	for i := int32(0); i < toSend; i++ {
		sc.Publish("foo", hw)
	}

	// Make sure we get an error on bad ackWait
	if _, err := sc.Subscribe("foo", nil, AckWait(20*time.Millisecond)); err == nil {
		t.Fatalf("Expected an error for back AckWait time under 1 second\n")
	}

	ch := make(chan bool)
	sch := make(chan bool)
	received := int32(0)

	ackRedeliverTime := 1 * time.Second

	sub, err := sc.Subscribe("foo", func(m *Msg) {
		if nr := atomic.AddInt32(&received, 1); nr == toSend {
			ch <- true
		} else if nr == 2*toSend {
			sch <- true
		}

	}, DeliverAllAvailable(), MaxInflight(int(toSend+1)), AckWait(ackRedeliverTime), SetManualAckMode())
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer sub.Unsubscribe()

	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive first delivery of all messages")
	}
	if nr := atomic.LoadInt32(&received); nr != toSend {
		t.Fatalf("Expected to get 100 messages, got %d\n", nr)
	}
	if err := Wait(sch); err != nil {
		t.Fatal("Did not receive second re-delivery of all messages")
	}
	if nr := atomic.LoadInt32(&received); nr != 2*toSend {
		t.Fatalf("Expected to get 200 messages, got %d\n", nr)
	}
}

func TestDurableSubscriber(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	toSend := int32(100)
	hw := []byte("Hello World")

	// Capture the messages that are delivered.
	var msgsGuard sync.Mutex
	savedMsgs := make([]*Msg, 0, toSend)

	for i := int32(0); i < toSend; i++ {
		sc.Publish("foo", hw)
	}

	ch := make(chan bool)
	received := int32(0)

	_, err := sc.Subscribe("foo", func(m *Msg) {
		if nr := atomic.AddInt32(&received, 1); nr == 10 {
			sc.Close()
			ch <- true
		} else {
			msgsGuard.Lock()
			savedMsgs = append(savedMsgs, m)
			msgsGuard.Unlock()
		}
	}, DeliverAllAvailable(), DurableName("durable-foo"))
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}

	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive first delivery of all messages")
	}
	// Reset in case we get more messages in the above callback
	ch = make(chan bool)

	if nr := atomic.LoadInt32(&received); nr != 10 {
		t.Fatalf("Expected to get only 10 messages, got %d\n", nr)
	}
	// This is auto-ack, so undo received for check.
	// Close will prevent ack from going out, so #10 will be redelivered
	atomic.AddInt32(&received, -1)

	// sc is closed here from above..

	// Recreate the connection
	sc, err = Connect(clusterName, clientName, PubAckWait(50*time.Millisecond))
	if err != nil {
		t.Fatalf("Expected to connect correctly, got err %v\n", err)
	}
	defer sc.Close()

	// Create the same durable subscription.
	_, err = sc.Subscribe("foo", func(m *Msg) {
		msgsGuard.Lock()
		savedMsgs = append(savedMsgs, m)
		msgsGuard.Unlock()
		if nr := atomic.AddInt32(&received, 1); nr == toSend {
			ch <- true
		}
	}, DeliverAllAvailable(), DurableName("durable-foo"))
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}

	// Check that durables cannot be subscribed to again by same client.
	_, err = sc.Subscribe("foo", nil, DurableName("durable-foo"))
	if err == nil || err.Error() != server.ErrDupDurable.Error() {
		t.Fatalf("Expected ErrDupSubscription error, got %v\n", err)
	}

	// Check that durables with same name, but subscribed to differen subject are ok.
	_, err = sc.Subscribe("bar", nil, DurableName("durable-foo"))
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}

	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive delivery of all messages")
	}

	if nr := atomic.LoadInt32(&received); nr != toSend {
		t.Fatalf("Expected to get %d messages, got %d\n", toSend, nr)
	}
	msgsGuard.Lock()
	numSaved := len(savedMsgs)
	msgsGuard.Unlock()
	if numSaved != int(toSend) {
		t.Fatalf("Expected len(savedMsgs) to be %d, got %d\n", toSend, numSaved)
	}
	// Check we received them in order
	msgsGuard.Lock()
	defer msgsGuard.Unlock()
	for i, m := range savedMsgs {
		seqExpected := uint64(i + 1)
		if m.Sequence != seqExpected {
			t.Fatalf("Got wrong seq, expected %d, got %d\n", seqExpected, m.Sequence)
		}
	}
}

func TestRedeliveredFlag(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	toSend := int32(100)
	hw := []byte("Hello World")

	for i := int32(0); i < toSend; i++ {
		if err := sc.Publish("foo", hw); err != nil {
			t.Fatalf("Error publishing message: %v\n", err)
		}
	}

	ch := make(chan bool)
	received := int32(0)

	msgsLock := &sync.Mutex{}
	msgs := make(map[uint64]*Msg)

	// Test we only receive MaxInflight if we do not ack
	sub, err := sc.Subscribe("foo", func(m *Msg) {
		// Remember the message.
		msgsLock.Lock()
		msgs[m.Sequence] = m
		msgsLock.Unlock()

		// Only Ack odd numbers
		if m.Sequence%2 != 0 {
			if err := m.Ack(); err != nil {
				t.Fatalf("Unexpected error on Ack: %v\n", err)
			}
		}

		if nr := atomic.AddInt32(&received, 1); nr == toSend {
			ch <- true
		}
	}, DeliverAllAvailable(), AckWait(1*time.Second), SetManualAckMode())
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer sub.Unsubscribe()

	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive at least 10 messages")
	}
	time.Sleep(1500 * time.Millisecond) // Wait for redelivery

	msgsLock.Lock()
	defer msgsLock.Unlock()

	for _, m := range msgs {
		// Expect all even msgs to have been redelivered.
		if m.Sequence%2 == 0 && !m.Redelivered {
			t.Fatalf("Expected a redelivered flag to be set on msg %d\n", m.Sequence)
		}
	}
}

// TestNoDuplicatesOnSubscriberStart tests that a subscriber does not
// receive duplicate when requesting a replay while messages are being
// published on its subject.
func TestNoDuplicatesOnSubscriberStart(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc, err := Connect(clusterName, clientName)
	if err != nil {
		t.Fatalf("Expected to connect correctly, got err %v\n", err)
	}

	defer sc.Close()

	batch := int32(100)
	ch := make(chan bool)
	pch := make(chan bool)
	received := int32(0)
	sent := int32(0)

	mcb := func(m *Msg) {
		// signal when we've reached the expected messages count
		if nr := atomic.AddInt32(&received, 1); nr == atomic.LoadInt32(&sent) {
			ch <- true
		}
	}

	publish := func() {
		// publish until the receiver starts, then one additional batch.
		// This primes NATS Streaming with messages, and gives us a point to stop
		// when the subscriber has started processing messages.
		for atomic.LoadInt32(&received) == 0 {
			for i := int32(0); i < batch; i++ {
				atomic.AddInt32(&sent, 1)
				sc.PublishAsync("foo", []byte("hello"), nil)
			}
			// signal that we've published a batch.
			pch <- true
		}
	}

	go publish()

	// wait until the publisher has published at least one batch
	Wait(pch)

	// start the subscriber
	sub, err := sc.Subscribe("foo", mcb, DeliverAllAvailable())
	if err != nil {
		t.Fatalf("Expected no error on Subscribe, got %v\n", err)
	}

	defer sub.Unsubscribe()

	// Wait for our expected count.
	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive our messages")
	}

	// Wait to see if the subscriber receives any duplicate messages.
	time.Sleep(250 * time.Millisecond)

	// Make sure we've receive the exact count of sent messages.
	if atomic.LoadInt32(&received) != atomic.LoadInt32(&sent) {
		t.Fatalf("Expected %d msgs but received %d\n", sent, received)
	}
}

func TestMaxChannels(t *testing.T) {
	// Set a small number of max channels
	opts := server.GetDefaultOptions()
	opts.ID = clusterName
	opts.MaxChannels = 10

	// Run a NATS Streaming server
	s := runServerWithOpts(opts)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	hw := []byte("Hello World")
	var subject string

	// These all should work fine
	for i := 0; i < opts.MaxChannels; i++ {
		subject = fmt.Sprintf("CHAN-%d", i)
		sc.PublishAsync(subject, hw, nil)
	}
	// This one should error
	if err := sc.Publish("CHAN_MAX", hw); err == nil {
		t.Fatalf("Expected an error signaling too many channels\n")
	}
}

func TestRaceOnClose(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	// Seems that this sleep makes it happen all the time.
	time.Sleep(1250 * time.Millisecond)
}

func TestRaceAckOnClose(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	toSend := 100

	// Send our messages
	for i := 0; i < toSend; i++ {
		if err := sc.Publish("foo", []byte("msg")); err != nil {
			t.Fatalf("Unexpected error on publish: %v", err)
		}
	}

	cb := func(m *Msg) {
		m.Ack()
	}
	if _, err := sc.Subscribe("foo", cb, SetManualAckMode(),
		DeliverAllAvailable()); err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	// Close while ack'ing may happen
	time.Sleep(10 * time.Millisecond)
	sc.Close()
}

func TestNatsConn(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()
	sc := NewDefaultConnection(t)
	defer sc.Close()

	// Make sure we can get the STAN-created Conn.
	nc := sc.NatsConn()

	if nc.Status() != nats.CONNECTED {
		t.Fatal("Should have status set to CONNECTED")
	}
	nc.Close()
	if nc.Status() != nats.CLOSED {
		t.Fatal("Should have status set to CLOSED")
	}

	sc.Close()
	if sc.NatsConn() != nil {
		t.Fatal("Wrapped conn should be nil after close")
	}

	// Bail if we have a custom connection but not connected
	cnc := nats.Conn{Opts: nats.GetDefaultOptions()}
	if _, err := Connect(clusterName, clientName, NatsConn(&cnc)); err != ErrBadConnection {
		t.Fatalf("Expected to get an invalid connection error, got %v", err)
	}

	// Allow custom conn only if already connected
	opts := nats.GetDefaultOptions()
	nc, err := opts.Connect()
	if err != nil {
		t.Fatalf("Expected to connect correctly, got err %v", err)
	}
	sc, err = Connect(clusterName, clientName, NatsConn(nc))
	if err != nil {
		t.Fatalf("Expected to connect correctly, got err %v", err)
	}
	nc.Close()
	if nc.Status() != nats.CLOSED {
		t.Fatal("Should have status set to CLOSED")
	}
	sc.Close()

	// Make sure we can get the Conn we provide.
	opts = nats.GetDefaultOptions()
	nc, err = opts.Connect()
	if err != nil {
		t.Fatalf("Expected to connect correctly, got err %v", err)
	}
	defer nc.Close()
	sc, err = Connect(clusterName, clientName, NatsConn(nc))
	if err != nil {
		t.Fatalf("Expected to connect correctly, got err %v", err)
	}
	defer sc.Close()
	if sc.NatsConn() != nc {
		t.Fatal("Unexpected wrapped conn")
	}
}

func TestMaxPubAcksInflight(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("Unexpected error on connect: %v", err)
	}
	defer nc.Close()

	sc, err := Connect(clusterName, clientName,
		MaxPubAcksInflight(1),
		PubAckWait(time.Second),
		NatsConn(nc))
	if err != nil {
		t.Fatalf("Expected to connect correctly, got err %v", err)
	}
	// Don't defer the close of connection since the server is stopped,
	// the close would delay the test.

	// Cause the ACK to not come by shutdown the server now
	s.Shutdown()

	msg := []byte("hello")

	// Send more than one message, if MaxPubAcksInflight() works, one
	// of the publish call should block for up to PubAckWait.
	start := time.Now()
	for i := 0; i < 2; i++ {
		if _, err := sc.PublishAsync("foo", msg, nil); err != nil {
			t.Fatalf("Unexpected error on publish: %v", err)
		}
	}
	end := time.Now()
	// So if the loop ended before the PubAckWait timeout, then it's a failure.
	if end.Sub(start) < time.Second {
		t.Fatal("Should have blocked after 1 message sent")
	}
}

func TestNatsURLOption(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc, err := Connect(clusterName, clientName, NatsURL("nats://localhost:5555"))
	if err == nil {
		sc.Close()
		t.Fatal("Expected connect to fail")
	}
}

func TestSubscriptionPending(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	nc := sc.NatsConn()

	total := 100
	msg := []byte("0123456789")

	inCb := make(chan bool)
	block := make(chan bool)
	cb := func(m *Msg) {
		inCb <- true
		<-block
	}

	sub, _ := sc.QueueSubscribe("foo", "bar", cb)
	defer sub.Unsubscribe()

	// Publish five messages
	for i := 0; i < total; i++ {
		sc.Publish("foo", msg)
	}
	nc.Flush()

	// Wait for our first message
	if err := Wait(inCb); err != nil {
		t.Fatal("No message received")
	}

	m, b, _ := sub.Pending()
	// FIXME(jack0) - nats streaming appends clientid, guid, and subject to messages so bytes pending is greater than message size
	mlen := len(msg) + 19
	totalSize := total * mlen

	if m != total && m != total-1 {
		t.Fatalf("Expected msgs of %d or %d, got %d\n", total, total-1, m)
	}
	if b != totalSize && b != totalSize-mlen {
		t.Fatalf("Expected bytes of %d or %d, got %d\n",
			totalSize, totalSize-mlen, b)
	}

	// Make sure max has been set. Since we block after the first message is
	// received, MaxPending should be >= total - 1 and <= total
	mm, bm, _ := sub.MaxPending()
	if mm < total-1 || mm > total {
		t.Fatalf("Expected max msgs (%d) to be between %d and %d\n",
			mm, total-1, total)
	}
	if bm < totalSize-mlen || bm > totalSize {
		t.Fatalf("Expected max bytes (%d) to be between %d and %d\n",
			bm, totalSize, totalSize-mlen)
	}
	// Check that clear works.
	sub.ClearMaxPending()
	mm, bm, _ = sub.MaxPending()
	if mm != 0 {
		t.Fatalf("Expected max msgs to be 0 vs %d after clearing\n", mm)
	}
	if bm != 0 {
		t.Fatalf("Expected max bytes to be 0 vs %d after clearing\n", bm)
	}

	close(block)
	sub.Unsubscribe()

	// These calls should fail once the subscription is closed.
	if _, _, err := sub.Pending(); err == nil {
		t.Fatal("Calling Pending() on closed subscription should fail")
	}
	if _, _, err := sub.MaxPending(); err == nil {
		t.Fatal("Calling MaxPending() on closed subscription should fail")
	}
	if err := sub.ClearMaxPending(); err == nil {
		t.Fatal("Calling ClearMaxPending() on closed subscription should fail")
	}
}

func TestTimeoutOnRequests(t *testing.T) {
	ns := natsd.RunDefaultServer()
	defer ns.Shutdown()

	opts := server.GetDefaultOptions()
	opts.ID = clusterName
	opts.NATSServerURL = nats.DefaultURL
	s := runServerWithOpts(opts)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	sub1, err := sc.Subscribe("foo", func(_ *Msg) {})
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	sub2, err := sc.Subscribe("foo", func(_ *Msg) {})
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}

	// For this test, change the reqTimeout to very low value
	sc.(*conn).Lock()
	sc.(*conn).opts.ConnectTimeout = 10 * time.Millisecond
	sc.(*conn).Unlock()

	// Shutdown server
	s.Shutdown()

	// Subscribe
	if _, err := sc.Subscribe("foo", func(_ *Msg) {}); err != ErrSubReqTimeout {
		t.Fatalf("Expected %v error, got %v", ErrSubReqTimeout, err)
	}

	// If connecting to an old server...
	if sc.(*conn).subCloseRequests == "" {
		// Trick the API into thinking that it can send,
		// and make sure the call times-out
		sc.(*conn).Lock()
		sc.(*conn).subCloseRequests = "sub.close.subject"
		sc.(*conn).Unlock()
	}
	// Subscription Close
	if err := sub1.Close(); err != ErrCloseReqTimeout {
		t.Fatalf("Expected %v error, got %v", ErrCloseReqTimeout, err)
	}
	// Unsubscribe
	if err := sub2.Unsubscribe(); err != ErrUnsubReqTimeout {
		t.Fatalf("Expected %v error, got %v", ErrUnsubReqTimeout, err)
	}
	// Connection Close
	if err := sc.Close(); err != ErrCloseReqTimeout {
		t.Fatalf("Expected %v error, got %v", ErrCloseReqTimeout, err)
	}
}

func TestSlowAsyncSubscriber(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	nc := sc.NatsConn()

	bch := make(chan bool)

	sub, _ := sc.Subscribe("foo", func(_ *Msg) {
		// block to back us up..
		<-bch
	})
	// Make sure these are the defaults
	pm, pb, _ := sub.PendingLimits()
	if pm != nats.DefaultSubPendingMsgsLimit {
		t.Fatalf("Pending limit for number of msgs incorrect, expected %d, got %d\n", nats.DefaultSubPendingMsgsLimit, pm)
	}
	if pb != nats.DefaultSubPendingBytesLimit {
		t.Fatalf("Pending limit for number of bytes incorrect, expected %d, got %d\n", nats.DefaultSubPendingBytesLimit, pb)
	}

	// Set new limits
	pml := 100
	pbl := 1024 * 1024

	sub.SetPendingLimits(pml, pbl)

	// Make sure the set is correct
	pm, pb, _ = sub.PendingLimits()
	if pm != pml {
		t.Fatalf("Pending limit for number of msgs incorrect, expected %d, got %d\n", pml, pm)
	}
	if pb != pbl {
		t.Fatalf("Pending limit for number of bytes incorrect, expected %d, got %d\n", pbl, pb)
	}

	for i := 0; i < (int(pml) + 100); i++ {
		sc.Publish("foo", []byte("Hello"))
	}

	timeout := 5 * time.Second
	start := time.Now()
	err := nc.FlushTimeout(timeout)
	elapsed := time.Since(start)
	if elapsed >= timeout {
		t.Fatalf("Flush did not return before timeout")
	}
	// We want flush to work, so expect no error for it.
	if err != nil {
		t.Fatalf("Expected no error from Flush()\n")
	}
	if nc.LastError() != nats.ErrSlowConsumer {
		t.Fatal("Expected LastError to indicate slow consumer")
	}
	// release the sub
	bch <- true
}

func TestSubscriberClose(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	// If old server, Close() is expected to fail.
	supported := sc.(*conn).subCloseRequests != ""

	checkClose := func(sub Subscription) {
		err := sub.Close()
		if supported && err != nil {
			t.Fatalf("Unexpected error on close: %v", err)
		} else if !supported && err != ErrNoServerSupport {
			t.Fatalf("Expected %v, got %v", ErrNoServerSupport, err)
		}
	}

	count := 1
	if supported {
		count = 2
	}
	for i := 0; i < count; i++ {
		sub, err := sc.Subscribe("foo", func(_ *Msg) {})
		if err != nil {
			t.Fatalf("Unexpected error on subscribe: %v", err)
		}
		checkClose(sub)

		qsub, err := sc.QueueSubscribe("foo", "group", func(_ *Msg) {})
		if err != nil {
			t.Fatalf("Unexpected error on subscribe: %v", err)
		}
		checkClose(qsub)

		if supported {
			// Repeat the tests but pretend server does not support close
			sc.(*conn).Lock()
			sc.(*conn).subCloseRequests = ""
			sc.(*conn).Unlock()
			supported = false
		}
	}

	sub, err := sc.Subscribe("foo", func(_ *Msg) {})
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	closedNC, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("Unexpected error on connect: %v", err)
	}
	closedNC.Close()
	// Swap current NATS connection with this closed connection
	sc.(*conn).Lock()
	savedNC := sc.(*conn).nc
	sc.(*conn).nc = closedNC
	sc.(*conn).Unlock()
	if err := sub.Unsubscribe(); err == nil {
		t.Fatal("Expected error on close")
	}
	// Restore NATS connection
	sc.(*conn).Lock()
	sc.(*conn).nc = savedNC
	sc.(*conn).Unlock()

	sc.Close()

	closeSubscriber(t, "dursub", "sub")
	closeSubscriber(t, "durqueuesub", "queue")
}

func TestOptionNatsName(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()
	sc := NewDefaultConnection(t)
	defer sc.Close()

	// Make sure we can get the STAN-created Conn.
	nc := sc.NatsConn()

	if n := nc.Opts.Name; n != clientName {
		t.Fatalf("Unexpected nats client name: %s", n)
	}
}

func closeSubscriber(t *testing.T, channel, subType string) {
	sc := NewDefaultConnection(t)
	defer sc.Close()

	// Send 1 message
	if err := sc.Publish(channel, []byte("msg")); err != nil {
		t.Fatalf("Unexpected error on publish: %v", err)
	}
	count := 0
	ch := make(chan bool)
	errCh := make(chan bool)
	cb := func(m *Msg) {
		count++
		if m.Sequence != uint64(count) {
			errCh <- true
			return
		}
		ch <- true
	}
	// Create a durable
	var sub Subscription
	var err error
	if subType == "sub" {
		sub, err = sc.Subscribe(channel, cb, DurableName("dur"), DeliverAllAvailable())
	} else {
		sub, err = sc.QueueSubscribe(channel, "group", cb, DurableName("dur"), DeliverAllAvailable())
	}
	if err != nil {
		stackFatalf(t, "Unexpected error on subscribe: %v", err)
	}
	// Wait to receive 1st message
	if err := Wait(ch); err != nil {
		stackFatalf(t, "Did not get our message")
	}
	// Wait a bit to reduce risk of server processing unsubscribe before ACK
	time.Sleep(500 * time.Millisecond)
	// Close durable
	err = sub.Close()
	// Feature supported or not by the server
	supported := sc.(*conn).subCloseRequests != ""
	// If connecting to an older server, error is expected
	if !supported && err != ErrNoServerSupport {
		stackFatalf(t, "Expected %v error, got %v", ErrNoServerSupport, err)
	}
	if !supported {
		// Nothing much to test
		sub.Unsubscribe()
		return
	}
	// Here, server supports feature
	if err != nil {
		stackFatalf(t, "Unexpected error on close: %v", err)
	}
	// Send 2nd message
	if err := sc.Publish(channel, []byte("msg")); err != nil {
		stackFatalf(t, "Unexpected error on publish: %v", err)
	}
	// Restart durable
	if subType == "sub" {
		sub, err = sc.Subscribe(channel, cb, DurableName("dur"), DeliverAllAvailable())
	} else {
		sub, err = sc.QueueSubscribe(channel, "group", cb, DurableName("dur"), DeliverAllAvailable())
	}
	if err != nil {
		stackFatalf(t, "Unexpected error on subscribe: %v", err)
	}
	defer sub.Unsubscribe()
	select {
	case <-errCh:
		stackFatalf(t, "Unexpected message received")
	case <-ch:
	case <-time.After(5 * time.Second):
		stackFatalf(t, "Timeout waiting for messages")
	}
}

func TestDuplicateProcessingOfPubAck(t *testing.T) {
	// We run our tests on Windows VM and this test would fail because
	// server would be a slow consumer. So skipping for now.
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}
	s := RunServer(clusterName)
	defer s.Shutdown()

	// Use a very small timeout purposely
	sc, err := Connect(clusterName, clientName, PubAckWait(time.Millisecond))
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	defer sc.Close()

	total := 10000
	pubAcks := make(map[string]struct{}, total)
	gotBug := false
	errCh := make(chan error)
	msg := []byte("msg")
	count := 0
	done := make(chan bool)
	mu := &sync.Mutex{}

	ackHandler := func(guid string, err error) {
		mu.Lock()
		if gotBug {
			mu.Unlock()
			return
		}
		if _, exist := pubAcks[guid]; exist {
			gotBug = true
			errCh <- fmt.Errorf("Duplicate processing of PubAck %d guid=%v", (count + 1), guid)
			mu.Unlock()
			return
		}
		pubAcks[guid] = struct{}{}
		count++
		if count == total {
			done <- true
		}
		mu.Unlock()
	}
	for i := 0; i < total; i++ {
		sc.PublishAsync("foo", msg, ackHandler)
	}
	select {
	case <-done:
	case e := <-errCh:
		t.Fatal(e)
	case <-time.After(10 * time.Second):
		t.Fatal("Test took too long")
	}
	// If we are here is that we have published `total` messages.
	// Since the bug is about processing duplicate PubAck,
	// wait a bit more.
	select {
	case e := <-errCh:
		t.Fatal(e)
	case <-time.After(100 * time.Millisecond):
		// This is more than the PubAckWait, so we should be good now.
	}
}

func TestSubDelivered(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	total := 10
	count := 0
	ch := make(chan bool)
	sub, err := sc.Subscribe("foo", func(_ *Msg) {
		count++
		if count == total {
			ch <- true
		}
	})
	if err != nil {
		t.Fatalf("Unexpected error on subscriber: %v", err)
	}
	defer sub.Unsubscribe()

	for i := 0; i < total; i++ {
		if err := sc.Publish("foo", []byte("hello")); err != nil {
			t.Fatalf("Unexpected error on publish: %v", err)
		}
	}
	// Wait for all messages
	if err := Wait(ch); err != nil {
		t.Fatal("Did not get our messages")
	}
	if n, err := sub.Delivered(); err != nil || n != int64(total) {
		t.Fatalf("Expected %d messages delivered, got %d, err=%v", total, n, err)
	}
	sub.Unsubscribe()
	if n, err := sub.Delivered(); err != ErrBadSubscription || n != int64(-1) {
		t.Fatalf("Expected ErrBadSubscription, got %d, err=%v", n, err)
	}
}

func TestSubDropped(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	total := 1000
	count := 0
	ch := make(chan bool)
	blocked := make(chan bool)
	ready := make(chan bool)
	sub, err := sc.Subscribe("foo", func(_ *Msg) {
		count++
		if count == 1 {
			ready <- true
			<-blocked
			ch <- true
		}
	})
	if err != nil {
		t.Fatalf("Unexpected error on subscriber: %v", err)
	}
	defer sub.Unsubscribe()

	// Set low pending limits
	sub.SetPendingLimits(1, -1)

	for i := 0; i < total; i++ {
		if err := sc.Publish("foo", []byte("hello")); err != nil {
			t.Fatalf("Unexpected error on publish: %v", err)
		}
	}
	// Wait for sub to receive first message and block
	if err := Wait(ready); err != nil {
		t.Fatal("Did not get our first message")
	}

	// Messages should be dropped
	if n, err := sub.Dropped(); err != nil || n == 0 {
		t.Fatalf("Messages should have been dropped, got %d, err=%v", n, err)
	}

	// Unblock and wait for end
	close(blocked)
	if err := Wait(ch); err != nil {
		t.Fatal("Callback did not return")
	}
	sub.Unsubscribe()
	// Now subscription is closed, this should return error
	if n, err := sub.Dropped(); err != ErrBadSubscription || n != -1 {
		t.Fatalf("Expected ErrBadSubscription, got %d, err=%v", n, err)
	}
}

func TestSubIsValid(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	sub, err := sc.Subscribe("foo", func(_ *Msg) {})
	if err != nil {
		t.Fatalf("Unexpected error on subscriber: %v", err)
	}
	defer sub.Unsubscribe()
	if !sub.IsValid() {
		t.Fatal("Subscription should be valid")
	}
	sub.Unsubscribe()
	if sub.IsValid() {
		t.Fatal("Subscription should not be valid")
	}
}

func TestPingsInvalidOptions(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()
	for _, test := range []struct {
		name string
		opt  Option
	}{
		{
			"Negative interval",
			Pings(-1, 10),
		},
		{
			"Zero interval",
			Pings(-1, 10),
		},
		{
			"Negative maxOut ",
			Pings(1, -1),
		},
		{
			"Too small maxOut",
			Pings(1, 1),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			sc, err := Connect(clusterName, clientName, test.opt)
			if sc != nil {
				sc.Close()
			}
			if err == nil {
				t.Fatalf("Expected error")
			}
		})
	}
}

func pingInMillis(interval int) int {
	return interval * -1
}

func TestPings(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	testAllowMillisecInPings = true
	defer func() { testAllowMillisecInPings = false }()

	// Create a sub on the subject the pings are sent to
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	count := 0
	ch := make(chan bool, 1)
	nc.Subscribe(DefaultDiscoverPrefix+"."+clusterName+".pings", func(m *nats.Msg) {
		count++
		// Wait more than the number of maxOut
		if count == 10 {
			ch <- true
		}
	})

	errCh := make(chan error, 1)
	sc, err := Connect(clusterName, clientName,
		Pings(pingInMillis(50), 5),
		SetConnectionLostHandler(func(sc Conn, err error) {
			errCh <- err
		}))
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	defer sc.Close()

	if err := Wait(ch); err != nil {
		t.Fatal("Did not get our pings")
	}
	// Kill the server and expect the error callback to fire
	s.Shutdown()
	select {
	case e := <-errCh:
		if e != ErrMaxPings {
			t.Fatalf("Expected error %v, got %v", ErrMaxPings, e)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Error callback should have fired")
	}
	// Check that connection is closed.
	c := sc.(*conn)
	c.RLock()
	c.pingMu.Lock()
	timerIsNil := c.pingTimer == nil
	c.pingMu.Unlock()
	c.RUnlock()
	if !timerIsNil {
		t.Fatalf("Expected timer to be nil")
	}
	if sc.NatsConn() != nil {
		t.Fatalf("Expected nats conn to be nil")
	}

	s = RunServer(clusterName)
	defer s.Shutdown()

	sc, err = Connect(clusterName, clientName,
		Pings(pingInMillis(50), 100),
		SetConnectionLostHandler(func(sc Conn, err error) {
			errCh <- err
		}))
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	defer sc.Close()

	// Kill NATS connection, expect different error
	sc.NatsConn().Close()
	select {
	case e := <-errCh:
		if e != nats.ErrConnectionClosed {
			t.Fatalf("Expected error %v, got %v", nats.ErrConnectionClosed, e)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Error callback should have fired")
	}
}

func TestPingsCloseUnlockPubCalls(t *testing.T) {
	ns := natsd.RunDefaultServer()
	defer ns.Shutdown()

	testAllowMillisecInPings = true
	defer func() { testAllowMillisecInPings = false }()

	opts := server.GetDefaultOptions()
	opts.NATSServerURL = nats.DefaultURL
	opts.ID = clusterName
	s := runServerWithOpts(opts)
	defer s.Shutdown()

	sc, err := Connect(clusterName, clientName,
		MaxPubAcksInflight(1),
		Pings(pingInMillis(50), 10))
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	defer sc.Close()

	s.Shutdown()

	total := 100
	ch := make(chan bool, 1)
	ec := int32(0)
	ah := func(g string, e error) {
		if c := atomic.AddInt32(&ec, 1); c == int32(total) {
			ch <- true
		}
	}
	wg := sync.WaitGroup{}
	wg.Add(total)
	for i := 0; i < total/2; i++ {
		go func() {
			sc.PublishAsync("foo", []byte("hello"), ah)
			wg.Done()
		}()
		go func() {
			if err := sc.Publish("foo", []byte("hello")); err != nil {
				if c := atomic.AddInt32(&ec, 1); c == int32(total) {
					ch <- true
				}
			}
			wg.Done()
		}()
	}
	if err := Wait(ch); err != nil {
		t.Fatal("Did not get all the expected failures")
	}
	wg.Wait()
}

func TestConnErrHandlerNotCalledOnNormalClose(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	errCh := make(chan error, 1)
	sc, err := Connect(clusterName, clientName,
		SetConnectionLostHandler(func(_ Conn, err error) {
			errCh <- err
		}))
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	sc.Close()
	select {
	case <-errCh:
		t.Fatalf("ConnErrHandler should not have been invoked in normal close")
	case <-time.After(250 * time.Millisecond):
		// ok
	}
}

type pubFailsOnClientReplacedDialer struct {
	sync.Mutex
	conn net.Conn
	fail bool
	ch   chan bool
}

func (d *pubFailsOnClientReplacedDialer) Dial(network, address string) (net.Conn, error) {
	d.Lock()
	defer d.Unlock()
	if d.fail {
		return nil, fmt.Errorf("error on purpose")
	}
	c, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	d.conn = c
	d.ch <- true
	return c, nil
}

func TestPubFailsOnClientReplaced(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	cd := &pubFailsOnClientReplacedDialer{ch: make(chan bool, 1)}

	nc, err := nats.Connect(nats.DefaultURL,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(50*time.Millisecond),
		nats.SetCustomDialer(cd))
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	defer nc.Close()

	// Consume dial success notification
	<-cd.ch

	sc, err := Connect(clusterName, clientName, NatsConn(nc))
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	// Send a message and ensure it is ok.
	if err := sc.Publish("foo", []byte("hello")); err != nil {
		t.Fatalf("Error on publish: %v", err)
	}

	// Cause failure of client connection
	cd.Lock()
	cd.fail = true
	cd.conn.Close()
	cd.Unlock()

	// Create new client with same client ID
	sc2, err := Connect(clusterName, clientName)
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	// Verify that this client can publish
	if err := sc2.Publish("foo", []byte("hello")); err != nil {
		t.Fatalf("Error on publish: %v", err)
	}

	// Allow first client to "reconnect"
	cd.Lock()
	cd.fail = false
	cd.Unlock()

	// Wait for the reconnect
	<-cd.ch

	// Wait a bit and try to publish
	time.Sleep(50 * time.Millisecond)
	// It should fail
	if err := sc.Publish("foo", []byte("hello")); err == nil {
		t.Fatalf("Publish of first client should have failed")
	}
}

func TestPingsResponseError(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	testAllowMillisecInPings = true
	defer func() { testAllowMillisecInPings = false }()

	cd := &pubFailsOnClientReplacedDialer{ch: make(chan bool, 1)}

	nc, err := nats.Connect(nats.DefaultURL,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(50*time.Millisecond),
		nats.SetCustomDialer(cd))
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	defer nc.Close()

	// Consume dial success notification
	<-cd.ch

	errCh := make(chan error, 1)
	sc, err := Connect(clusterName, clientName,
		NatsConn(nc),
		// Make it big enough so that we get the response error before we reach the max
		Pings(pingInMillis(50), 100),
		SetConnectionLostHandler(func(_ Conn, err error) {
			errCh <- err
		}))
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	// Send a message and ensure it is ok.
	if err := sc.Publish("foo", []byte("hello")); err != nil {
		t.Fatalf("Error on publish: %v", err)
	}

	// Cause failure of client connection
	cd.Lock()
	cd.fail = true
	cd.conn.Close()
	cd.Unlock()

	// Create new client with same client ID
	sc2, err := Connect(clusterName, clientName)
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	// Verify that this client can publish
	if err := sc2.Publish("foo", []byte("hello")); err != nil {
		t.Fatalf("Error on publish: %v", err)
	}

	// Allow first client to "reconnect"
	cd.Lock()
	cd.fail = false
	cd.Unlock()

	// Wait for the reconnect
	<-cd.ch

	// Wait for the error callback
	select {
	case e := <-errCh:
		if !strings.Contains(e.Error(), "replaced") {
			t.Fatalf("Expected error saying that client was replaced, got %v", e)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Error callback not invoked")
	}
}

func TestClientIDAndConnIDInPubMsg(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	// By default, when connecting to a 0.10.0+ server, the PubMsg
	// now only contains a connection ID, not the publish call.
	sc := NewDefaultConnection(t)
	defer sc.Close()

	c := sc.(*conn)
	c.RLock()
	pubSubj := c.pubPrefix
	connID := c.connID
	c.RUnlock()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	defer nc.Close()
	ch := make(chan bool, 1)
	nc.Subscribe(pubSubj+".foo", func(m *nats.Msg) {
		pubMsg := &pb.PubMsg{}
		pubMsg.Unmarshal(m.Data)
		if pubMsg.ClientID == clientName && bytes.Equal(pubMsg.ConnID, connID) {
			ch <- true
		}
	})
	nc.Flush()

	if sc.Publish("foo", []byte("hello")); err != nil {
		t.Fatalf("Error on publish: %v", err)
	}
	// Verify that client ID and ConnID are properly set
	if err := Wait(ch); err != nil {
		t.Fatal("Invalid ClientID and/or ConnID")
	}
}
