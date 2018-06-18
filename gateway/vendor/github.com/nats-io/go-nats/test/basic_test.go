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
	"bytes"
	"math"
	"regexp"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nats-io/go-nats"
)

func TestCloseLeakingGoRoutines(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	// Give time for things to settle before capturing the number of
	// go routines
	time.Sleep(500 * time.Millisecond)

	base := runtime.NumGoroutine()

	nc := NewDefaultConnection(t)

	nc.Flush()
	nc.Close()

	// Give time for things to settle before capturing the number of
	// go routines
	time.Sleep(500 * time.Millisecond)

	delta := (runtime.NumGoroutine() - base)
	if delta > 0 {
		t.Fatalf("%d Go routines still exist post Close()", delta)
	}
	// Make sure we can call Close() multiple times
	nc.Close()
}

func TestLeakingGoRoutinesOnFailedConnect(t *testing.T) {
	// Give time for things to settle before capturing the number of
	// go routines
	time.Sleep(500 * time.Millisecond)

	base := runtime.NumGoroutine()

	nc, err := nats.Connect(nats.DefaultURL)
	if err == nil {
		nc.Close()
		t.Fatalf("Expected failure to connect")
	}

	// Give time for things to settle before capturing the number of
	// go routines
	time.Sleep(500 * time.Millisecond)

	delta := (runtime.NumGoroutine() - base)
	if delta > 0 {
		t.Fatalf("%d Go routines still exist post Close()", delta)
	}
}

func TestConnectedServer(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	u := nc.ConnectedUrl()
	if u == "" || u != nats.DefaultURL {
		t.Fatalf("Unexpected connected URL of %s\n", u)
	}
	srv := nc.ConnectedServerId()
	if srv == "" {
		t.Fatal("Expected a connected server id")
	}
	nc.Close()
	u = nc.ConnectedUrl()
	if u != "" {
		t.Fatalf("Expected a nil connected URL, got %s\n", u)
	}
	srv = nc.ConnectedServerId()
	if srv != "" {
		t.Fatalf("Expected a nil connect server, got %s\n", srv)
	}
}

