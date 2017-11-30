//
// Blackfriday Markdown Processor
// Available at http://github.com/russross/blackfriday
//
// Copyright © 2011 Russ Ross <russ@russross.com>.
// Distributed under the Simplified BSD License.
// See README.md for details.
//

//
// Markdown 1.0.3 reference tests
//

package blackfriday

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestReference(t *testing.T) {
	files := []string{
		"Amps and angle encoding",
		"Auto links",
		"Backslash escapes",
		"Blockquotes with code blocks",
		"Code Blocks",
		"Code Spans",
		"Hard-wrapped paragraphs with list-like lines",
		"Horizontal rules",
		"Inline HTML (Advanced)",
		"Inline HTML (Simple)",
		"Inline HTML comments",
		"Links, inline style",
		"Links, reference style",
		"Links, shortcut references",
		"Literal quotes in titles",
		"Markdown Documentation - Basics",
		"Markdown Documentation - Syntax",
		"Nested blockquotes",
		"Ordered and unordered lists",
		"Strong and em together",
		"Tabs",
		"Tidyness",
	}
	doTestsReference(t, files, 0)
}

func TestReference_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	files := []string{
		"Amps and angle encoding",
		"Auto links",
		"Backslash escapes",
		"Blockquotes with code blocks",
		"Code Blocks",
		"Code Spans",
		"Hard-wrapped paragraphs with list-like lines no empty line before block",
		"Horizontal rules",
		"Inline HTML (Advanced)",
		"Inline HTML (Simple)",
		"Inline HTML comments",
		"Links, inline style",
		"Links, reference style",
		"Links, shortcut references",
		"Literal quotes in titles",
		"Markdown Documentation - Basics",
		"Markdown Documentation - Syntax",
		"Nested blockquotes",
		"Ordered and unordered lists",
		"Strong and em together",
		"Tabs",
		"Tidyness",
	}
	doTestsReference(t, files, NoEmptyLineBeforeBlock)
}

// benchResultAnchor is an anchor variable to store the result of a benchmarked
// code so that compiler could never optimize away the call to runMarkdown()
var benchResultAnchor string

func BenchmarkReference(b *testing.B) {
	params := TestParams{extensions: CommonExtensions}
	files := []string{
		"Amps and angle encoding",
		"Auto links",
		"Backslash escapes",
		"Blockquotes with code blocks",
		"Code Blocks",
		"Code Spans",
		"Hard-wrapped paragraphs with list-like lines",
		"Horizontal rules",
		"Inline HTML (Advanced)",
		"Inline HTML (Simple)",
		"Inline HTML comments",
		"Links, inline style",
		"Links, reference style",
		"Links, shortcut references",
		"Literal quotes in titles",
		"Markdown Documentation - Basics",
		"Markdown Documentation - Syntax",
		"Nested blockquotes",
		"Ordered and unordered lists",
		"Strong and em together",
		"Tabs",
		"Tidyness",
	}
	var tests []string
	for _, basename := range files {
		filename := filepath.Join("testdata", basename+".text")
		inputBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			b.Errorf("Couldn't open '%s', error: %v\n", filename, err)
			continue
		}
		tests = append(tests, string(inputBytes))
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, test := range tests {
			benchResultAnchor = runMarkdown(test, params)
		}
	}
}
