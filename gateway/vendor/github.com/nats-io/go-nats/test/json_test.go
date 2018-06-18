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
	"reflect"
	"testing"
	"time"

	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/builtin"
)

func NewJsonEncodedConn(tl TestLogger) *nats.EncodedConn {
	ec, err := nats.NewEncodedConn(NewConnection(tl, TEST_PORT), nats.JSON_ENCODER)
	if err != nil {
		tl.Fatalf("Failed to create an encoded connection: %v\n", err)
	}
	return ec
}

func TestEncBuiltinJsonMarshalString(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewJsonEncodedConn(t)
	defer ec.Close()
	ch := make(chan bool)

	testString := "Hello World!"

	ec.Subscribe("json_string", func(s string) {
		if s != testString {
			t.Fatalf("Received test string of '%s', wanted '%s'\n", s, testString)
		}
		ch <- true
	})
	ec.Publish("json_string", testString)
	if e := Wait(ch); e != nil {
		t.Fatal("Did not receive the message")
	}
}

func TestEncBuiltinJsonMarshalInt(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewJsonEncodedConn(t)
	defer ec.Close()
	ch := make(chan bool)

	testN := 22

	ec.Subscribe("json_int", func(n int) {
		if n != testN {
			t.Fatalf("Received test int of '%d', wanted '%d'\n", n, testN)
		}
		ch <- true
	})
	ec.Publish("json_int", testN)
	if e := Wait(ch); e != nil {
		t.Fatal("Did not receive the message")
	}
}

type person struct {
	Name     string
	Address  string
	Age      int
	Children map[string]*person
	Assets   map[string]uint
}

func TestEncBuiltinJsonMarshalStruct(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewJsonEncodedConn(t)
	defer ec.Close()
	ch := make(chan bool)

	me := &person{Name: "derek", Age: 22, Address: "140 New Montgomery St"}
	me.Children = make(map[string]*person)

	me.Children["sam"] = &person{Name: "sam", Age: 19, Address: "140 New Montgomery St"}
	me.Children["meg"] = &person{Name: "meg", Age: 17, Address: "140 New Montgomery St"}

	me.Assets = make(map[string]uint)
	me.Assets["house"] = 1000
	me.Assets["car"] = 100

	ec.Subscribe("json_struct", func(p *person) {
		if !reflect.DeepEqual(p, me) {
			t.Fatal("Did not receive the correct struct response")
		}
		ch <- true
	})

	ec.Publish("json_struct", me)
	if e := Wait(ch); e != nil {
		t.Fatal("Did not receive the message")
	}
}

func BenchmarkJsonMarshalStruct(b *testing.B) {
	me := &person{Name: "derek", Age: 22, Address: "140 New Montgomery St"}
	me.Children = make(map[string]*person)

	me.Children["sam"] = &person{Name: "sam", Age: 19, Address: "140 New Montgomery St"}
	me.Children["meg"] = &person{Name: "meg", Age: 17, Address: "140 New Montgomery St"}

	encoder := &builtin.JsonEncoder{}
	for n := 0; n < b.N; n++ {
		if _, err := encoder.Encode("protobuf_test", me); err != nil {
			b.Fatal("Couldn't serialize object", err)
		}
	}
}

func BenchmarkPublishJsonStruct(b *testing.B) {
	// stop benchmark for set-up
	b.StopTimer()

	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewJsonEncodedConn(b)
	defer ec.Close()
	ch := make(chan bool)

	me := &person{Name: "derek", Age: 22, Address: "140 New Montgomery St"}
	me.Children = make(map[string]*person)

	me.Children["sam"] = &person{Name: "sam", Age: 19, Address: "140 New Montgomery St"}
	me.Children["meg"] = &person{Name: "meg", Age: 17, Address: "140 New Montgomery St"}

	ec.Subscribe("json_struct", func(p *person) {
		if !reflect.DeepEqual(p, me) {
			b.Fatalf("Did not receive the correct struct response")
		}
		ch <- true
	})

	// resume benchmark
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		ec.Publish("json_struct", me)
		if e := Wait(ch); e != nil {
			b.Fatal("Did not receive the message")
		}
	}

}

func TestEncBuiltinNotMarshableToJson(t *testing.T) {
	je := &builtin.JsonEncoder{}
	ch := make(chan bool)
	_, err := je.Encode("foo", ch)
	if err == nil {
		t.Fatal("Expected an error when failing encoding")
	}
}

func TestEncBuiltinFailedEncodedPublish(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewJsonEncodedConn(t)
	defer ec.Close()

	ch := make(chan bool)
	err := ec.Publish("foo", ch)
	if err == nil {
		t.Fatal("Expected an error trying to publish a channel")
	}
	err = ec.PublishRequest("foo", "bar", ch)
	if err == nil {
		t.Fatal("Expected an error trying to publish a channel")
	}
	var cr chan bool
	err = ec.Request("foo", ch, &cr, 1*time.Second)
	if err == nil {
		t.Fatal("Expected an error trying to publish a channel")
	}
	err = ec.LastError()
	if err != nil {
		t.Fatalf("Expected LastError to be nil: %q ", err)
	}
}

func TestEncBuiltinDecodeConditionals(t *testing.T) {
	je := &builtin.JsonEncoder{}

	b, err := je.Encode("foo", 22)
	if err != nil {
		t.Fatalf("Expected no error when encoding, got %v\n", err)
	}
	var foo string
	var bar []byte
	err = je.Decode("foo", b, &foo)
	if err != nil {
		t.Fatalf("Expected no error when decoding, got %v\n", err)
	}
	err = je.Decode("foo", b, &bar)
	if err != nil {
		t.Fatalf("Expected no error when decoding, got %v\n", err)
	}
}