func TestMultipleClose(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			nc.Close()
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestBadOptionTimeoutConnect(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	opts := nats.GetDefaultOptions()
	opts.Timeout = -1
	opts.Url = "nats://localhost:4222"

	_, err := opts.Connect()
	if err == nil {
		t.Fatal("Expected an error")
	}
	if err != nats.ErrNoServers {
		t.Fatalf("Expected a ErrNoServers error: Got %v\n", err)
	}
}

func TestSimplePublish(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	if err := nc.Publish("foo", []byte("Hello World")); err != nil {
		t.Fatal("Failed to publish string message: ", err)
	}
}

func TestSimplePublishNoData(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	if err := nc.Publish("foo", nil); err != nil {
		t.Fatal("Failed to publish empty message: ", err)
	}
}

func TestPublishDoesNotFailOnSlowConsumer(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	sub, err := nc.SubscribeSync("foo")
	if err != nil {
		t.Fatalf("Unable to create subscription: %v", err)
	}

	if err := sub.SetPendingLimits(1, 1000); err != nil {
		t.Fatalf("Unable to set pending limits: %v", err)
	}

	var pubErr error

	msg := []byte("Hello")
	for i := 0; i < 10; i++ {
		pubErr = nc.Publish("foo", msg)
		if pubErr != nil {
			break
		}
		nc.Flush()
	}

	if pubErr != nil {
		t.Fatalf("Publish() should not fail because of slow consumer. Got '%v'", pubErr)
	}
}

func TestAsyncSubscribe(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	omsg := []byte("Hello World")
	ch := make(chan bool)

	// Callback is mandatory
	if _, err := nc.Subscribe("foo", nil); err == nil {
		t.Fatal("Creating subscription without callback should have failed")
	}

	_, err := nc.Subscribe("foo", func(m *nats.Msg) {
		if !bytes.Equal(m.Data, omsg) {
			t.Fatal("Message received does not match")
		}
		if m.Sub == nil {
			t.Fatal("Callback does not have a valid Subscription")
		}
		ch <- true
	})
	if err != nil {
		t.Fatal("Failed to subscribe: ", err)
	}
	nc.Publish("foo", omsg)
	if e := Wait(ch); e != nil {
		t.Fatal("Message not received for subscription")
	}
}

func TestAsyncSubscribeRoutineLeakOnUnsubscribe(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	ch := make(chan bool)

	// Give time for things to settle before capturing the number of
	// go routines
	time.Sleep(500 * time.Millisecond)

	// Take the base once the connection is established, but before
	// the subscriber is created.
	base := runtime.NumGoroutine()

	sub, err := nc.Subscribe("foo", func(m *nats.Msg) { ch <- true })
	if err != nil {
		t.Fatal("Failed to subscribe: ", err)
	}

	// Send to ourself
	nc.Publish("foo", []byte("hello"))

	// This ensures that the async delivery routine is up and running.
	if err := Wait(ch); err != nil {
		t.Fatal("Failed to receive message")
	}

	// Make sure to give it time to go back into wait
	time.Sleep(200 * time.Millisecond)

	// Explicit unsubscribe
	sub.Unsubscribe()

	// Give time for things to settle before capturing the number of
	// go routines
	time.Sleep(500 * time.Millisecond)

	delta := (runtime.NumGoroutine() - base)
	if delta > 0 {
		t.Fatalf("%d Go routines still exist post Unsubscribe()", delta)
	}
}

func TestAsyncSubscribeRoutineLeakOnClose(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	ch := make(chan bool)

	// Give time for things to settle before capturing the number of
	// go routines
	time.Sleep(500 * time.Millisecond)

	// Take the base before creating the connection, since we are going
	// to close it before taking the delta.
	base := runtime.NumGoroutine()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	_, err := nc.Subscribe("foo", func(m *nats.Msg) { ch <- true })
	if err != nil {
		t.Fatal("Failed to subscribe: ", err)
	}

	// Send to ourself
	nc.Publish("foo", []byte("hello"))

	// This ensures that the async delivery routine is up and running.
	if err := Wait(ch); err != nil {
		t.Fatal("Failed to receive message")
	}

	// Make sure to give it time to go back into wait
	time.Sleep(200 * time.Millisecond)

	// Close connection without explicit unsubscribe
	nc.Close()

	// Give time for things to settle before capturing the number of
	// go routines
	time.Sleep(500 * time.Millisecond)

	delta := (runtime.NumGoroutine() - base)
	if delta > 0 {
		t.Fatalf("%d Go routines still exist post Close()", delta)
	}
}

func TestSyncSubscribe(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	sub, err := nc.SubscribeSync("foo")
	if err != nil {
		t.Fatal("Failed to subscribe: ", err)
	}
	omsg := []byte("Hello World")
	nc.Publish("foo", omsg)
	msg, err := sub.NextMsg(1 * time.Second)
	if err != nil || !bytes.Equal(msg.Data, omsg) {
		t.Fatal("Message received does not match")
	}
}

func TestPubSubWithReply(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	sub, err := nc.SubscribeSync("foo")
	if err != nil {
		t.Fatal("Failed to subscribe: ", err)
	}
	omsg := []byte("Hello World")
	nc.PublishMsg(&nats.Msg{Subject: "foo", Reply: "bar", Data: omsg})
	msg, err := sub.NextMsg(10 * time.Second)
	if err != nil || !bytes.Equal(msg.Data, omsg) {
		t.Fatal("Message received does not match")
	}
}

func TestFlush(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	omsg := []byte("Hello World")
	for i := 0; i < 10000; i++ {
		nc.Publish("flush", omsg)
	}
	if err := nc.FlushTimeout(0); err == nil {
		t.Fatal("Calling FlushTimeout() with invalid timeout should fail")
	}
	if err := nc.Flush(); err != nil {
		t.Fatalf("Received error from flush: %s\n", err)
	}
	if nb, _ := nc.Buffered(); nb > 0 {
		t.Fatalf("Outbound buffer not empty: %d bytes\n", nb)
	}

	nc.Close()
	if _, err := nc.Buffered(); err == nil {
		t.Fatal("Calling Buffered() on closed connection should fail")
	}
}

func TestQueueSubscriber(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	s1, _ := nc.QueueSubscribeSync("foo", "bar")
	s2, _ := nc.QueueSubscribeSync("foo", "bar")
	omsg := []byte("Hello World")
	nc.Publish("foo", omsg)
	nc.Flush()
	r1, _ := s1.QueuedMsgs()
	r2, _ := s2.QueuedMsgs()
	if (r1 + r2) != 1 {
		t.Fatal("Received too many messages for multiple queue subscribers")
	}
	// Drain messages
	s1.NextMsg(time.Second)
	s2.NextMsg(time.Second)

	total := 1000
	for i := 0; i < total; i++ {
		nc.Publish("foo", omsg)
	}
	nc.Flush()
	v := uint(float32(total) * 0.15)
	r1, _ = s1.QueuedMsgs()
	r2, _ = s2.QueuedMsgs()
	if r1+r2 != total {
		t.Fatalf("Incorrect number of messages: %d vs %d", (r1 + r2), total)
	}
	expected := total / 2
	d1 := uint(math.Abs(float64(expected - r1)))
	d2 := uint(math.Abs(float64(expected - r2)))
	if d1 > v || d2 > v {
		t.Fatalf("Too much variance in totals: %d, %d > %d", d1, d2, v)
	}
}

func TestReplyArg(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	ch := make(chan bool)
	replyExpected := "bar"

	nc.Subscribe("foo", func(m *nats.Msg) {
		if m.Reply != replyExpected {
			t.Fatalf("Did not receive correct reply arg in callback: "+
				"('%s' vs '%s')", m.Reply, replyExpected)
		}
		ch <- true
	})
	nc.PublishMsg(&nats.Msg{Subject: "foo", Reply: replyExpected, Data: []byte("Hello")})
	if e := Wait(ch); e != nil {
		t.Fatal("Did not receive callback")
	}
}

func TestSyncReplyArg(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	replyExpected := "bar"
	sub, _ := nc.SubscribeSync("foo")
	nc.PublishMsg(&nats.Msg{Subject: "foo", Reply: replyExpected, Data: []byte("Hello")})
	msg, err := sub.NextMsg(1 * time.Second)
	if err != nil {
		t.Fatal("Received an err on NextMsg()")
	}
	if msg.Reply != replyExpected {
		t.Fatalf("Did not receive correct reply arg in callback: "+
			"('%s' vs '%s')", msg.Reply, replyExpected)
	}
}

func TestUnsubscribe(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	received := int32(0)
	max := int32(10)
	ch := make(chan bool)
	nc.Subscribe("foo", func(m *nats.Msg) {
		atomic.AddInt32(&received, 1)
		if received == max {
			err := m.Sub.Unsubscribe()
			if err != nil {
				t.Fatal("Unsubscribe failed with err:", err)
			}
			ch <- true
		}
	})
	send := 20
	for i := 0; i < send; i++ {
		nc.Publish("foo", []byte("hello"))
	}
	nc.Flush()
	<-ch

	r := atomic.LoadInt32(&received)
	if r != max {
		t.Fatalf("Received wrong # of messages after unsubscribe: %d vs %d",
			r, max)
	}
}

func TestDoubleUnsubscribe(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	sub, err := nc.SubscribeSync("foo")
	if err != nil {
		t.Fatal("Failed to subscribe: ", err)
	}
	if err = sub.Unsubscribe(); err != nil {
		t.Fatal("Unsubscribe failed with err:", err)
	}
	if err = sub.Unsubscribe(); err == nil {
		t.Fatal("Unsubscribe should have reported an error")
	}
}

func TestRequestTimeout(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	if _, err := nc.Request("foo", []byte("help"), 10*time.Millisecond); err == nil {
		t.Fatalf("Expected to receive a timeout error")
	}
}

func TestOldRequest(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc, err := nats.Connect(nats.DefaultURL, nats.UseOldRequestStyle())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer nc.Close()

	response := []byte("I will help you")
	nc.Subscribe("foo", func(m *nats.Msg) {
		nc.Publish(m.Reply, response)
	})
	msg, err := nc.Request("foo", []byte("help"), 500*time.Millisecond)
	if err != nil {
		t.Fatalf("Received an error on Request test: %s", err)
	}
	if !bytes.Equal(msg.Data, response) {
		t.Fatalf("Received invalid response")
	}
}

func TestRequest(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	response := []byte("I will help you")
	nc.Subscribe("foo", func(m *nats.Msg) {
		nc.Publish(m.Reply, response)
	})
	msg, err := nc.Request("foo", []byte("help"), 500*time.Millisecond)
	if err != nil {
		t.Fatalf("Received an error on Request test: %s", err)
	}
	if !bytes.Equal(msg.Data, response) {
		t.Fatalf("Received invalid response")
	}
}

func TestRequestNoBody(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	response := []byte("I will help you")
	nc.Subscribe("foo", func(m *nats.Msg) {
		nc.Publish(m.Reply, response)
	})
	msg, err := nc.Request("foo", nil, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("Received an error on Request test: %s", err)
	}
	if !bytes.Equal(msg.Data, response) {
		t.Fatalf("Received invalid response")
	}
}

func TestSimultaneousRequests(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	response := []byte("I will help you")
	nc.Subscribe("foo", func(m *nats.Msg) {
		nc.Publish(m.Reply, response)
	})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			if _, err := nc.Request("foo", nil, 2*time.Second); err != nil {
				t.Fatalf("Expected to receive a timeout error")
			} else {
				wg.Done()
			}
		}()
	}
	wg.Wait()
}

