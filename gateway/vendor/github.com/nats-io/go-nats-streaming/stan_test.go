package stan

////////////////////////////////////////////////////////////////////////////////
// Package scoped specific tests here..
////////////////////////////////////////////////////////////////////////////////

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
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
	natstest "github.com/nats-io/go-nats/test"
	"github.com/nats-io/nats-streaming-server/server"
	"github.com/nats-io/nats-streaming-server/test"
)

func RunServer(ID string) *server.StanServer {
	return test.RunServer(ID)
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
		if ok == false {
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

func TestTimeoutPublishAsync(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc, err := Connect(clusterName, clientName, PubAckWait(50*time.Millisecond))
	if err != nil {
		t.Fatalf("Expected to connect correctly, got err %v\n", err)
	}
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
	// Give a chance to ACKs to make it to the server.
	// This step is not necessary. Worse could happen is that messages
	// are redelivered. This is why we check on !m.Redelivered in the
	// callback to validate the counts.
	time.Sleep(500 * time.Millisecond)
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
	delta := time.Now().Sub(startTime)

	sub, err = sc.Subscribe("foo", mcb, StartAtTimeDelta(delta))
	if err != nil {
		t.Fatalf("Expected no error on Subscribe, got %v\n", err)
	}
	defer sub.Unsubscribe()

	if err := Wait(ch); err != nil {
		t.Fatal("Did not receive our messages")
	}
}

func TestSubscriptionStartAtWithEmptyStore(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	startTime := time.Now()

	mcb := func(m *Msg) {
	}

	sub, err := sc.Subscribe("foo", mcb, StartAtTime(startTime))
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	sub.Unsubscribe()

	sub, err = sc.Subscribe("foo", mcb, StartAtSequence(0))
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	sub.Unsubscribe()

	sub, err = sc.Subscribe("foo", mcb, StartWithLastReceived())
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	sub.Unsubscribe()

	sub, err = sc.Subscribe("foo", mcb)
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	sub.Unsubscribe()
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

	// test nil
	var nsub *subscription
	err := nsub.Unsubscribe()
	if err == nil || err != ErrBadSubscription {
		t.Fatalf("Expected a bad subscription err, got %v\n", err)
	}

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

func TestSubscribeShrink(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	nsubs := 1000
	subs := make([]Subscription, 0, nsubs)
	for i := 1; i <= nsubs; i++ {
		// Create a valid one
		sub, err := sc.Subscribe("foo", nil)
		if err != nil {
			t.Fatalf("Got an error on subscribe: %v\n", err)
		}
		subs = append(subs, sub)
	}
	// Now unsubsribe them all
	for _, sub := range subs {
		err := sub.Unsubscribe()
		if err != nil {
			t.Fatalf("Got an error on unsubscribe: %v\n", err)
		}
	}
}

func TestDupClientID(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	_, err := Connect(clusterName, clientName, PubAckWait(50*time.Millisecond))
	if err == nil {
		t.Fatalf("Expected to get an error for duplicate clientID\n")
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

	for i := 0; i < 10; i++ {
		sc.Publish("foo", []byte("ok"))
	}

	if err := sc.Publish("foo", []byte("Hello World!")); err == nil || err != ErrConnectionClosed {
		t.Fatalf("Expected an ErrConnectionClosed error on publish to a closed connection, got %v\n", err)
	}

	if err := sub.Unsubscribe(); err == nil || err != ErrConnectionClosed {
		t.Fatalf("Expected an ErrConnectionClosed error on unsubscribe to a closed connection, got %v\n", err)
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

func checkTime(t *testing.T, label string, time1, time2 time.Time, expected time.Duration, tolerance time.Duration) {
	duration := time2.Sub(time1)

	if duration < (expected-tolerance) || duration > (expected+tolerance) {
		t.Fatalf("%s not in range: %v (expected %v +/- %v)", label, duration, expected, tolerance)
	}
}

func testRedelivery(t *testing.T, count int, queueSub bool) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	toSend := int32(count)
	hw := []byte("Hello World")

	ch := make(chan bool)
	acked := int32(0)
	secondRedelivery := false
	firstDeliveryCount := int32(0)
	firstRedeliveryCount := int32(0)
	var startDelivery time.Time
	var startFirstRedelivery time.Time
	var startSecondRedelivery time.Time

	ackRedeliverTime := 1 * time.Second

	recvCb := func(m *Msg) {
		if m.Redelivered {
			if secondRedelivery {
				if startSecondRedelivery.IsZero() {
					startSecondRedelivery = time.Now()
				}
				acks := atomic.AddInt32(&acked, 1)
				if acks <= toSend {
					m.Ack()
					if acks == toSend {
						ch <- true
					}
				}
			} else {
				if startFirstRedelivery.IsZero() {
					startFirstRedelivery = time.Now()
				}
				if atomic.AddInt32(&firstRedeliveryCount, 1) == toSend {
					secondRedelivery = true
				}
			}
		} else {
			if startDelivery.IsZero() {
				startDelivery = time.Now()
			}
			atomic.AddInt32(&firstDeliveryCount, 1)
		}
	}

	var sub Subscription
	var err error
	if queueSub {
		sub, err = sc.QueueSubscribe("foo", "bar", recvCb, AckWait(ackRedeliverTime), SetManualAckMode())
	} else {
		sub, err = sc.Subscribe("foo", recvCb, AckWait(ackRedeliverTime), SetManualAckMode())
	}
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v\n", err)
	}
	defer sub.Unsubscribe()

	for i := int32(0); i < toSend; i++ {
		sc.Publish("foo", hw)
	}

	// If this succeeds, it means that we got all messages first delivered,
	// and then at least 2 * toSend messages received as redelivered.
	if err := Wait(ch); err != nil {
		t.Fatal("Did not ack all expected messages")
	}

	// Wait a period and bit more to make sure that no more message are
	// redelivered (acked will then be > toSend)
	time.Sleep(ackRedeliverTime + 100*time.Millisecond)

	// Verify first redelivery happens when expected
	checkTime(t, "First redelivery", startDelivery, startFirstRedelivery, ackRedeliverTime, ackRedeliverTime/2)

	// Verify second redelivery happens when expected
	checkTime(t, "Second redelivery", startFirstRedelivery, startSecondRedelivery, ackRedeliverTime, ackRedeliverTime/2)

	// Check counts
	if delivered := atomic.LoadInt32(&firstDeliveryCount); delivered != toSend {
		t.Fatalf("Did not receive all messages during delivery: %v vs %v", delivered, toSend)
	}
	if firstRedelivered := atomic.LoadInt32(&firstRedeliveryCount); firstRedelivered != toSend {
		t.Fatalf("Did not receive all messages during first redelivery: %v vs %v", firstRedelivered, toSend)
	}
	if acks := atomic.LoadInt32(&acked); acks != toSend {
		t.Fatalf("Did not get expected acks: %v vs %v", acks, toSend)
	}
}

func TestLowRedeliveryToSubMoreThanOnce(t *testing.T) {
	testRedelivery(t, 10, false)
}

func TestHighRedeliveryToSubMoreThanOnce(t *testing.T) {
	testRedelivery(t, 100, false)
}

func TestLowRedeliveryToQueueSubMoreThanOnce(t *testing.T) {
	testRedelivery(t, 10, true)
}

func TestHighRedeliveryToQueueSubMoreThanOnce(t *testing.T) {
	testRedelivery(t, 100, true)
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
			// Reduce risk of test failure by allowing server to
			// process acks before processing Close() requesting
			time.Sleep(time.Second)
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

func TestPubMultiQueueSub(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	ch := make(chan bool)
	received := int32(0)
	s1Received := int32(0)
	s2Received := int32(0)
	toSend := int32(1000)

	var s1, s2 Subscription

	msgMapLock := &sync.Mutex{}
	msgMap := make(map[uint64]struct{})

	mcb := func(m *Msg) {
		// Remember the message sequence.
		msgMapLock.Lock()
		if _, ok := msgMap[m.Sequence]; ok {
			t.Fatalf("Detected duplicate for sequence: %d\n", m.Sequence)
		}
		msgMap[m.Sequence] = struct{}{}
		msgMapLock.Unlock()
		// Track received for each receiver.
		if m.Sub == s1 {
			atomic.AddInt32(&s1Received, 1)
		} else if m.Sub == s2 {
			atomic.AddInt32(&s2Received, 1)
		} else {
			t.Fatalf("Received message on unknown subscription")
		}
		// Track total
		if nr := atomic.AddInt32(&received, 1); nr == int32(toSend) {
			ch <- true
		}
	}

	s1, err := sc.QueueSubscribe("foo", "bar", mcb)
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer s1.Unsubscribe()

	s2, err = sc.QueueSubscribe("foo", "bar", mcb)
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer s2.Unsubscribe()

	// Publish out the messages.
	for i := int32(0); i < toSend; i++ {
		data := []byte(fmt.Sprintf("%d", i))
		sc.Publish("foo", data)
	}
	if err := WaitTime(ch, 10*time.Second); err != nil {
		t.Fatal("Did not receive our messages")
	}

	if nr := atomic.LoadInt32(&received); nr != toSend {
		t.Fatalf("Did not receive correct number of messages: %d vs %d\n", nr, toSend)
	}

	s1r := atomic.LoadInt32(&s1Received)
	s2r := atomic.LoadInt32(&s2Received)

	v := uint(float32(toSend) * 0.25) // 25 percent
	expected := toSend / 2
	d1 := uint(math.Abs(float64(expected - s1r)))
	d2 := uint(math.Abs(float64(expected - s2r)))
	if d1 > v || d2 > v {
		t.Fatalf("Too much variance in totals: %d, %d > %d", d1, d2, v)
	}
}

func TestPubMultiQueueSubWithSlowSubscriber(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	ch := make(chan bool)
	s2BlockedCh := make(chan bool)
	received := int32(0)
	s1Received := int32(0)
	s2Received := int32(0)
	toSend := int32(100)

	var s1, s2 Subscription

	msgMapLock := &sync.Mutex{}
	msgMap := make(map[uint64]struct{})

	mcb := func(m *Msg) {
		// Remember the message sequence.
		msgMapLock.Lock()
		if _, ok := msgMap[m.Sequence]; ok {
			t.Fatalf("Detected duplicate for sequence: %d\n", m.Sequence)
		}
		msgMap[m.Sequence] = struct{}{}
		msgMapLock.Unlock()
		// Track received for each receiver.
		if m.Sub == s1 {
			atomic.AddInt32(&s1Received, 1)
		} else if m.Sub == s2 {
			// Block this subscriber
			<-s2BlockedCh
			atomic.AddInt32(&s2Received, 1)
		} else {
			t.Fatalf("Received message on unknown subscription")
		}
		// Track total
		if nr := atomic.AddInt32(&received, 1); nr == int32(toSend) {
			ch <- true
		}
	}

	s1, err := sc.QueueSubscribe("foo", "bar", mcb)
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer s1.Unsubscribe()

	s2, err = sc.QueueSubscribe("foo", "bar", mcb)
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer s2.Unsubscribe()

	// Publish out the messages.
	for i := int32(0); i < toSend; i++ {
		data := []byte(fmt.Sprintf("%d", i))
		sc.Publish("foo", data)
	}

	close(s2BlockedCh)

	if err := WaitTime(ch, 10*time.Second); err != nil {
		t.Fatal("Did not receive our messages")
	}

	if nr := atomic.LoadInt32(&received); nr != toSend {
		t.Fatalf("Did not receive correct number of messages: %d vs %d\n", nr, toSend)
	}

	s1r := atomic.LoadInt32(&s1Received)
	s2r := atomic.LoadInt32(&s2Received)

	// We have no guarantee that s2 received only 1 or 2 messages, but it should
	// not have received more than half
	if s2r > toSend/2 {
		t.Fatalf("Expected sub2 to receive no more than half, got %d\n", s2r)
	}

	if s1r != toSend-s2r {
		t.Fatalf("Expected %d msgs for sub1, got %d\n", toSend-s2r, s1r)
	}
}

func TestPubMultiQueueSubWithRedelivery(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	ch := make(chan bool)
	received := int32(0)
	s1Received := int32(0)
	toSend := int32(50)

	var s1, s2 Subscription

	mcb := func(m *Msg) {
		// Track received for each receiver.

		if m.Sub == s1 {
			m.Ack()
			atomic.AddInt32(&s1Received, 1)

			// Track total only for sub1
			if nr := atomic.AddInt32(&received, 1); nr == int32(toSend) {
				ch <- true
			}
		} else if m.Sub == s2 {
			// We will not ack this subscriber
		} else {
			t.Fatalf("Received message on unknown subscription")
		}
	}

	s1, err := sc.QueueSubscribe("foo", "bar", mcb, SetManualAckMode())
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer s1.Unsubscribe()

	s2, err = sc.QueueSubscribe("foo", "bar", mcb, SetManualAckMode(), AckWait(1*time.Second))
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer s2.Unsubscribe()

	// Publish out the messages.
	for i := int32(0); i < toSend; i++ {
		data := []byte(fmt.Sprintf("%d", i))
		sc.Publish("foo", data)
	}
	if err := WaitTime(ch, 30*time.Second); err != nil {
		t.Fatal("Did not receive our messages")
	}

	if nr := atomic.LoadInt32(&received); nr != toSend {
		t.Fatalf("Did not receive correct number of messages: %d vs %d\n", nr, toSend)
	}
}

func TestPubMultiQueueSubWithDelayRedelivery(t *testing.T) {
	// Run a NATS Streaming server
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	ch := make(chan bool)
	toSend := int32(500)
	ackCount := int32(0)

	var s1, s2 Subscription

	mcb := func(m *Msg) {
		// Track received for each receiver.
		if m.Sub == s1 {

			m.Ack()

			// if we've acked everything, signal
			nr := atomic.AddInt32(&ackCount, 1)

			if nr == int32(toSend) {
				ch <- true
			}

			if nr > 0 && nr%(toSend/2) == 0 {

				// This depends on the internal algorithm where the
				// best resend subscriber is the one with the least number
				// of outstanding acks.
				//
				// Sleep to allow the acks to back up, so s2 will look
				// like a better subscriber to send messages to.
				time.Sleep(time.Millisecond * 200)
			}
		} else if m.Sub == s2 {
			// We will not ack this subscriber
		} else {
			t.Fatalf("Received message on unknown subscription")
		}
	}

	s1, err := sc.QueueSubscribe("foo", "bar", mcb, SetManualAckMode())
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer s1.Unsubscribe()

	s2, err = sc.QueueSubscribe("foo", "bar", mcb, SetManualAckMode(), AckWait(1*time.Second))
	if err != nil {
		t.Fatalf("Unexpected error on Subscribe, got %v", err)
	}
	defer s2.Unsubscribe()

	// Publish out the messages.
	for i := int32(0); i < toSend; i++ {
		data := []byte(fmt.Sprintf("%d", i))
		sc.Publish("foo", data)
	}

	if err := WaitTime(ch, 30*time.Second); err != nil {
		t.Fatalf("Did not ack expected count of messages: %v", toSend)
	}

	if nr := atomic.LoadInt32(&ackCount); nr != toSend {
		t.Fatalf("Did not ack the correct number of messages: %d vs %d\n", nr, toSend)
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
	s := server.RunServerWithOpts(opts, nil)
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
		t.Fatalf("Expected an error signalling too many channels\n")
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
	cnc := nats.Conn{Opts: nats.DefaultOptions}
	sc, err := Connect(clusterName, clientName, NatsConn(&cnc))
	if err != ErrBadConnection {
		stackFatalf(t, "Expected to get an invalid connection error, got %v", err)
	}

	// Allow custom conn only if already connected
	opts := nats.DefaultOptions
	nc, err = opts.Connect()
	if err != nil {
		stackFatalf(t, "Expected to connect correctly, got err %v", err)
	}
	sc, err = Connect(clusterName, clientName, NatsConn(nc))
	if err != nil {
		stackFatalf(t, "Expected to connect correctly, got err %v", err)
	}
	nc.Close()
	if nc.Status() != nats.CLOSED {
		t.Fatal("Should have status set to CLOSED")
	}

	// Make sure we can get the Conn we provide.
	nc = natstest.NewDefaultConnection(t)
	sc, err = Connect(clusterName, clientName, NatsConn(nc))
	if err != nil {
		stackFatalf(t, "Expected to connect correctly, got err %v", err)
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

func TestTimeoutOnRequests(t *testing.T) {
	ns := natsd.RunDefaultServer()
	defer ns.Shutdown()

	opts := server.GetDefaultOptions()
	opts.ID = clusterName
	opts.NATSServerURL = nats.DefaultURL
	s := server.RunServerWithOpts(opts, nil)
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

func TestSubscriberClose(t *testing.T) {
	s := RunServer(clusterName)
	defer s.Shutdown()

	sc := NewDefaultConnection(t)
	defer sc.Close()

	// If old server, Close() is expected to fail.
	supported := sc.(*conn).subCloseRequests != ""

	sub, err := sc.Subscribe("foo", func(_ *Msg) {})
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	err = sub.Close()
	if supported && err != nil {
		t.Fatalf("Unexpected error on close: %v", err)
	} else if !supported && err != ErrNoServerSupport {
		t.Fatalf("Expected %v, got %v", ErrNoServerSupport, err)
	}

	sub, err = sc.QueueSubscribe("foo", "group", func(_ *Msg) {})
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	err = sub.Close()
	if supported && err != nil {
		t.Fatalf("Unexpected error on close: %v", err)
	} else if !supported && err != ErrNoServerSupport {
		t.Fatalf("Expected %v, got %v", ErrNoServerSupport, err)
	}

	sc.Close()

	closeSubscriber(t, "dursub", "sub")
	closeSubscriber(t, "durqueuesub", "queue")
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
