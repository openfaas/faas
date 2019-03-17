// Copyright (c) 2017 Uber Technologies, Inc.

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

package goleak

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ensure that testingT is a subset of testing.TB.
var _ = TestingT(testing.TB(nil))

// testOptions passes a shorter max sleep time, used so tests don't wait
// ~1 second in cases where we expect FindLeaks to error out.
func testOptions() Option {
	return maxSleep(time.Millisecond)
}

func TestFindLeaks(t *testing.T) {
	require.NoError(t, FindLeaks(), "Should find no leaks by default")

	bg := startBlockedG()
	err := FindLeaks(testOptions())
	require.Error(t, err, "Should find leaks with leaked goroutine")
	assert.Contains(t, err.Error(), "blockedG")
	assert.Contains(t, err.Error(), "created by go.uber.org/goleak.startBlockedG")

	// Once we unblock the goroutine, we shouldn't have leaks.
	bg.unblock()
	require.NoError(t, FindLeaks(), "Should find no leaks by default")
}

func TestFindLeaksRetry(t *testing.T) {
	// for i := 0; i < 10; i++ {
	bg := startBlockedG()
	require.Error(t, FindLeaks(testOptions()), "Should find leaks with leaked goroutine")

	go func() {
		time.Sleep(time.Millisecond)
		bg.unblock()
	}()
	require.NoError(t, FindLeaks(), "FindLeaks should retry while background goroutine ends")
}

type fakeT struct {
	errors []string
}

func (ft *fakeT) Error(args ...interface{}) {
	ft.errors = append(ft.errors, fmt.Sprint(args))
}

func TestVerifyNoLeaks(t *testing.T) {
	ft := &fakeT{}
	VerifyNoLeaks(ft)
	require.Empty(t, ft.errors, "Expect no errors from VerifyNoLeaks")

	bg := startBlockedG()
	VerifyNoLeaks(ft, testOptions())
	require.NotEmpty(t, ft.errors, "Expect errors from VerifyNoLeaks on leaked goroutine")
	bg.unblock()
}
