// Copyright 2013-2018 The NATS Authors
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
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nats-io/go-nats"
)

// More advanced tests on subscriptions

func TestServerAutoUnsub(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()
	received := int32(0)
	max := int32(10)

	// Call this to make sure that we have everything setup connection wise
	nc.Flush()

	base := runtime.NumGoroutine()

	sub, err := nc.Subscribe("foo", func(_ *nats.Msg) {
		atomic.AddInt32(&received, 1)
	})
	if err != nil {
		t.Fatal("Failed to subscribe: ", err)
	}
	sub.AutoUnsubscribe(int(max))
	total := 100
	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	nc.Flush()
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&received) != max {
		t.Fatalf("Received %d msgs, wanted only %d\n", received, max)
	}
	if sub.IsValid() {
		t.Fatal("Expected subscription to be invalid after hitting max")
	}
	if err := sub.AutoUnsubscribe(10); err == nil {
		t.Fatal("Calling AutoUnsubscribe() on closed subscription should fail")
	}
	delta := (runtime.NumGoroutine() - base)
	if delta > 0 {
		t.Fatalf("%d Go routines still exist post max subscriptions hit", delta)
	}
}

func TestClientSyncAutoUnsub(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()
	received := 0
	max := 10
	sub, _ := nc.SubscribeSync("foo")
	sub.AutoUnsubscribe(max)
	total := 100
	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	nc.Flush()
	for {
		_, err := sub.NextMsg(10 * time.Millisecond)
		if err != nil {
			if err != nats.ErrMaxMessages {
				t.Fatalf("Expected '%v', but got: '%v'\n", nats.ErrBadSubscription, err.Error())
			}
			break
		}
		received++
	}
	if received != max {
		t.Fatalf("Received %d msgs, wanted only %d\n", received, max)
	}
	if sub.IsValid() {
		t.Fatal("Expected subscription to be invalid after hitting max")
	}
	if err := sub.AutoUnsubscribe(10); err == nil {
		t.Fatal("Calling AutoUnsubscribe() ob closed subscription should fail")
	}
}

func TestClientASyncAutoUnsub(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()
	received := int32(0)
	max := int32(10)
	sub, err := nc.Subscribe("foo", func(_ *nats.Msg) {
		atomic.AddInt32(&received, 1)
	})
	if err != nil {
		t.Fatal("Failed to subscribe: ", err)
	}
	sub.AutoUnsubscribe(int(max))
	total := 100
	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	nc.Flush()
	time.Sleep(10 * time.Millisecond)

	if atomic.LoadInt32(&received) != max {
		t.Fatalf("Received %d msgs, wanted only %d\n", received, max)
	}
	if err := sub.AutoUnsubscribe(10); err == nil {
		t.Fatal("Calling AutoUnsubscribe() on closed subscription should fail")
	}
}

