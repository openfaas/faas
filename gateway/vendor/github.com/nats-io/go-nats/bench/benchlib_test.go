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

package bench

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/go-nats"
)

const (
	MsgSize = 8
	Million = 1000 * 1000
)

var baseTime = time.Now()

func millionMessagesSecondSample(seconds int) *Sample {
	messages := Million * seconds
	start := baseTime
	end := start.Add(time.Second * time.Duration(seconds))
	nc := new(nats.Conn)

	s := NewSample(messages, MsgSize, start, end, nc)
	s.MsgCnt = uint64(messages)
	s.MsgBytes = uint64(messages * MsgSize)
	s.IOBytes = s.MsgBytes
	return s
}

func TestDuration(t *testing.T) {
	s := millionMessagesSecondSample(1)
	duration := s.End.Sub(s.Start)
	if duration != s.Duration() || duration != time.Second {
		t.Fatal("Expected sample duration to be 1 second")
	}
}

func TestSeconds(t *testing.T) {
	s := millionMessagesSecondSample(1)
	seconds := s.End.Sub(s.Start).Seconds()
	if seconds != s.Seconds() || seconds != 1.0 {
		t.Fatal("Expected sample seconds to be 1 second")
	}
}

func TestRate(t *testing.T) {
	s := millionMessagesSecondSample(60)
	if s.Rate() != Million {
		t.Fatal("Expected rate at 1 million msgs")
	}
}

func TestThoughput(t *testing.T) {
	s := millionMessagesSecondSample(60)
	if s.Throughput() != Million*MsgSize {
		t.Fatalf("Expected throughput at %d million bytes/sec", MsgSize)
	}
}

func TestStrings(t *testing.T) {
	s := millionMessagesSecondSample(60)
	if len(s.String()) == 0 {
		t.Fatal("Sample didn't provide a String")
	}
}

func TestGroupDuration(t *testing.T) {
	sg := NewSampleGroup()
	sg.AddSample(millionMessagesSecondSample(1))
	sg.AddSample(millionMessagesSecondSample(2))
	duration := sg.End.Sub(sg.Start)
	if duration != sg.Duration() || duration != time.Duration(2)*time.Second {
		t.Fatal("Expected aggregate duration to be 2.0 seconds")
	}
}

func TestGroupSeconds(t *testing.T) {
	sg := NewSampleGroup()
	sg.AddSample(millionMessagesSecondSample(1))
	sg.AddSample(millionMessagesSecondSample(2))
	sg.AddSample(millionMessagesSecondSample(3))
	seconds := sg.End.Sub(sg.Start).Seconds()
	if seconds != sg.Seconds() || seconds != 3.0 {
		t.Fatal("Expected aggregate seconds to be 3.0 seconds")
	}
}

func TestGroupRate(t *testing.T) {
	sg := NewSampleGroup()
	sg.AddSample(millionMessagesSecondSample(1))
	sg.AddSample(millionMessagesSecondSample(2))
	sg.AddSample(millionMessagesSecondSample(3))
	if sg.Rate() != Million*2 {
		t.Fatal("Expected MsgRate at 2 million msg/sec")
	}
}

func TestGroupThoughput(t *testing.T) {
	sg := NewSampleGroup()
	sg.AddSample(millionMessagesSecondSample(1))
	sg.AddSample(millionMessagesSecondSample(2))
	sg.AddSample(millionMessagesSecondSample(3))
	if sg.Throughput() != 2*Million*MsgSize {
		t.Fatalf("Expected througput at %d million bytes/sec", 2*MsgSize)
	}
}

func TestMinMaxRate(t *testing.T) {
	sg := NewSampleGroup()
	sg.AddSample(millionMessagesSecondSample(1))
	sg.AddSample(millionMessagesSecondSample(2))
	sg.AddSample(millionMessagesSecondSample(3))
	if sg.MinRate() != sg.MaxRate() {
		t.Fatal("Expected MinRate == MaxRate")
	}
}

