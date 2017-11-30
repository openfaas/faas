//
// Blackfriday Markdown Processor
// Available at http://github.com/russross/blackfriday
//
// Copyright © 2011 Russ Ross <russ@russross.com>.
// Distributed under the Simplified BSD License.
// See README.md for details.
//

//
// Helper functions for unit testing
//

package blackfriday

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"testing"
)

type TestParams struct {
	extensions        Extensions
	referenceOverride ReferenceOverrideFunc
	HTMLFlags
	HTMLRendererParameters
}

func execRecoverableTestSuite(t *testing.T, tests []string, params TestParams, suite func(candidate *string)) {
	// Catch and report panics. This is useful when running 'go test -v' on
	// the integration server. When developing, though, crash dump is often
	// preferable, so recovery can be easily turned off with doRecover = false.
	var candidate string
	const doRecover = true
	if doRecover {
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("\npanic while processing [%#v]: %s\n", candidate, err)
			}
		}()
	}
	suite(&candidate)
}

func runMarkdown(input string, params TestParams) string {
	params.HTMLRendererParameters.Flags = params.HTMLFlags
	renderer := NewHTMLRenderer(params.HTMLRendererParameters)
	return string(Run([]byte(input), WithRenderer(renderer),
		WithExtensions(params.extensions),
		WithRefOverride(params.referenceOverride)))
}

// doTests runs full document tests using MarkdownCommon configuration.
func doTests(t *testing.T, tests []string) {
	doTestsParam(t, tests, TestParams{
		extensions: CommonExtensions,
		HTMLRendererParameters: HTMLRendererParameters{
			Flags: CommonHTMLFlags,
		},
	})
}

func doTestsBlock(t *testing.T, tests []string, extensions Extensions) {
	doTestsParam(t, tests, TestParams{
		extensions: extensions,
		HTMLFlags:  UseXHTML,
	})
}

func doTestsParam(t *testing.T, tests []string, params TestParams) {
	execRecoverableTestSuite(t, tests, params, func(candidate *string) {
		for i := 0; i+1 < len(tests); i += 2 {
			input := tests[i]
			*candidate = input
			expected := tests[i+1]
			actual := runMarkdown(*candidate, params)
			if actual != expected {
				t.Errorf("\nInput   [%#v]\nExpected[%#v]\nActual  [%#v]",
					*candidate, expected, actual)
			}

			// now test every substring to stress test bounds checking
			if !testing.Short() {
				for start := 0; start < len(input); start++ {
					for end := start + 1; end <= len(input); end++ {
						*candidate = input[start:end]
						runMarkdown(*candidate, params)
					}
				}
			}
		}
	})
}

func doTestsInline(t *testing.T, tests []string) {
	doTestsInlineParam(t, tests, TestParams{})
}

func doLinkTestsInline(t *testing.T, tests []string) {
	doTestsInline(t, tests)

	prefix := "http://localhost"
	params := HTMLRendererParameters{AbsolutePrefix: prefix}
	transformTests := transformLinks(tests, prefix)
	doTestsInlineParam(t, transformTests, TestParams{
		HTMLRendererParameters: params,
	})
	doTestsInlineParam(t, transformTests, TestParams{
		HTMLFlags:              UseXHTML,
		HTMLRendererParameters: params,
	})
}

func doSafeTestsInline(t *testing.T, tests []string) {
	doTestsInlineParam(t, tests, TestParams{HTMLFlags: Safelink})

	// All the links in this test should not have the prefix appended, so
	// just rerun it with different parameters and the same expectations.
	prefix := "http://localhost"
	params := HTMLRendererParameters{AbsolutePrefix: prefix}
	transformTests := transformLinks(tests, prefix)
	doTestsInlineParam(t, transformTests, TestParams{
		HTMLFlags:              Safelink,
		HTMLRendererParameters: params,
	})
}

func doTestsInlineParam(t *testing.T, tests []string, params TestParams) {
	params.extensions |= Autolink | Strikethrough
	params.HTMLFlags |= UseXHTML
	doTestsParam(t, tests, params)
}

func transformLinks(tests []string, prefix string) []string {
	newTests := make([]string, len(tests))
	anchorRe := regexp.MustCompile(`<a href="/(.*?)"`)
	imgRe := regexp.MustCompile(`<img src="/(.*?)"`)
	for i, test := range tests {
		if i%2 == 1 {
			test = anchorRe.ReplaceAllString(test, `<a href="`+prefix+`/$1"`)
			test = imgRe.ReplaceAllString(test, `<img src="`+prefix+`/$1"`)
		}
		newTests[i] = test
	}
	return newTests
}

func doTestsReference(t *testing.T, files []string, flag Extensions) {
	params := TestParams{extensions: flag}
	execRecoverableTestSuite(t, files, params, func(candidate *string) {
		for _, basename := range files {
			filename := filepath.Join("testdata", basename+".text")
			inputBytes, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Errorf("Couldn't open '%s', error: %v\n", filename, err)
				continue
			}
			input := string(inputBytes)

			filename = filepath.Join("testdata", basename+".html")
			expectedBytes, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Errorf("Couldn't open '%s', error: %v\n", filename, err)
				continue
			}
			expected := string(expectedBytes)

			actual := string(runMarkdown(input, params))
			if actual != expected {
				t.Errorf("\n    [%#v]\nExpected[%#v]\nActual  [%#v]",
					basename+".text", expected, actual)
			}

			// now test every prefix of every input to check for
			// bounds checking
			if !testing.Short() {
				start, max := 0, len(input)
				for end := start + 1; end <= max; end++ {
					*candidate = input[start:end]
					runMarkdown(*candidate, params)
				}
			}
		}
	})
}
