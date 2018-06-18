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
	"testing"
	"time"

	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/builtin"
)

const TEST_PORT = 8168

func NewDefaultEConn(t *testing.T) *nats.EncodedConn {
	ec, err := nats.NewEncodedConn(NewConnection(t, TEST_PORT), nats.DEFAULT_ENCODER)
	if err != nil {
		t.Fatalf("Failed to create an encoded connection: %v\n", err)
	}
	return ec
}

func TestEncBuiltinConstructorErrs(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	c := NewConnection(t, TEST_PORT)
	_, err := nats.NewEncodedConn(nil, "default")
	if err == nil {
		t.Fatal("Expected err for nil connection")
	}
	_, err = nats.NewEncodedConn(c, "foo22")
	if err == nil {
		t.Fatal("Expected err for bad encoder")
	}
	c.Close()
	_, err = nats.NewEncodedConn(c, "default")
	if err == nil {
		t.Fatal("Expected err for closed connection")
	}

}

func TestEncBuiltinMarshalString(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()
	ch := make(chan bool)

	testString := "Hello World!"

	ec.Subscribe("enc_string", func(s string) {
		if s != testString {
			t.Fatalf("Received test string of '%s', wanted '%s'\n", s, testString)
		}
		ch <- true
	})
	ec.Publish("enc_string", testString)
	if e := Wait(ch); e != nil {
		if ec.LastError() != nil {
			e = ec.LastError()
		}
		t.Fatalf("Did not receive the message: %s", e)
	}
}

func TestEncBuiltinMarshalBytes(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()
	ch := make(chan bool)

	testBytes := []byte("Hello World!")

	ec.Subscribe("enc_bytes", func(b []byte) {
		if !bytes.Equal(b, testBytes) {
			t.Fatalf("Received test bytes of '%s', wanted '%s'\n", b, testBytes)
		}
		ch <- true
	})
	ec.Publish("enc_bytes", testBytes)
	if e := Wait(ch); e != nil {
		if ec.LastError() != nil {
			e = ec.LastError()
		}
		t.Fatalf("Did not receive the message: %s", e)
	}
}

func TestEncBuiltinMarshalInt(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()
	ch := make(chan bool)

	testN := 22

	ec.Subscribe("enc_int", func(n int) {
		if n != testN {
			t.Fatalf("Received test number of %d, wanted %d\n", n, testN)
		}
		ch <- true
	})
	ec.Publish("enc_int", testN)
	if e := Wait(ch); e != nil {
		if ec.LastError() != nil {
			e = ec.LastError()
		}
		t.Fatalf("Did not receive the message: %s", e)
	}
}

func TestEncBuiltinMarshalInt32(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()
	ch := make(chan bool)

	testN := 22

	ec.Subscribe("enc_int", func(n int32) {
		if n != int32(testN) {
			t.Fatalf("Received test number of %d, wanted %d\n", n, testN)
		}
		ch <- true
	})
	ec.Publish("enc_int", testN)
	if e := Wait(ch); e != nil {
		if ec.LastError() != nil {
			e = ec.LastError()
		}
		t.Fatalf("Did not receive the message: %s", e)
	}
}

func TestEncBuiltinMarshalInt64(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()
	ch := make(chan bool)

	testN := 22

	ec.Subscribe("enc_int", func(n int64) {
		if n != int64(testN) {
			t.Fatalf("Received test number of %d, wanted %d\n", n, testN)
		}
		ch <- true
	})
	ec.Publish("enc_int", testN)
	if e := Wait(ch); e != nil {
		if ec.LastError() != nil {
			e = ec.LastError()
		}
		t.Fatalf("Did not receive the message: %s", e)
	}
}

func TestEncBuiltinMarshalFloat32(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()
	ch := make(chan bool)

	testN := float32(22)

	ec.Subscribe("enc_float", func(n float32) {
		if n != testN {
			t.Fatalf("Received test number of %f, wanted %f\n", n, testN)
		}
		ch <- true
	})
	ec.Publish("enc_float", testN)
	if e := Wait(ch); e != nil {
		if ec.LastError() != nil {
			e = ec.LastError()
		}
		t.Fatalf("Did not receive the message: %s", e)
	}
}

func TestEncBuiltinMarshalFloat64(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()
	ch := make(chan bool)

	testN := float64(22.22)

	ec.Subscribe("enc_float", func(n float64) {
		if n != testN {
			t.Fatalf("Received test number of %f, wanted %f\n", n, testN)
		}
		ch <- true
	})
	ec.Publish("enc_float", testN)
	if e := Wait(ch); e != nil {
		if ec.LastError() != nil {
			e = ec.LastError()
		}
		t.Fatalf("Did not receive the message: %s", e)
	}
}