func TestRequestClose(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		nc.Close()
	}()
	if _, err := nc.Request("foo", []byte("help"), 2*time.Second); err != nats.ErrInvalidConnection && err != nats.ErrConnectionClosed {
		t.Fatalf("Expected connection error: got %v", err)
	}
	wg.Wait()
}

func TestRequestCloseTimeout(t *testing.T) {
	// Make sure we return a timeout when we close
	// the connection even if response is queued.

	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	response := []byte("I will help you")
	nc.Subscribe("foo", func(m *nats.Msg) {
		nc.Publish(m.Reply, response)
		nc.Close()
	})
	if _, err := nc.Request("foo", nil, 1*time.Second); err == nil {
		t.Fatalf("Expected to receive a timeout error")
	}
}

func TestFlushInCB(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	ch := make(chan bool)

	nc.Subscribe("foo", func(_ *nats.Msg) {
		nc.Flush()
		ch <- true
	})
	nc.Publish("foo", []byte("Hello"))
	if e := Wait(ch); e != nil {
		t.Fatal("Flush did not return properly in callback")
	}
}

func TestReleaseFlush(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)

	for i := 0; i < 1000; i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	go nc.Close()
	nc.Flush()
}

func TestInbox(t *testing.T) {
	inbox := nats.NewInbox()
	if matched, _ := regexp.Match(`_INBOX.\S`, []byte(inbox)); !matched {
		t.Fatal("Bad INBOX format")
	}
}

