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
	"runtime"
	"testing"
	"time"

	"github.com/nats-io/go-nats"
)

func TestBadChan(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	ec := NewEConn(t)
	defer ec.Close()

	if err := ec.BindSendChan("foo", "not a chan"); err == nil {
		t.Fatalf("Expected an Error when sending a non-channel\n")
	}

	if _, err := ec.BindRecvChan("foo", "not a chan"); err == nil {
		t.Fatalf("Expected an Error when sending a non-channel\n")
	}

	if err := ec.BindSendChan("foo", "not a chan"); err != nats.ErrChanArg {
		t.Fatalf("Expected an ErrChanArg when sending a non-channel\n")
	}

	if _, err := ec.BindRecvChan("foo", "not a chan"); err != nats.ErrChanArg {
		t.Fatalf("Expected an ErrChanArg when sending a non-channel\n")
	}
}

func TestSimpleSendChan(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	ec := NewEConn(t)
	defer ec.Close()

	recv := make(chan bool)

	numSent := int32(22)
	ch := make(chan int32)

	if err := ec.BindSendChan("foo", ch); err != nil {
		t.Fatalf("Failed to bind to a send channel: %v\n", err)
	}

	ec.Subscribe("foo", func(num int32) {
		if num != numSent {
			t.Fatalf("Failed to receive correct value: %d vs %d\n", num, numSent)
		}
		recv <- true
	})

	// Send to 'foo'
	ch <- numSent

	if e := Wait(recv); e != nil {
		if ec.LastError() != nil {
			e = ec.LastError()
		}
		t.Fatalf("Did not receive the message: %s", e)
	}
	close(ch)
}

func TestFailedChannelSend(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	ec := NewEConn(t)
	defer ec.Close()

	nc := ec.Conn
	ch := make(chan bool)
	wch := make(chan bool)

	nc.Opts.AsyncErrorCB = func(c *nats.Conn, s *nats.Subscription, e error) {
		wch <- true
	}

	if err := ec.BindSendChan("foo", ch); err != nil {
		t.Fatalf("Failed to bind to a receive channel: %v\n", err)
	}

	nc.Flush()

	go func() {
		time.Sleep(100 * time.Millisecond)
		nc.Close()
	}()

	func() {
		for {
			select {
			case ch <- true:
			case <-wch:
				return
			case <-time.After(time.Second):
				t.Fatal("Failed to get async error cb")
			}
		}
	}()

	ec = NewEConn(t)
	defer ec.Close()

	nc = ec.Conn
	bch := make(chan []byte)

	nc.Opts.AsyncErrorCB = func(c *nats.Conn, s *nats.Subscription, e error) {
		wch <- true
	}

	if err := ec.BindSendChan("foo", bch); err != nil {
		t.Fatalf("Failed to bind to a receive channel: %v\n", err)
	}

	buf := make([]byte, 2*1024*1024)
	bch <- buf

	if e := Wait(wch); e != nil {
		t.Fatal("Failed to call async err handler")
	}
}