func TestAutoUnsubAndReconnect(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	rch := make(chan bool)

	nc, err := nats.Connect(nats.DefaultURL,
		nats.ReconnectWait(50*time.Millisecond),
		nats.ReconnectHandler(func(_ *nats.Conn) { rch <- true }))
	if err != nil {
		t.Fatalf("Unable to connect: %v", err)
	}
	defer nc.Close()

	received := int32(0)
	max := int32(10)
	sub, err := nc.Subscribe("foo", func(_ *nats.Msg) {
		atomic.AddInt32(&received, 1)
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	sub.AutoUnsubscribe(int(max))

	// Send less than the max
	total := int(max / 2)
	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	nc.Flush()

	// Restart the server
	s.Shutdown()
	s = RunDefaultServer()
	defer s.Shutdown()

	// and wait to reconnect
	if err := Wait(rch); err != nil {
		t.Fatal("Failed to get the reconnect cb")
	}

	// Now send more than the total max.
	total = int(3 * max)
	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	nc.Flush()

	// Wait a bit before checking.
	time.Sleep(50 * time.Millisecond)

	// We should have received only up-to-max messages.
	if atomic.LoadInt32(&received) != max {
		t.Fatalf("Received %d msgs, wanted only %d\n", received, max)
	}
}

func TestAutoUnsubWithParallelNextMsgCalls(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	rch := make(chan bool, 1)

	nc, err := nats.Connect(nats.DefaultURL,
		nats.ReconnectWait(50*time.Millisecond),
		nats.ReconnectHandler(func(_ *nats.Conn) { rch <- true }))
	if err != nil {
		t.Fatalf("Unable to connect: %v", err)
	}
	defer nc.Close()

	numRoutines := 3
	max := 100
	total := max * 2
	received := int64(0)

	var wg sync.WaitGroup

	sub, err := nc.SubscribeSync("foo")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	sub.AutoUnsubscribe(int(max))
	nc.Flush()

	wg.Add(numRoutines)

	for i := 0; i < numRoutines; i++ {
		go func(s *nats.Subscription, idx int) {
			for {
				// The first to reach the max delivered will cause the
				// subscription to be removed, which will kick out all
				// other calls to NextMsg. So don't be afraid of the long
				// timeout.
				_, err := s.NextMsg(3 * time.Second)
				if err != nil {
					break
				}
				atomic.AddInt64(&received, 1)
			}
			wg.Done()
		}(sub, i)
	}

	msg := []byte("Hello")
	for i := 0; i < max/2; i++ {
		nc.Publish("foo", msg)
	}
	nc.Flush()

	s.Shutdown()
	s = RunDefaultServer()
	defer s.Shutdown()

	// Make sure we got the reconnected cb
	if err := Wait(rch); err != nil {
		t.Fatal("Failed to get reconnected cb")
	}

	for i := 0; i < total; i++ {
		nc.Publish("foo", msg)
	}
	nc.Flush()

	wg.Wait()
	if atomic.LoadInt64(&received) != int64(max) {
		t.Fatalf("Wrong number of received msg: %v instead of %v", atomic.LoadInt64(&received), max)
	}
}

func TestAutoUnsubscribeFromCallback(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("Unable to connect: %v", err)
	}
	defer nc.Close()

	max := 10
	resetUnsubMark := int64(max / 2)
	limit := int64(100)
	received := int64(0)

	msg := []byte("Hello")

	// Auto-unsubscribe within the callback with a value lower
	// than what was already received.

	sub, err := nc.Subscribe("foo", func(m *nats.Msg) {
		r := atomic.AddInt64(&received, 1)
		if r == resetUnsubMark {
			m.Sub.AutoUnsubscribe(int(r - 1))
			nc.Flush()
		}
		if r == limit {
			// Something went wrong... fail now
			t.Fatal("Got more messages than expected")
		}
		nc.Publish("foo", msg)
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	sub.AutoUnsubscribe(int(max))
	nc.Flush()

	// Trigger the first message, the other are sent from the callback.
	nc.Publish("foo", msg)
	nc.Flush()

	time.Sleep(100 * time.Millisecond)

	recv := atomic.LoadInt64(&received)
	if recv != resetUnsubMark {
		t.Fatalf("Wrong number of received messages. Original max was %v reset to %v, actual received: %v",
			max, resetUnsubMark, recv)
	}

	// Now check with AutoUnsubscribe with higher value than original
	received = int64(0)
	newMax := int64(2 * max)

	sub, err = nc.Subscribe("foo", func(m *nats.Msg) {
		r := atomic.AddInt64(&received, 1)
		if r == resetUnsubMark {
			m.Sub.AutoUnsubscribe(int(newMax))
			nc.Flush()
		}
		if r == limit {
			// Something went wrong... fail now
			t.Fatal("Got more messages than expected")
		}
		nc.Publish("foo", msg)
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	sub.AutoUnsubscribe(int(max))
	nc.Flush()

	// Trigger the first message, the other are sent from the callback.
	nc.Publish("foo", msg)
	nc.Flush()

	time.Sleep(100 * time.Millisecond)

	recv = atomic.LoadInt64(&received)
	if recv != newMax {
		t.Fatalf("Wrong number of received messages. Original max was %v reset to %v, actual received: %v",
			max, newMax, recv)
	}
}

func TestCloseSubRelease(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	sub, _ := nc.SubscribeSync("foo")
	start := time.Now()
	go func() {
		time.Sleep(5 * time.Millisecond)
		nc.Close()
	}()
	_, err := sub.NextMsg(50 * time.Millisecond)
	if err == nil {
		t.Fatalf("Expected an error from NextMsg")
	}
	elapsed := time.Since(start)

	// On Windows, the minimum waitTime is at least 15ms.
	if elapsed > 20*time.Millisecond {
		t.Fatalf("Too much time has elapsed to release NextMsg: %dms",
			(elapsed / time.Millisecond))
	}
}

func TestIsValidSubscriber(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	sub, err := nc.SubscribeSync("foo")
	if err != nil {
		t.Fatalf("Error on subscribe: %v", err)
	}
	if !sub.IsValid() {
		t.Fatalf("Subscription should be valid")
	}
	for i := 0; i < 10; i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	nc.Flush()
	_, err = sub.NextMsg(200 * time.Millisecond)
	if err != nil {
		t.Fatalf("NextMsg returned an error")
	}
	sub.Unsubscribe()
	_, err = sub.NextMsg(200 * time.Millisecond)
	if err == nil {
		t.Fatalf("NextMsg should have returned an error")
	}
}

func TestSlowSubscriber(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	sub, _ := nc.SubscribeSync("foo")
	sub.SetPendingLimits(100, 1024)

	for i := 0; i < 200; i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	timeout := 5 * time.Second
	start := time.Now()
	nc.FlushTimeout(timeout)
	elapsed := time.Since(start)
	if elapsed >= timeout {
		t.Fatalf("Flush did not return before timeout: %d > %d", elapsed, timeout)
	}
	// Make sure NextMsg returns an error to indicate slow consumer
	_, err := sub.NextMsg(200 * time.Millisecond)
	if err == nil {
		t.Fatalf("NextMsg did not return an error")
	}
}

func TestSlowChanSubscriber(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	ch := make(chan *nats.Msg, 64)
	sub, _ := nc.ChanSubscribe("foo", ch)
	sub.SetPendingLimits(100, 1024)

	for i := 0; i < 200; i++ {
		nc.Publish("foo", []byte("Hello"))
	}
	timeout := 5 * time.Second
	start := time.Now()
	nc.FlushTimeout(timeout)
	elapsed := time.Since(start)
	if elapsed >= timeout {
		t.Fatalf("Flush did not return before timeout: %d > %d", elapsed, timeout)
	}
}

func TestSlowAsyncSubscriber(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	bch := make(chan bool)

	sub, _ := nc.Subscribe("foo", func(_ *nats.Msg) {
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
		nc.Publish("foo", []byte("Hello"))
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

func TestAsyncErrHandler(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	opts := nats.GetDefaultOptions()

	nc, err := opts.Connect()
	if err != nil {
		t.Fatalf("Could not connect to server: %v\n", err)
	}
	defer nc.Close()

	subj := "async_test"
	bch := make(chan bool)

	sub, err := nc.Subscribe(subj, func(_ *nats.Msg) {
		// block to back us up..
		<-bch
	})
	if err != nil {
		t.Fatalf("Could not subscribe: %v\n", err)
	}

	limit := 10
	toSend := 100

	// Limit internal subchan length to trip condition easier.
	sub.SetPendingLimits(limit, 1024)

	ch := make(chan bool)

	aeCalled := int64(0)

	nc.SetErrorHandler(func(c *nats.Conn, s *nats.Subscription, e error) {
		atomic.AddInt64(&aeCalled, 1)

		if s != sub {
			t.Fatal("Did not receive proper subscription")
		}
		if e != nats.ErrSlowConsumer {
			t.Fatalf("Did not receive proper error: %v vs %v\n", e, nats.ErrSlowConsumer)
		}
		// Suppress additional calls
		if atomic.LoadInt64(&aeCalled) == 1 {
			// release the sub
			defer close(bch)
			// release the test
			ch <- true
		}
	})

	b := []byte("Hello World!")
	// First one trips the ch wait in subscription callback.
	nc.Publish(subj, b)
	nc.Flush()
	for i := 0; i < toSend; i++ {
		nc.Publish(subj, b)
	}
	if err := nc.Flush(); err != nil {
		t.Fatalf("Got an error on Flush:%v\n", err)
	}

	if e := Wait(ch); e != nil {
		t.Fatal("Failed to call async err handler")
	}
	// Make sure dropped stats is correct.
	if d, _ := sub.Dropped(); d != toSend-limit {
		t.Fatalf("Expected Dropped to be %d, got %d\n", toSend-limit, d)
	}
	if ae := atomic.LoadInt64(&aeCalled); ae != 1 {
		t.Fatalf("Expected err handler to be called once, got %d\n", ae)
	}

	sub.Unsubscribe()
	if _, err := sub.Dropped(); err == nil {
		t.Fatal("Calling Dropped() on closed subscription should fail")
	}
}

func TestAsyncErrHandlerChanSubscription(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	opts := nats.GetDefaultOptions()

	nc, err := opts.Connect()
	if err != nil {
		t.Fatalf("Could not connect to server: %v\n", err)
	}
	defer nc.Close()

	subj := "chan_test"

	limit := 10
	toSend := 100

	// Create our own channel.
	mch := make(chan *nats.Msg, limit)
	sub, err := nc.ChanSubscribe(subj, mch)
	if err != nil {
		t.Fatalf("Could not subscribe: %v\n", err)
	}
	ch := make(chan bool)
	aeCalled := int64(0)

	nc.SetErrorHandler(func(c *nats.Conn, s *nats.Subscription, e error) {
		atomic.AddInt64(&aeCalled, 1)
		if e != nats.ErrSlowConsumer {
			t.Fatalf("Did not receive proper error: %v vs %v\n",
				e, nats.ErrSlowConsumer)
		}
		// Suppress additional calls
		if atomic.LoadInt64(&aeCalled) == 1 {
			// release the test
			ch <- true
		}
	})

	b := []byte("Hello World!")
	for i := 0; i < toSend; i++ {
		nc.Publish(subj, b)
	}
	nc.Flush()

	if e := Wait(ch); e != nil {
		t.Fatal("Failed to call async err handler")
	}
	// Make sure dropped stats is correct.
	if d, _ := sub.Dropped(); d != toSend-limit {
		t.Fatalf("Expected Dropped to be %d, go %d\n", toSend-limit, d)
	}
	if ae := atomic.LoadInt64(&aeCalled); ae != 1 {
		t.Fatalf("Expected err handler to be called once, got %d\n", ae)
	}

	sub.Unsubscribe()
	if _, err := sub.Dropped(); err == nil {
		t.Fatal("Calling Dropped() on closed subscription should fail")
	}
}

// Test to make sure that we can send and async receive messages on
// different subjects within a callback.
func TestAsyncSubscriberStarvation(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	// Helper
	nc.Subscribe("helper", func(m *nats.Msg) {
		nc.Publish(m.Reply, []byte("Hello"))
	})

	ch := make(chan bool)

	// Kickoff
	nc.Subscribe("start", func(m *nats.Msg) {
		// Helper Response
		response := nats.NewInbox()
		nc.Subscribe(response, func(_ *nats.Msg) {
			ch <- true
		})
		nc.PublishRequest("helper", response, []byte("Help Me!"))
	})

	nc.Publish("start", []byte("Begin"))
	nc.Flush()

	if e := Wait(ch); e != nil {
		t.Fatal("Was stalled inside of callback waiting on another callback")
	}
}

func TestAsyncSubscribersOnClose(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	toSend := 10
	callbacks := int32(0)
	ch := make(chan bool, toSend)

	nc.Subscribe("foo", func(_ *nats.Msg) {
		atomic.AddInt32(&callbacks, 1)
		<-ch
	})

	for i := 0; i < toSend; i++ {
		nc.Publish("foo", []byte("Hello World!"))
	}
	nc.Flush()
	time.Sleep(10 * time.Millisecond)
	nc.Close()

	// Release callbacks
	for i := 1; i < toSend; i++ {
		ch <- true
	}

	// Wait for some time.
	time.Sleep(10 * time.Millisecond)
	seen := atomic.LoadInt32(&callbacks)
	if seen != 1 {
		t.Fatalf("Expected only one callback, received %d callbacks\n", seen)
	}
}

func TestNextMsgCallOnAsyncSub(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()
	sub, err := nc.Subscribe("foo", func(_ *nats.Msg) {
	})
	if err != nil {
		t.Fatal("Failed to subscribe: ", err)
	}
	_, err = sub.NextMsg(time.Second)
	if err == nil {
		t.Fatal("Expected an error call NextMsg() on AsyncSubscriber")
	}
}

func TestNextMsgCallOnClosedSub(t *testing.T) {
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

	_, err = sub.NextMsg(time.Second)
	if err == nil {
		t.Fatal("Expected an error calling NextMsg() on closed subscription")
	} else if err != nats.ErrBadSubscription {
		t.Fatalf("Expected '%v', but got: '%v'\n", nats.ErrBadSubscription, err.Error())
	}
}

func TestChanSubscriber(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	// Create our own channel.
	ch := make(chan *nats.Msg, 128)

	// Channel is mandatory
	if _, err := nc.ChanSubscribe("foo", nil); err == nil {
		t.Fatal("Creating subscription without channel should have failed")
	}

	_, err := nc.ChanSubscribe("foo", ch)
	if err != nil {
		t.Fatal("Failed to subscribe: ", err)
	}

	// Send some messages to ourselves.
	total := 100
	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}

	received := 0
	tm := time.NewTimer(5 * time.Second)
	defer tm.Stop()

	// Go ahead and receive
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				t.Fatalf("Got an error reading from channel")
			}
		case <-tm.C:
			t.Fatalf("Timed out waiting on messages")
		}
		received++
		if received >= total {
			return
		}
	}
}

func TestChanQueueSubscriber(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	// Create our own channel.
	ch1 := make(chan *nats.Msg, 64)
	ch2 := make(chan *nats.Msg, 64)

	nc.ChanQueueSubscribe("foo", "bar", ch1)
	nc.ChanQueueSubscribe("foo", "bar", ch2)

	// Send some messages to ourselves.
	total := 100
	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}

	received := 0
	tm := time.NewTimer(5 * time.Second)
	defer tm.Stop()

	chk := func(ok bool) {
		if !ok {
			t.Fatalf("Got an error reading from channel")
		} else {
			received++
		}
	}

	// Go ahead and receive
	for {
		select {
		case _, ok := <-ch1:
			chk(ok)
		case _, ok := <-ch2:
			chk(ok)
		case <-tm.C:
			t.Fatalf("Timed out waiting on messages")
		}
		if received >= total {
			return
		}
	}
}

func TestChanSubscriberPendingLimits(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()
	ncp := NewDefaultConnection(t)
	defer ncp.Close()

	// There was a defect that prevented to receive more than
	// the default pending message limit. Trying to send more
	// than this limit.
	total := nats.DefaultSubPendingMsgsLimit + 100

	for typeSubs := 0; typeSubs < 3; typeSubs++ {

		func() {
			// Create our own channel.
			ch := make(chan *nats.Msg, total)

			var err error
			var sub *nats.Subscription
			switch typeSubs {
			case 0:
				sub, err = nc.ChanSubscribe("foo", ch)
			case 1:
				sub, err = nc.ChanQueueSubscribe("foo", "bar", ch)
			case 2:
				sub, err = nc.QueueSubscribeSyncWithChan("foo", "bar", ch)
			}
			if err != nil {
				t.Fatalf("Unexpected error on subscribe: %v", err)
			}
			defer sub.Unsubscribe()
			nc.Flush()

			// Send some messages
			go func() {
				for i := 0; i < total; i++ {
					if err := ncp.Publish("foo", []byte("Hello")); err != nil {
						t.Fatalf("Unexpected error on publish: %v", err)
					}
				}
			}()

			received := 0
			tm := time.NewTimer(5 * time.Second)
			defer tm.Stop()

			chk := func(ok bool) {
				if !ok {
					t.Fatalf("Got an error reading from channel")
				} else {
					received++
				}
			}

			// Go ahead and receive
			for {
				select {
				case _, ok := <-ch:
					chk(ok)
				case <-tm.C:
					t.Fatalf("Timed out waiting on messages")
				}
				if received >= total {
					return
				}
			}
		}()
	}
}

func TestQueueChanQueueSubscriber(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	// Create our own channel.
	ch1 := make(chan *nats.Msg, 64)
	ch2 := make(chan *nats.Msg, 64)

	nc.QueueSubscribeSyncWithChan("foo", "bar", ch1)
	nc.QueueSubscribeSyncWithChan("foo", "bar", ch2)

	// Send some messages to ourselves.
	total := 100
	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}

	recv1 := 0
	recv2 := 0
	tm := time.NewTimer(5 * time.Second)
	defer tm.Stop()
	runTimer := time.NewTimer(500 * time.Millisecond)
	defer runTimer.Stop()

	chk := func(ok bool, which int) {
		if !ok {
			t.Fatalf("Got an error reading from channel")
		} else {
			if which == 1 {
				recv1++
			} else {
				recv2++
			}
		}
	}

	// Go ahead and receive
recvLoop:
	for {
		select {
		case _, ok := <-ch1:
			chk(ok, 1)
		case _, ok := <-ch2:
			chk(ok, 2)
		case <-tm.C:
			t.Fatalf("Timed out waiting on messages")
		case <-runTimer.C:
			break recvLoop
		}
	}

	if recv1+recv2 > total {
		t.Fatalf("Received more messages than expected: %v vs %v", (recv1 + recv2), total)
	}
}

func TestUnsubscribeChanOnSubscriber(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	// Create our own channel.
	ch := make(chan *nats.Msg, 8)
	sub, _ := nc.ChanSubscribe("foo", ch)

	// Send some messages to ourselves.
	total := 100
	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}

	sub.Unsubscribe()
	for len(ch) > 0 {
		<-ch
	}
	// Make sure we can send to the channel still.
	// Test that we do not close it.
	ch <- &nats.Msg{}
}

func TestCloseChanOnSubscriber(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	// Create our own channel.
	ch := make(chan *nats.Msg, 8)
	nc.ChanSubscribe("foo", ch)

	// Send some messages to ourselves.
	total := 100
	for i := 0; i < total; i++ {
		nc.Publish("foo", []byte("Hello"))
	}

	nc.Close()
	for len(ch) > 0 {
		<-ch
	}
	// Make sure we can send to the channel still.
	// Test that we do not close it.
	ch <- &nats.Msg{}
}

func TestAsyncSubscriptionPending(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	// Send some messages to ourselves.
	total := 100
	msg := []byte("0123456789")

	inCb := make(chan bool)
	block := make(chan bool)

	sub, _ := nc.Subscribe("foo", func(_ *nats.Msg) {
		inCb <- true
		<-block
	})
	defer sub.Unsubscribe()

	for i := 0; i < total; i++ {
		nc.Publish("foo", msg)
	}
	nc.Flush()

	// Wait that a message is received, so checks are safe
	if err := Wait(inCb); err != nil {
		t.Fatal("No message received")
	}

	// Test old way
	q, _ := sub.QueuedMsgs()
	if q != total && q != total-1 {
		t.Fatalf("Expected %d or %d, got %d\n", total, total-1, q)
	}

	// New way, make sure the same and check bytes.
	m, b, _ := sub.Pending()
	mlen := len(msg)
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

func TestAsyncSubscriptionPendingDrain(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	// Send some messages to ourselves.
	total := 100
	msg := []byte("0123456789")

	sub, _ := nc.Subscribe("foo", func(_ *nats.Msg) {})
	defer sub.Unsubscribe()

	for i := 0; i < total; i++ {
		nc.Publish("foo", msg)
	}
	nc.Flush()

	// Wait for all delivered.
	for d, _ := sub.Delivered(); d != int64(total); d, _ = sub.Delivered() {
		time.Sleep(10 * time.Millisecond)
	}

	m, b, _ := sub.Pending()
	if m != 0 {
		t.Fatalf("Expected msgs of 0, got %d\n", m)
	}
	if b != 0 {
		t.Fatalf("Expected bytes of 0, got %d\n", b)
	}

	sub.Unsubscribe()
	if _, err := sub.Delivered(); err == nil {
		t.Fatal("Calling Delivered() on closed subscription should fail")
	}
}

func TestSyncSubscriptionPendingDrain(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	// Send some messages to ourselves.
	total := 100
	msg := []byte("0123456789")

	sub, _ := nc.SubscribeSync("foo")
	defer sub.Unsubscribe()

	for i := 0; i < total; i++ {
		nc.Publish("foo", msg)
	}
	nc.Flush()

	// Wait for all delivered.
	for d, _ := sub.Delivered(); d != int64(total); d, _ = sub.Delivered() {
		sub.NextMsg(10 * time.Millisecond)
	}

	m, b, _ := sub.Pending()
	if m != 0 {
		t.Fatalf("Expected msgs of 0, got %d\n", m)
	}
	if b != 0 {
		t.Fatalf("Expected bytes of 0, got %d\n", b)
	}

	sub.Unsubscribe()
	if _, err := sub.Delivered(); err == nil {
		t.Fatal("Calling Delivered() on closed subscription should fail")
	}
}

func TestSyncSubscriptionPending(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	sub, _ := nc.SubscribeSync("foo")
	defer sub.Unsubscribe()

	// Send some messages to ourselves.
	total := 100
	msg := []byte("0123456789")
	for i := 0; i < total; i++ {
		nc.Publish("foo", msg)
	}
	nc.Flush()

	// Test old way
	q, _ := sub.QueuedMsgs()
	if q != total && q != total-1 {
		t.Fatalf("Expected %d or %d, got %d\n", total, total-1, q)
	}

	// New way, make sure the same and check bytes.
	m, b, _ := sub.Pending()
	mlen := len(msg)

	if m != total {
		t.Fatalf("Expected msgs of %d, got %d\n", total, m)
	}
	if b != total*mlen {
		t.Fatalf("Expected bytes of %d, got %d\n", total*mlen, b)
	}

	// Now drain some down and make sure pending is correct
	for i := 0; i < total-1; i++ {
		sub.NextMsg(10 * time.Millisecond)
	}
	m, b, _ = sub.Pending()
	if m != 1 {
		t.Fatalf("Expected msgs of 1, got %d\n", m)
	}
	if b != mlen {
		t.Fatalf("Expected bytes of %d, got %d\n", mlen, b)
	}
}

func TestSetPendingLimits(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	payload := []byte("hello")
	payloadLen := len(payload)
	toSend := 100

	var sub *nats.Subscription

	// Check for invalid values
	invalid := func() error {
		if err := sub.SetPendingLimits(0, 1); err == nil {
			return fmt.Errorf("Setting limit with 0 should fail")
		}
		if err := sub.SetPendingLimits(1, 0); err == nil {
			return fmt.Errorf("Setting limit with 0 should fail")
		}
		return nil
	}
	// function to send messages
	send := func(subject string, count int) {
		for i := 0; i < count; i++ {
			if err := nc.Publish(subject, payload); err != nil {
				t.Fatalf("Unexpected error on publish: %v", err)
			}
		}
		nc.Flush()
	}

	// Check pending vs expected values
	var limitCount, limitBytes int
	var expectedCount, expectedBytes int
	checkPending := func() error {
		lc, lb, err := sub.PendingLimits()
		if err != nil {
			return err
		}
		if lc != limitCount || lb != limitBytes {
			return fmt.Errorf("Unexpected limits, expected %v msgs %v bytes, got %v msgs %v bytes",
				limitCount, limitBytes, lc, lb)
		}
		msgs, bytes, err := sub.Pending()
		if err != nil {
			return fmt.Errorf("Unexpected error getting pending counts: %v", err)
		}
		if (msgs != expectedCount && msgs != expectedCount-1) ||
			(bytes != expectedBytes && bytes != expectedBytes-payloadLen) {
			return fmt.Errorf("Unexpected counts, expected %v msgs %v bytes, got %v msgs %v bytes",
				expectedCount, expectedBytes, msgs, bytes)
		}
		return nil
	}

	recv := make(chan bool)
	block := make(chan bool)
	cb := func(m *nats.Msg) {
		recv <- true
		<-block
		m.Sub.Unsubscribe()
	}
	subj := "foo"
	sub, err := nc.Subscribe(subj, cb)
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	defer sub.Unsubscribe()
	if err := invalid(); err != nil {
		t.Fatalf("%v", err)
	}
	// Check we apply limit only for size
	limitCount = -1
	limitBytes = (toSend / 2) * payloadLen
	if err := sub.SetPendingLimits(limitCount, limitBytes); err != nil {
		t.Fatalf("Unexpected error setting limits: %v", err)
	}
	// Send messages
	send(subj, toSend)
	// Wait for message to be received
	if err := Wait(recv); err != nil {
		t.Fatal("Did not get our message")
	}
	expectedBytes = limitBytes
	expectedCount = limitBytes / payloadLen
	if err := checkPending(); err != nil {
		t.Fatalf("%v", err)
	}
	// Release callback
	block <- true

	subj = "bar"
	sub, err = nc.Subscribe(subj, cb)
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	defer sub.Unsubscribe()
	// Check we apply limit only for count
	limitCount = toSend / 4
	limitBytes = -1
	if err := sub.SetPendingLimits(limitCount, limitBytes); err != nil {
		t.Fatalf("Unexpected error setting limits: %v", err)
	}
	// Send messages
	send(subj, toSend)
	// Wait for message to be received
	if err := Wait(recv); err != nil {
		t.Fatal("Did not get our message")
	}
	expectedCount = limitCount
	expectedBytes = limitCount * payloadLen
	if err := checkPending(); err != nil {
		t.Fatalf("%v", err)
	}
	// Release callback
	block <- true

	subj = "baz"
	sub, err = nc.SubscribeSync(subj)
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	defer sub.Unsubscribe()
	if err := invalid(); err != nil {
		t.Fatalf("%v", err)
	}
	// Check we apply limit only for size
	limitCount = -1
	limitBytes = (toSend / 2) * payloadLen
	if err := sub.SetPendingLimits(limitCount, limitBytes); err != nil {
		t.Fatalf("Unexpected error setting limits: %v", err)
	}
	// Send messages
	send(subj, toSend)
	expectedBytes = limitBytes
	expectedCount = limitBytes / payloadLen
	if err := checkPending(); err != nil {
		t.Fatalf("%v", err)
	}
	sub.Unsubscribe()
	nc.Flush()

	subj = "boz"
	sub, err = nc.SubscribeSync(subj)
	if err != nil {
		t.Fatalf("Unexpected error on subscribe: %v", err)
	}
	defer sub.Unsubscribe()
	// Check we apply limit only for count
	limitCount = toSend / 4
	limitBytes = -1
	if err := sub.SetPendingLimits(limitCount, limitBytes); err != nil {
		t.Fatalf("Unexpected error setting limits: %v", err)
	}
	// Send messages
	send(subj, toSend)
	expectedCount = limitCount
	expectedBytes = limitCount * payloadLen
	if err := checkPending(); err != nil {
		t.Fatalf("%v", err)
	}
	sub.Unsubscribe()
	nc.Flush()
}

func TestSubscriptionTypes(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	nc := NewDefaultConnection(t)
	defer nc.Close()

	sub, _ := nc.Subscribe("foo", func(_ *nats.Msg) {})
	defer sub.Unsubscribe()
	if st := sub.Type(); st != nats.AsyncSubscription {
		t.Fatalf("Expected AsyncSubscription, got %v\n", st)
	}
	// Check Pending
	if err := sub.SetPendingLimits(1, 100); err != nil {
		t.Fatalf("We should be able to SetPendingLimits()")
	}
	if _, _, err := sub.Pending(); err != nil {
		t.Fatalf("We should be able to call Pending()")
	}
	sub.Unsubscribe()
	if err := sub.SetPendingLimits(1, 100); err == nil {
		t.Fatal("Calling SetPendingLimits() on closed subscription should fail")
	}
	if _, _, err := sub.PendingLimits(); err == nil {
		t.Fatal("Calling PendingLimits() on closed subscription should fail")
	}

	sub, _ = nc.SubscribeSync("foo")
	defer sub.Unsubscribe()
	if st := sub.Type(); st != nats.SyncSubscription {
		t.Fatalf("Expected SyncSubscription, got %v\n", st)
	}
	// Check Pending
	if err := sub.SetPendingLimits(1, 100); err != nil {
		t.Fatalf("We should be able to SetPendingLimits()")
	}
	if _, _, err := sub.Pending(); err != nil {
		t.Fatalf("We should be able to call Pending()")
	}
	sub.Unsubscribe()
	if err := sub.SetPendingLimits(1, 100); err == nil {
		t.Fatal("Calling SetPendingLimits() on closed subscription should fail")
	}
	if _, _, err := sub.PendingLimits(); err == nil {
		t.Fatal("Calling PendingLimits() on closed subscription should fail")
	}

	sub, _ = nc.ChanSubscribe("foo", make(chan *nats.Msg))
	defer sub.Unsubscribe()
	if st := sub.Type(); st != nats.ChanSubscription {
		t.Fatalf("Expected ChanSubscription, got %v\n", st)
	}
	// Check Pending
	if err := sub.SetPendingLimits(1, 100); err == nil {
		t.Fatalf("We should NOT be able to SetPendingLimits() on ChanSubscriber")
	}
	if _, _, err := sub.Pending(); err == nil {
		t.Fatalf("We should NOT be able to call Pending() on ChanSubscriber")
	}
	if _, _, err := sub.MaxPending(); err == nil {
		t.Fatalf("We should NOT be able to call MaxPending() on ChanSubscriber")
	}
	if err := sub.ClearMaxPending(); err == nil {
		t.Fatalf("We should NOT be able to call ClearMaxPending() on ChanSubscriber")
	}
	if _, _, err := sub.PendingLimits(); err == nil {
		t.Fatalf("We should NOT be able to call PendingLimits() on ChanSubscriber")
	}

}
