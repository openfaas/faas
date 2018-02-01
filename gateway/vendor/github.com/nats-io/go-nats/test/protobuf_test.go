package test

import (
	"reflect"
	"testing"
	"time"

	"github.com/nats-io/go-nats"

	"github.com/nats-io/go-nats/encoders/protobuf"
	pb "github.com/nats-io/go-nats/encoders/protobuf/testdata"
)

func NewProtoEncodedConn(tl TestLogger) *nats.EncodedConn {
	ec, err := nats.NewEncodedConn(NewConnection(tl, TEST_PORT), protobuf.PROTOBUF_ENCODER)
	if err != nil {
		tl.Fatalf("Failed to create an encoded connection: %v\n", err)
	}
	return ec
}

func TestEncProtoMarshalStruct(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewProtoEncodedConn(t)
	defer ec.Close()
	ch := make(chan bool)

	me := &pb.Person{Name: "derek", Age: 22, Address: "140 New Montgomery St"}
	me.Children = make(map[string]*pb.Person)

	me.Children["sam"] = &pb.Person{Name: "sam", Age: 19, Address: "140 New Montgomery St"}
	me.Children["meg"] = &pb.Person{Name: "meg", Age: 17, Address: "140 New Montgomery St"}

	ec.Subscribe("protobuf_test", func(p *pb.Person) {
		if !reflect.DeepEqual(p, me) {
			t.Fatal("Did not receive the correct protobuf response")
		}
		ch <- true
	})

	ec.Publish("protobuf_test", me)
	if e := Wait(ch); e != nil {
		t.Fatal("Did not receive the message")
	}
}

func TestEncProtoNilRequest(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewProtoEncodedConn(t)
	defer ec.Close()

	testPerson := &pb.Person{Name: "Anatolii", Age: 25, Address: "Ukraine, Nikolaev"}

	//Subscribe with empty interface shouldn't failed on empty message
	ec.Subscribe("nil_test", func(_, reply string, _ interface{}) {
		ec.Publish(reply, testPerson)
	})

	resp := new(pb.Person)

	//Request with nil argument shouldn't failed with nil argument
	err := ec.Request("nil_test", nil, resp, 100*time.Millisecond)
	ec.Flush()

	if err != nil {
		t.Error("Fail to send empty message via encoded proto connection")
	}

	if !reflect.DeepEqual(testPerson, resp) {
		t.Error("Fail to receive encoded response")
	}
}

func BenchmarkProtobufMarshalStruct(b *testing.B) {
	me := &pb.Person{Name: "derek", Age: 22, Address: "140 New Montgomery St"}
	me.Children = make(map[string]*pb.Person)

	me.Children["sam"] = &pb.Person{Name: "sam", Age: 19, Address: "140 New Montgomery St"}
	me.Children["meg"] = &pb.Person{Name: "meg", Age: 17, Address: "140 New Montgomery St"}

	encoder := &protobuf.ProtobufEncoder{}
	for n := 0; n < b.N; n++ {
		if _, err := encoder.Encode("protobuf_test", me); err != nil {
			b.Fatal("Couldn't serialize object", err)
		}
	}
}

func BenchmarkPublishProtobufStruct(b *testing.B) {
	// stop benchmark for set-up
	b.StopTimer()

	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	ec := NewProtoEncodedConn(b)
	defer ec.Close()
	ch := make(chan bool)

	me := &pb.Person{Name: "derek", Age: 22, Address: "140 New Montgomery St"}
	me.Children = make(map[string]*pb.Person)

	me.Children["sam"] = &pb.Person{Name: "sam", Age: 19, Address: "140 New Montgomery St"}
	me.Children["meg"] = &pb.Person{Name: "meg", Age: 17, Address: "140 New Montgomery St"}

	ec.Subscribe("protobuf_test", func(p *pb.Person) {
		if !reflect.DeepEqual(p, me) {
			b.Fatalf("Did not receive the correct protobuf response")
		}
		ch <- true
	})

	// resume benchmark
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		ec.Publish("protobuf_test", me)
		if e := Wait(ch); e != nil {
			b.Fatal("Did not receive the message")
		}
	}
}