func TestAvgRate(t *testing.T) {
	sg := NewSampleGroup()
	sg.AddSample(millionMessagesSecondSample(1))
	sg.AddSample(millionMessagesSecondSample(2))
	sg.AddSample(millionMessagesSecondSample(3))
	if sg.MinRate() != sg.AvgRate() {
		t.Fatal("Expected MinRate == AvgRate")
	}
}

func TestStdDev(t *testing.T) {
	sg := NewSampleGroup()
	sg.AddSample(millionMessagesSecondSample(1))
	sg.AddSample(millionMessagesSecondSample(2))
	sg.AddSample(millionMessagesSecondSample(3))
	if sg.StdDev() != 0.0 {
		t.Fatal("Expected stddev to be zero")
	}
}

func TestBenchSetup(t *testing.T) {
	bench := NewBenchmark("test", 1, 1)
	bench.AddSubSample(millionMessagesSecondSample(1))
	bench.AddPubSample(millionMessagesSecondSample(1))
	bench.Close()
	if len(bench.RunID) == 0 {
		t.Fatal("Bench doesn't have a RunID")
	}
	if len(bench.Pubs.Samples) != 1 {
		t.Fatal("Expected one publisher")
	}
	if len(bench.Subs.Samples) != 1 {
		t.Fatal("Expected one subscriber")
	}
	if bench.MsgCnt != 2*Million {
		t.Fatal("Expected 2 million msgs")
	}
	if bench.IOBytes != 2*Million*MsgSize {
		t.Fatalf("Expected %d million bytes", 2*MsgSize)
	}
	if bench.Duration() != time.Second {
		t.Fatal("Expected duration to be 1 second")
	}
}

func makeBench(subs, pubs int) *Benchmark {
	bench := NewBenchmark("test", subs, pubs)
	for i := 0; i < subs; i++ {
		bench.AddSubSample(millionMessagesSecondSample(1))
	}
	for i := 0; i < pubs; i++ {
		bench.AddPubSample(millionMessagesSecondSample(1))
	}
	bench.Close()
	return bench
}

func TestCsv(t *testing.T) {
	bench := makeBench(1, 1)
	csv := bench.CSV()
	lines := strings.Split(csv, "\n")
	if len(lines) != 4 {
		t.Fatal("Expected 4 lines of output from the CSV string")
	}

	fields := strings.Split(lines[1], ",")
	if len(fields) != 7 {
		t.Fatal("Expected 7 fields")
	}
}

func TestBenchStrings(t *testing.T) {
	bench := makeBench(1, 1)
	s := bench.Report()
	lines := strings.Split(s, "\n")
	if len(lines) != 4 {
		t.Fatal("Expected 3 lines of output: header, pub, sub, empty")
	}

	bench = makeBench(2, 2)
	s = bench.Report()
	lines = strings.Split(s, "\n")
	if len(lines) != 10 {
		fmt.Printf("%q\n", s)

		t.Fatal("Expected 11 lines of output: header, pub header, pub x 2, stats, sub headers, sub x 2, stats, empty")
	}
}

func TestMsgsPerClient(t *testing.T) {
	zero := MsgsPerClient(0, 0)
	if len(zero) != 0 {
		t.Fatal("Expected 0 length for 0 clients")
	}
	onetwo := MsgsPerClient(1, 2)
	if len(onetwo) != 2 || onetwo[0] != 1 || onetwo[1] != 0 {
		t.Fatal("Expected uneven distribution")
	}
	twotwo := MsgsPerClient(2, 2)
	if len(twotwo) != 2 || twotwo[0] != 1 || twotwo[1] != 1 {
		t.Fatal("Expected even distribution")
	}
	threetwo := MsgsPerClient(3, 2)
	if len(threetwo) != 2 || threetwo[0] != 2 || threetwo[1] != 1 {
		t.Fatal("Expected uneven distribution")
	}
}