func TestStats(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	data := []byte("The quick brown fox jumped over the lazy dog")
	iter := 10

	for i := 0; i < iter; i++ {
		nc.Publish("foo", data)
	}

	if nc.OutMsgs != uint64(iter) {
		t.Fatalf("Not properly tracking OutMsgs: received %d, wanted %d\n", nc.OutMsgs, iter)
	}
	obb := uint64(iter * len(data))
	if nc.OutBytes != obb {
		t.Fatalf("Not properly tracking OutBytes: received %d, wanted %d\n", nc.OutBytes, obb)
	}

	// Clear outbound
	nc.OutMsgs, nc.OutBytes = 0, 0

	// Test both sync and async versions of subscribe.
	nc.Subscribe("foo", func(_ *nats.Msg) {})
	nc.SubscribeSync("foo")

	for i := 0; i < iter; i++ {
		nc.Publish("foo", data)
	}
	nc.Flush()

	if nc.InMsgs != uint64(2*iter) {
		t.Fatalf("Not properly tracking InMsgs: received %d, wanted %d\n", nc.InMsgs, 2*iter)
	}

	ibb := 2 * obb
	if nc.InBytes != ibb {
		t.Fatalf("Not properly tracking InBytes: received %d, wanted %d\n", nc.InBytes, ibb)
	}
}

