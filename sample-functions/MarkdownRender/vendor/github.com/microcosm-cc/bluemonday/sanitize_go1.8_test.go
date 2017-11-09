// Copyright (c) 2014, David Kitchen <david@buro9.com>
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// * Neither the name of the organisation (Microcosm) nor the names of its
//   contributors may be used to endorse or promote products derived from
//   this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// +build go1.8

package bluemonday

import (
	"sync"
	"testing"
)

func TestXSSGo18(t *testing.T) {

	p := UGCPolicy()

	tests := []test{
		{
			in:       `<IMG SRC="jav&#x0D;ascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       "<IMG SRC=`javascript:alert(\"RSnake says, 'XSS'\")`>",
			expected: ``,
		},
	}

	// These tests are run concurrently to enable the race detector to pick up
	// potential issues
	wg := sync.WaitGroup{}
	wg.Add(len(tests))
	for ii, tt := range tests {
		go func(ii int, tt test) {
			out := p.Sanitize(tt.in)
			if out != tt.expected {
				t.Errorf(
					"test %d failed;\ninput   : %s\noutput  : %s\nexpected: %s",
					ii,
					tt.in,
					out,
					tt.expected,
				)
			}
			wg.Done()
		}(ii, tt)
	}
	wg.Wait()
}