func TestEncBuiltinMarshalBool(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()
	ch := make(chan bool)
	expected := make(chan bool, 1)

	ec.Subscribe("enc_bool", func(b bool) {
		val := <-expected
		if b != val {
			t.Fatal("Boolean values did not match")
		}
		ch <- true
	})

	expected <- false
	ec.Publish("enc_bool", false)
	if e := Wait(ch); e != nil {
		if ec.LastError() != nil {
			e = ec.LastError()
		}
		t.Fatalf("Did not receive the message: %s", e)
	}

	expected <- true
	ec.Publish("enc_bool", true)
	if e := Wait(ch); e != nil {
		if ec.LastError() != nil {
			e = ec.LastError()
		}
		t.Fatalf("Did not receive the message: %s", e)
	}
}

func TestEncBuiltinExtendedSubscribeCB(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()

	ch := make(chan bool)

	testString := "Hello World!"
	subject := "cb_args"

	ec.Subscribe(subject, func(subj, s string) {
		if s != testString {
			t.Fatalf("Received test string of '%s', wanted '%s'\n", s, testString)
		}
		if subj != subject {
			t.Fatalf("Received subject of '%s', wanted '%s'\n", subj, subject)
		}
		ch <- true
	})
	ec.Publish(subject, testString)
	if e := Wait(ch); e != nil {
		if ec.LastError() != nil {
			e = ec.LastError()
		}
		t.Fatalf("Did not receive the message: %s", e)
	}
}

func TestEncBuiltinExtendedSubscribeCB2(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()

	ch := make(chan bool)

	testString := "Hello World!"
	oSubj := "cb_args"
	oReply := "foobar"

	ec.Subscribe(oSubj, func(subj, reply, s string) {
		if s != testString {
			t.Fatalf("Received test string of '%s', wanted '%s'\n", s, testString)
		}
		if subj != oSubj {
			t.Fatalf("Received subject of '%s', wanted '%s'\n", subj, oSubj)
		}
		if reply != oReply {
			t.Fatalf("Received reply of '%s', wanted '%s'\n", reply, oReply)
		}
		ch <- true
	})
	ec.PublishRequest(oSubj, oReply, testString)
	if e := Wait(ch); e != nil {
		if ec.LastError() != nil {
			e = ec.LastError()
		}
		t.Fatalf("Did not receive the message: %s", e)
	}
}

func TestEncBuiltinRawMsgSubscribeCB(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()

	ch := make(chan bool)

	testString := "Hello World!"
	oSubj := "cb_args"
	oReply := "foobar"

	ec.Subscribe(oSubj, func(m *nats.Msg) {
		s := string(m.Data)
		if s != testString {
			t.Fatalf("Received test string of '%s', wanted '%s'\n", s, testString)
		}
		if m.Subject != oSubj {
			t.Fatalf("Received subject of '%s', wanted '%s'\n", m.Subject, oSubj)
		}
		if m.Reply != oReply {
			t.Fatalf("Received reply of '%s', wanted '%s'\n", m.Reply, oReply)
		}
		ch <- true
	})
	ec.PublishRequest(oSubj, oReply, testString)
	if e := Wait(ch); e != nil {
		if ec.LastError() != nil {
			e = ec.LastError()
		}
		t.Fatalf("Did not receive the message: %s", e)
	}
}

func TestEncBuiltinRequest(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()

	expectedResp := "I can help!"

	ec.Subscribe("help", func(subj, reply, req string) {
		ec.Publish(reply, expectedResp)
	})

	var resp string

	err := ec.Request("help", "help me", &resp, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed at receiving proper response: %v\n", err)
	}
	if resp != expectedResp {
		t.Fatalf("Received reply '%s', wanted '%s'\n", resp, expectedResp)
	}
}

func TestEncBuiltinRequestReceivesMsg(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()

	expectedResp := "I can help!"

	ec.Subscribe("help", func(subj, reply, req string) {
		ec.Publish(reply, expectedResp)
	})

	var resp nats.Msg

	err := ec.Request("help", "help me", &resp, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed at receiving proper response: %v\n", err)
	}
	if string(resp.Data) != expectedResp {
		t.Fatalf("Received reply '%s', wanted '%s'\n", string(resp.Data), expectedResp)
	}
}

func TestEncBuiltinAsyncMarshalErr(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewDefaultEConn(t)
	defer ec.Close()

	ch := make(chan bool)

	testString := "Hello World!"
	subject := "err_marshall"

	ec.Subscribe(subject, func(subj, num int) {
		// This will never get called.
	})

	ec.Conn.Opts.AsyncErrorCB = func(c *nats.Conn, s *nats.Subscription, err error) {
		ch <- true
	}

	ec.Publish(subject, testString)
	if e := Wait(ch); e != nil {
		t.Fatalf("Did not receive the message: %s", e)
	}
}

func TestEncBuiltinEncodeNil(t *testing.T) {
	de := &builtin.DefaultEncoder{}
	_, err := de.Encode("foo", nil)
	if err != nil {
		t.Fatalf("Expected no error encoding nil: %v", err)
	}
}

func TestEncBuiltinDecodeDefault(t *testing.T) {
	de := &builtin.DefaultEncoder{}
	b, err := de.Encode("foo", 22)
	if err != nil {
		t.Fatalf("Expected no error encoding number: %v", err)
	}
	var c chan bool
	err = de.Decode("foo", b, &c)
	if err == nil {
		t.Fatalf("Expected an error decoding")
	}
}
