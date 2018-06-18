// Copyright 2017-2018 The NATS Authors
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

package nats

import (
	"testing"
	"time"
)

func TestTimerPool(t *testing.T) {
	var tp timerPool

	for i := 0; i < 10; i++ {
		tm := tp.Get(time.Millisecond * 20)

		select {
		case <-tm.C:
			t.Errorf("Timer already expired")
			continue
		default:
		}

		select {
		case <-tm.C:
		case <-time.After(time.Millisecond * 100):
			t.Errorf("Timer didn't expire in time")
		}

		tp.Put(tm)
	}
}