func TestRaceSafeStats(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	go nc.Publish("foo", []byte("Hello World"))
	time.Sleep(200 * time.Millisecond)

	stats := nc.Stats()

	if stats.OutMsgs != uint64(1) {
		t.Fatalf("Not properly tracking OutMsgs: received %d, wanted %d\n", nc.OutMsgs, 1)
	}
}

func TestBadSubject(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()
	nc := NewDefaultConnection(t)
	defer nc.Close()

	err := nc.Publish("", []byte("Hello World"))
	if err == nil {
		t.Fatalf("Expected an error on bad subject to publish")
	}
	if err != nats.ErrBadSubject {
		t.Fatalf("Expected a ErrBadSubject error: Got %v\n", err)
	}
}

func TestOptions(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc, err := nats.Connect(nats.DefaultURL, nats.Name("myName"), nats.MaxReconnects(2), nats.ReconnectWait(50*time.Millisecond))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer nc.Close()

	rch := make(chan bool)
	cch := make(chan bool)

	nc.SetReconnectHandler(func(_ *nats.Conn) { rch <- true })
	nc.SetClosedHandler(func(_ *nats.Conn) { cch <- true })

	s.Shutdown()

	s = RunDefaultServer()
	defer s.Shutdown()

	if err := Wait(rch); err != nil {
		t.Fatal("Failed getting reconnected cb")
	}

	nc.Close()

	if err := Wait(cch); err != nil {
		t.Fatal("Failed getting closed cb")
	}

	nc, err = nats.Connect(nats.DefaultURL, nats.NoReconnect())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer nc.Close()

	nc.SetReconnectHandler(func(_ *nats.Conn) { rch <- true })
	nc.SetClosedHandler(func(_ *nats.Conn) { cch <- true })

	s.Shutdown()

	// We should not get a reconnect cb this time
	if err := WaitTime(rch, time.Second); err == nil {
		t.Fatal("Unexpected reconnect cb")
	}

	nc.Close()

	if err := Wait(cch); err != nil {
		t.Fatal("Failed getting closed cb")
	}
}