func TestSimpleRecvChan(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	ec := NewEConn(t)
	defer ec.Close()

	numSent := int32(22)
	ch := make(chan int32)

	if _, err := ec.BindRecvChan("foo", ch); err != nil {
		t.Fatalf("Failed to bind to a receive channel: %v\n", err)
	}

	ec.Publish("foo", numSent)

	// Receive from 'foo'
	select {
	case num := <-ch:
		if num != numSent {
			t.Fatalf("Failed to receive correct value: %d vs %d\n", num, numSent)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("Failed to receive a value, timed-out\n")
	}
	close(ch)
}

func TestQueueRecvChan(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	ec := NewEConn(t)
	defer ec.Close()

	numSent := int32(22)
	ch := make(chan int32)

	if _, err := ec.BindRecvQueueChan("foo", "bar", ch); err != nil {
		t.Fatalf("Failed to bind to a queue receive channel: %v\n", err)
	}

	ec.Publish("foo", numSent)

	// Receive from 'foo'
	select {
	case num := <-ch:
		if num != numSent {
			t.Fatalf("Failed to receive correct value: %d vs %d\n", num, numSent)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("Failed to receive a value, timed-out\n")
	}
	close(ch)
}

func TestDecoderErrRecvChan(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	ec := NewEConn(t)
	defer ec.Close()
	nc := ec.Conn
	wch := make(chan bool)

	nc.Opts.AsyncErrorCB = func(c *nats.Conn, s *nats.Subscription, e error) {
		wch <- true
	}

	ch := make(chan *int32)

	if _, err := ec.BindRecvChan("foo", ch); err != nil {
		t.Fatalf("Failed to bind to a send channel: %v\n", err)
	}

	ec.Publish("foo", "Hello World")

	if e := Wait(wch); e != nil {
		t.Fatal("Failed to call async err handler")
	}
}

func TestRecvChanPanicOnClosedChan(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	ec := NewEConn(t)
	defer ec.Close()

	ch := make(chan int)

	if _, err := ec.BindRecvChan("foo", ch); err != nil {
		t.Fatalf("Failed to bind to a send channel: %v\n", err)
	}

	close(ch)
	ec.Publish("foo", 22)
	ec.Flush()
}

func TestRecvChanAsyncLeakGoRoutines(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	ec := NewEConn(t)
	defer ec.Close()

	// Call this to make sure that we have everything setup connection wise
	ec.Flush()

	before := runtime.NumGoroutine()

	ch := make(chan int)

	if _, err := ec.BindRecvChan("foo", ch); err != nil {
		t.Fatalf("Failed to bind to a send channel: %v\n", err)
	}

	// Close the receive Channel
	close(ch)

	// The publish will trigger the close and shutdown of the Go routines
	ec.Publish("foo", 22)
	ec.Flush()

	time.Sleep(100 * time.Millisecond)

	delta := (runtime.NumGoroutine() - before)

	if delta > 0 {
		t.Fatalf("Leaked Go routine(s) : %d, closing channel should have closed them\n", delta)
	}
}

func TestRecvChanLeakGoRoutines(t *testing.T) {
	s := RunDefaultServer()
	defer s.Shutdown()

	ec := NewEConn(t)
	defer ec.Close()

	// Call this to make sure that we have everything setup connection wise
	ec.Flush()

	before := runtime.NumGoroutine()

	ch := make(chan int)

	sub, err := ec.BindRecvChan("foo", ch)
	if err != nil {
		t.Fatalf("Failed to bind to a send channel: %v\n", err)
	}
	sub.Unsubscribe()

	// Sleep a bit to wait for the Go routine to exit.
	time.Sleep(500 * time.Millisecond)

	delta := (runtime.NumGoroutine() - before)

	if delta > 0 {
		t.Fatalf("Leaked Go routine(s) : %d, closing channel should have closed them\n", delta)
	}
}

func TestRecvChanMultipleMessages(t *testing.T) {
	// Make sure we can receive more than one message.
	// In response to #25, which is a bug from fixing #22.

	s := RunDefaultServer()
	defer s.Shutdown()

	ec := NewEConn(t)
	defer ec.Close()

	// Num to send, should == len of messages queued.
	size := 10

	ch := make(chan int, size)

	if _, err := ec.BindRecvChan("foo", ch); err != nil {
		t.Fatalf("Failed to bind to a send channel: %v\n", err)
	}

	for i := 0; i < size; i++ {
		ec.Publish("foo", 22)
	}
	ec.Flush()
	time.Sleep(10 * time.Millisecond)

	if lch := len(ch); lch != size {
		t.Fatalf("Expected %d messages queued, got %d.", size, lch)
	}
}

func BenchmarkPublishSpeedViaChan(b *testing.B) {
	b.StopTimer()

	s := RunDefaultServer()
	defer s.Shutdown()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		b.Fatalf("Could not connect: %v\n", err)
	}
	ec, err := nats.NewEncodedConn(nc, nats.DEFAULT_ENCODER)
	if err != nil {
		b.Fatalf("Failed creating encoded connection: %v\n", err)
	}
	defer ec.Close()

	ch := make(chan int32, 1024)
	if err := ec.BindSendChan("foo", ch); err != nil {
		b.Fatalf("Failed to bind to a send channel: %v\n", err)
	}

	b.StartTimer()

	num := int32(22)

	for i := 0; i < b.N; i++ {
		ch <- num
	}
	// Make sure they are all processed.
	nc.Flush()
	b.StopTimer()
}
