// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package stack

import (
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _allDone chan struct{}

func waitForDone() {
	<-_allDone
}

func TestAll(t *testing.T) {
	// We use a global channel so that the function below does not
	// recieve any arguments, so we can test that parseFirstFunc works
	// regardless of arguments on the stack.
	_allDone = make(chan struct{})
	defer close(_allDone)

	for i := 0; i < 5; i++ {
		go waitForDone()
	}

	got := All()

	// We have exactly 7 gorotuines:
	// "main" goroutine
	// test goroutine
	// 5 goroutines started above.
	require.Len(t, got, 7)
	sort.Sort(byGoroutineID(got))

	assert.Contains(t, got[0].Full(), "testing.(*T).Run")
	assert.Contains(t, got[1].Full(), "TestAll")
	for i := 0; i < 5; i++ {
		assert.Contains(t, got[2+i].Full(), "stack.waitForDone")
	}
}

func TestCurrent(t *testing.T) {
	got := Current()
	assert.NotZero(t, got.ID(), "Should get non-zero goroutine id")
	assert.Equal(t, "running", got.State())
	assert.Equal(t, "go.uber.org/goleak/internal/stack.getStackBuffer", got.FirstFunction())

	wantFrames := []string{
		"stack.getStackBuffer",
		"stack.getStacks",
		"stack.Current",
		"stack.Current",
		"stack.TestCurrent",
	}
	all := got.Full()
	for _, frame := range wantFrames {
		assert.Contains(t, all, frame)
	}
	assert.Contains(t, got.String(), "in state")
	assert.Contains(t, got.String(), "on top of the stack")

	// Ensure that we are not returning the buffer without slicing it
	// from getStackBuffer.
	if len(got.Full()) > 1024 {
		t.Fatalf("Returned stack is too large")
	}
}

func TestAllLargeStack(t *testing.T) {
	const (
		stackDepth    = 100
		numGoroutines = 100
	)

	var started sync.WaitGroup

	done := make(chan struct{})
	for i := 0; i < numGoroutines; i++ {
		var f func(int)
		f = func(count int) {
			if count == 0 {
				started.Done()
				<-done
			}
			f(count - 1)
		}
		started.Add(1)
		go f(stackDepth)
	}

	started.Wait()
	buf := getStackBuffer(true /* all */)
	if len(buf) <= _defaultBufferSize {
		t.Fatalf("Expected larger stack buffer")
	}

	// Start enough goroutines so we exceed the default buffer size.
	close(done)
}

type byGoroutineID []Stack

func (ss byGoroutineID) Len() int           { return len(ss) }
func (ss byGoroutineID) Less(i, j int) bool { return ss[i].ID() < ss[j].ID() }
func (ss byGoroutineID) Swap(i, j int)      { ss[i], ss[j] = ss[j], ss[i] }