func TestNilConnection(t *testing.T) {
	var nc *nats.Conn
	data := []byte("ok")

	// Publish
	if err := nc.Publish("foo", data); err == nil || err != nats.ErrInvalidConnection {
		t.Fatalf("Expected ErrInvalidConnection error, got %v\n", err)
	}
	if err := nc.PublishMsg(nil); err == nil || err != nats.ErrInvalidMsg {
		t.Fatalf("Expected ErrInvalidMsg error, got %v\n", err)
	}
	if err := nc.PublishMsg(&nats.Msg{}); err == nil || err != nats.ErrInvalidConnection {
		t.Fatalf("Expected ErrInvalidConnection error, got %v\n", err)
	}
	if err := nc.PublishRequest("foo", "reply", data); err == nil || err != nats.ErrInvalidConnection {
		t.Fatalf("Expected ErrInvalidConnection error, got %v\n", err)
	}

	// Subscribe
	if _, err := nc.Subscribe("foo", nil); err == nil || err != nats.ErrInvalidConnection {
		t.Fatalf("Expected ErrInvalidConnection error, got %v\n", err)
	}
	if _, err := nc.SubscribeSync("foo"); err == nil || err != nats.ErrInvalidConnection {
		t.Fatalf("Expected ErrInvalidConnection error, got %v\n", err)
	}
	if _, err := nc.QueueSubscribe("foo", "bar", nil); err == nil || err != nats.ErrInvalidConnection {
		t.Fatalf("Expected ErrInvalidConnection error, got %v\n", err)
	}
	ch := make(chan *nats.Msg)
	if _, err := nc.ChanSubscribe("foo", ch); err == nil || err != nats.ErrInvalidConnection {
		t.Fatalf("Expected ErrInvalidConnection error, got %v\n", err)
	}
	if _, err := nc.ChanQueueSubscribe("foo", "bar", ch); err == nil || err != nats.ErrInvalidConnection {
		t.Fatalf("Expected ErrInvalidConnection error, got %v\n", err)
	}
	if _, err := nc.QueueSubscribeSyncWithChan("foo", "bar", ch); err == nil || err != nats.ErrInvalidConnection {
		t.Fatalf("Expected ErrInvalidConnection error, got %v\n", err)
	}

	// Flush
	if err := nc.Flush(); err == nil || err != nats.ErrInvalidConnection {
		t.Fatalf("Expected ErrInvalidConnection error, got %v\n", err)
	}
	if err := nc.FlushTimeout(time.Millisecond); err == nil || err != nats.ErrInvalidConnection {
		t.Fatalf("Expected ErrInvalidConnection error, got %v\n", err)
	}

	// Nil Subscribers
	var sub *nats.Subscription
	if sub.Type() != nats.NilSubscription {
		t.Fatalf("Got wrong type for nil subscription, %v\n", sub.Type())
	}
	if sub.IsValid() {
		t.Fatalf("Expected IsValid() to return false")
	}
	if err := sub.Unsubscribe(); err == nil || err != nats.ErrBadSubscription {
		t.Fatalf("Expected Unsubscribe to return proper error, got %v\n", err)
	}
	if err := sub.AutoUnsubscribe(1); err == nil || err != nats.ErrBadSubscription {
		t.Fatalf("Expected ErrBadSubscription error, got %v\n", err)
	}
	if _, err := sub.NextMsg(time.Millisecond); err == nil || err != nats.ErrBadSubscription {
		t.Fatalf("Expected ErrBadSubscription error, got %v\n", err)
	}
	if _, err := sub.QueuedMsgs(); err == nil || err != nats.ErrBadSubscription {
		t.Fatalf("Expected ErrBadSubscription error, got %v\n", err)
	}
	if _, _, err := sub.Pending(); err == nil || err != nats.ErrBadSubscription {
		t.Fatalf("Expected ErrBadSubscription error, got %v\n", err)
	}
	if _, _, err := sub.MaxPending(); err == nil || err != nats.ErrBadSubscription {
		t.Fatalf("Expected ErrBadSubscription error, got %v\n", err)
	}
	if err := sub.ClearMaxPending(); err == nil || err != nats.ErrBadSubscription {
		t.Fatalf("Expected ErrBadSubscription error, got %v\n", err)
	}
	if _, _, err := sub.PendingLimits(); err == nil || err != nats.ErrBadSubscription {
		t.Fatalf("Expected ErrBadSubscription error, got %v\n", err)
	}
	if err := sub.SetPendingLimits(1, 1); err == nil || err != nats.ErrBadSubscription {
		t.Fatalf("Expected ErrBadSubscription error, got %v\n", err)
	}
	if _, err := sub.Delivered(); err == nil || err != nats.ErrBadSubscription {
		t.Fatalf("Expected ErrBadSubscription error, got %v\n", err)
	}
	if _, err := sub.Dropped(); err == nil || err != nats.ErrBadSubscription {
		t.Fatalf("Expected ErrBadSubscription error, got %v\n", err)
	}
}
