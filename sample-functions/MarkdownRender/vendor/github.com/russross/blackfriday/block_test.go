//
// Blackfriday Markdown Processor
// Available at http://github.com/russross/blackfriday
//
// Copyright © 2011 Russ Ross <russ@russross.com>.
// Distributed under the Simplified BSD License.
// See README.md for details.
//

//
// Unit tests for block parsing
//

package blackfriday

import (
	"strings"
	"testing"
)

func TestPrefixHeaderNoExtensions(t *testing.T) {
	var tests = []string{
		"# Header 1\n",
		"<h1>Header 1</h1>\n",

		"## Header 2\n",
		"<h2>Header 2</h2>\n",

		"### Header 3\n",
		"<h3>Header 3</h3>\n",

		"#### Header 4\n",
		"<h4>Header 4</h4>\n",

		"##### Header 5\n",
		"<h5>Header 5</h5>\n",

		"###### Header 6\n",
		"<h6>Header 6</h6>\n",

		"####### Header 7\n",
		"<h6># Header 7</h6>\n",

		"#Header 1\n",
		"<h1>Header 1</h1>\n",

		"##Header 2\n",
		"<h2>Header 2</h2>\n",

		"###Header 3\n",
		"<h3>Header 3</h3>\n",

		"####Header 4\n",
		"<h4>Header 4</h4>\n",

		"#####Header 5\n",
		"<h5>Header 5</h5>\n",

		"######Header 6\n",
		"<h6>Header 6</h6>\n",

		"#######Header 7\n",
		"<h6>#Header 7</h6>\n",

		"Hello\n# Header 1\nGoodbye\n",
		"<p>Hello</p>\n\n<h1>Header 1</h1>\n\n<p>Goodbye</p>\n",

		"* List\n# Header\n* List\n",
		"<ul>\n<li><p>List</p>\n\n<h1>Header</h1></li>\n\n<li><p>List</p></li>\n</ul>\n",

		"* List\n#Header\n* List\n",
		"<ul>\n<li><p>List</p>\n\n<h1>Header</h1></li>\n\n<li><p>List</p></li>\n</ul>\n",

		"*   List\n    * Nested list\n    # Nested header\n",
		"<ul>\n<li><p>List</p>\n\n<ul>\n<li><p>Nested list</p>\n\n" +
			"<h1>Nested header</h1></li>\n</ul></li>\n</ul>\n",

		"#Header 1 \\#\n",
		"<h1>Header 1 #</h1>\n",

		"#Header 1 \\# foo\n",
		"<h1>Header 1 # foo</h1>\n",

		"#Header 1 #\\##\n",
		"<h1>Header 1 ##</h1>\n",
	}
	doTestsBlock(t, tests, 0)
}

func TestPrefixHeaderSpaceExtension(t *testing.T) {
	var tests = []string{
		"# Header 1\n",
		"<h1>Header 1</h1>\n",

		"## Header 2\n",
		"<h2>Header 2</h2>\n",

		"### Header 3\n",
		"<h3>Header 3</h3>\n",

		"#### Header 4\n",
		"<h4>Header 4</h4>\n",

		"##### Header 5\n",
		"<h5>Header 5</h5>\n",

		"###### Header 6\n",
		"<h6>Header 6</h6>\n",

		"####### Header 7\n",
		"<p>####### Header 7</p>\n",

		"#Header 1\n",
		"<p>#Header 1</p>\n",

		"##Header 2\n",
		"<p>##Header 2</p>\n",

		"###Header 3\n",
		"<p>###Header 3</p>\n",

		"####Header 4\n",
		"<p>####Header 4</p>\n",

		"#####Header 5\n",
		"<p>#####Header 5</p>\n",

		"######Header 6\n",
		"<p>######Header 6</p>\n",

		"#######Header 7\n",
		"<p>#######Header 7</p>\n",

		"Hello\n# Header 1\nGoodbye\n",
		"<p>Hello</p>\n\n<h1>Header 1</h1>\n\n<p>Goodbye</p>\n",

		"* List\n# Header\n* List\n",
		"<ul>\n<li><p>List</p>\n\n<h1>Header</h1></li>\n\n<li><p>List</p></li>\n</ul>\n",

		"* List\n#Header\n* List\n",
		"<ul>\n<li>List\n#Header</li>\n<li>List</li>\n</ul>\n",

		"*   List\n    * Nested list\n    # Nested header\n",
		"<ul>\n<li><p>List</p>\n\n<ul>\n<li><p>Nested list</p>\n\n" +
			"<h1>Nested header</h1></li>\n</ul></li>\n</ul>\n",
	}
	doTestsBlock(t, tests, SpaceHeadings)
}

func TestPrefixHeaderIdExtension(t *testing.T) {
	var tests = []string{
		"# Header 1 {#someid}\n",
		"<h1 id=\"someid\">Header 1</h1>\n",

		"# Header 1 {#someid}   \n",
		"<h1 id=\"someid\">Header 1</h1>\n",

		"# Header 1         {#someid}\n",
		"<h1 id=\"someid\">Header 1</h1>\n",

		"# Header 1 {#someid\n",
		"<h1>Header 1 {#someid</h1>\n",

		"# Header 1 {#someid\n",
		"<h1>Header 1 {#someid</h1>\n",

		"# Header 1 {#someid}}\n",
		"<h1 id=\"someid\">Header 1</h1>\n\n<p>}</p>\n",

		"## Header 2 {#someid}\n",
		"<h2 id=\"someid\">Header 2</h2>\n",

		"### Header 3 {#someid}\n",
		"<h3 id=\"someid\">Header 3</h3>\n",

		"#### Header 4 {#someid}\n",
		"<h4 id=\"someid\">Header 4</h4>\n",

		"##### Header 5 {#someid}\n",
		"<h5 id=\"someid\">Header 5</h5>\n",

		"###### Header 6 {#someid}\n",
		"<h6 id=\"someid\">Header 6</h6>\n",

		"####### Header 7 {#someid}\n",
		"<h6 id=\"someid\"># Header 7</h6>\n",

		"# Header 1 # {#someid}\n",
		"<h1 id=\"someid\">Header 1</h1>\n",

		"## Header 2 ## {#someid}\n",
		"<h2 id=\"someid\">Header 2</h2>\n",

		"Hello\n# Header 1\nGoodbye\n",
		"<p>Hello</p>\n\n<h1>Header 1</h1>\n\n<p>Goodbye</p>\n",

		"* List\n# Header {#someid}\n* List\n",
		"<ul>\n<li><p>List</p>\n\n<h1 id=\"someid\">Header</h1></li>\n\n<li><p>List</p></li>\n</ul>\n",

		"* List\n#Header {#someid}\n* List\n",
		"<ul>\n<li><p>List</p>\n\n<h1 id=\"someid\">Header</h1></li>\n\n<li><p>List</p></li>\n</ul>\n",

		"*   List\n    * Nested list\n    # Nested header {#someid}\n",
		"<ul>\n<li><p>List</p>\n\n<ul>\n<li><p>Nested list</p>\n\n" +
			"<h1 id=\"someid\">Nested header</h1></li>\n</ul></li>\n</ul>\n",
	}
	doTestsBlock(t, tests, HeadingIDs)
}

func TestPrefixHeaderIdExtensionWithPrefixAndSuffix(t *testing.T) {
	var tests = []string{
		"# header 1 {#someid}\n",
		"<h1 id=\"PRE:someid:POST\">header 1</h1>\n",

		"## header 2 {#someid}\n",
		"<h2 id=\"PRE:someid:POST\">header 2</h2>\n",

		"### header 3 {#someid}\n",
		"<h3 id=\"PRE:someid:POST\">header 3</h3>\n",

		"#### header 4 {#someid}\n",
		"<h4 id=\"PRE:someid:POST\">header 4</h4>\n",

		"##### header 5 {#someid}\n",
		"<h5 id=\"PRE:someid:POST\">header 5</h5>\n",

		"###### header 6 {#someid}\n",
		"<h6 id=\"PRE:someid:POST\">header 6</h6>\n",

		"####### header 7 {#someid}\n",
		"<h6 id=\"PRE:someid:POST\"># header 7</h6>\n",

		"# header 1 # {#someid}\n",
		"<h1 id=\"PRE:someid:POST\">header 1</h1>\n",

		"## header 2 ## {#someid}\n",
		"<h2 id=\"PRE:someid:POST\">header 2</h2>\n",

		"* List\n# Header {#someid}\n* List\n",
		"<ul>\n<li><p>List</p>\n\n<h1 id=\"PRE:someid:POST\">Header</h1></li>\n\n<li><p>List</p></li>\n</ul>\n",

		"* List\n#Header {#someid}\n* List\n",
		"<ul>\n<li><p>List</p>\n\n<h1 id=\"PRE:someid:POST\">Header</h1></li>\n\n<li><p>List</p></li>\n</ul>\n",

		"*   List\n    * Nested list\n    # Nested header {#someid}\n",
		"<ul>\n<li><p>List</p>\n\n<ul>\n<li><p>Nested list</p>\n\n" +
			"<h1 id=\"PRE:someid:POST\">Nested header</h1></li>\n</ul></li>\n</ul>\n",
	}

	parameters := HTMLRendererParameters{
		HeadingIDPrefix: "PRE:",
		HeadingIDSuffix: ":POST",
	}

	doTestsParam(t, tests, TestParams{
		extensions:             HeadingIDs,
		HTMLFlags:              UseXHTML,
		HTMLRendererParameters: parameters,
	})
}

func TestPrefixAutoHeaderIdExtension(t *testing.T) {
	var tests = []string{
		"# Header 1\n",
		"<h1 id=\"header-1\">Header 1</h1>\n",

		"# Header 1   \n",
		"<h1 id=\"header-1\">Header 1</h1>\n",

		"## Header 2\n",
		"<h2 id=\"header-2\">Header 2</h2>\n",

		"### Header 3\n",
		"<h3 id=\"header-3\">Header 3</h3>\n",

		"#### Header 4\n",
		"<h4 id=\"header-4\">Header 4</h4>\n",

		"##### Header 5\n",
		"<h5 id=\"header-5\">Header 5</h5>\n",

		"###### Header 6\n",
		"<h6 id=\"header-6\">Header 6</h6>\n",

		"####### Header 7\n",
		"<h6 id=\"header-7\"># Header 7</h6>\n",

		"Hello\n# Header 1\nGoodbye\n",
		"<p>Hello</p>\n\n<h1 id=\"header-1\">Header 1</h1>\n\n<p>Goodbye</p>\n",

		"* List\n# Header\n* List\n",
		"<ul>\n<li><p>List</p>\n\n<h1 id=\"header\">Header</h1></li>\n\n<li><p>List</p></li>\n</ul>\n",

		"* List\n#Header\n* List\n",
		"<ul>\n<li><p>List</p>\n\n<h1 id=\"header\">Header</h1></li>\n\n<li><p>List</p></li>\n</ul>\n",

		"*   List\n    * Nested list\n    # Nested header\n",
		"<ul>\n<li><p>List</p>\n\n<ul>\n<li><p>Nested list</p>\n\n" +
			"<h1 id=\"nested-header\">Nested header</h1></li>\n</ul></li>\n</ul>\n",

		"# Header\n\n# Header\n",
		"<h1 id=\"header\">Header</h1>\n\n<h1 id=\"header-1\">Header</h1>\n",

		"# Header 1\n\n# Header 1",
		"<h1 id=\"header-1\">Header 1</h1>\n\n<h1 id=\"header-1-1\">Header 1</h1>\n",

		"# Header\n\n# Header 1\n\n# Header\n\n# Header",
		"<h1 id=\"header\">Header</h1>\n\n<h1 id=\"header-1\">Header 1</h1>\n\n<h1 id=\"header-1-1\">Header</h1>\n\n<h1 id=\"header-1-2\">Header</h1>\n",
	}
	doTestsBlock(t, tests, AutoHeadingIDs)
}

func TestPrefixAutoHeaderIdExtensionWithPrefixAndSuffix(t *testing.T) {
	var tests = []string{
		"# Header 1\n",
		"<h1 id=\"PRE:header-1:POST\">Header 1</h1>\n",

		"# Header 1   \n",
		"<h1 id=\"PRE:header-1:POST\">Header 1</h1>\n",

		"## Header 2\n",
		"<h2 id=\"PRE:header-2:POST\">Header 2</h2>\n",

		"### Header 3\n",
		"<h3 id=\"PRE:header-3:POST\">Header 3</h3>\n",

		"#### Header 4\n",
		"<h4 id=\"PRE:header-4:POST\">Header 4</h4>\n",

		"##### Header 5\n",
		"<h5 id=\"PRE:header-5:POST\">Header 5</h5>\n",

		"###### Header 6\n",
		"<h6 id=\"PRE:header-6:POST\">Header 6</h6>\n",

		"####### Header 7\n",
		"<h6 id=\"PRE:header-7:POST\"># Header 7</h6>\n",

		"Hello\n# Header 1\nGoodbye\n",
		"<p>Hello</p>\n\n<h1 id=\"PRE:header-1:POST\">Header 1</h1>\n\n<p>Goodbye</p>\n",

		"* List\n# Header\n* List\n",
		"<ul>\n<li><p>List</p>\n\n<h1 id=\"PRE:header:POST\">Header</h1></li>\n\n<li><p>List</p></li>\n</ul>\n",

		"* List\n#Header\n* List\n",
		"<ul>\n<li><p>List</p>\n\n<h1 id=\"PRE:header:POST\">Header</h1></li>\n\n<li><p>List</p></li>\n</ul>\n",

		"*   List\n    * Nested list\n    # Nested header\n",
		"<ul>\n<li><p>List</p>\n\n<ul>\n<li><p>Nested list</p>\n\n" +
			"<h1 id=\"PRE:nested-header:POST\">Nested header</h1></li>\n</ul></li>\n</ul>\n",

		"# Header\n\n# Header\n",
		"<h1 id=\"PRE:header:POST\">Header</h1>\n\n<h1 id=\"PRE:header-1:POST\">Header</h1>\n",

		"# Header 1\n\n# Header 1",
		"<h1 id=\"PRE:header-1:POST\">Header 1</h1>\n\n<h1 id=\"PRE:header-1-1:POST\">Header 1</h1>\n",

		"# Header\n\n# Header 1\n\n# Header\n\n# Header",
		"<h1 id=\"PRE:header:POST\">Header</h1>\n\n<h1 id=\"PRE:header-1:POST\">Header 1</h1>\n\n<h1 id=\"PRE:header-1-1:POST\">Header</h1>\n\n<h1 id=\"PRE:header-1-2:POST\">Header</h1>\n",
	}

	parameters := HTMLRendererParameters{
		HeadingIDPrefix: "PRE:",
		HeadingIDSuffix: ":POST",
	}

	doTestsParam(t, tests, TestParams{
		extensions:             AutoHeadingIDs,
		HTMLFlags:              UseXHTML,
		HTMLRendererParameters: parameters,
	})
}

func TestPrefixMultipleHeaderExtensions(t *testing.T) {
	var tests = []string{
		"# Header\n\n# Header {#header}\n\n# Header 1",
		"<h1 id=\"header\">Header</h1>\n\n<h1 id=\"header-1\">Header</h1>\n\n<h1 id=\"header-1-1\">Header 1</h1>\n",
	}
	doTestsBlock(t, tests, AutoHeadingIDs|HeadingIDs)
}

func TestUnderlineHeaders(t *testing.T) {
	var tests = []string{
		"Header 1\n========\n",
		"<h1>Header 1</h1>\n",

		"Header 2\n--------\n",
		"<h2>Header 2</h2>\n",

		"A\n=\n",
		"<h1>A</h1>\n",

		"B\n-\n",
		"<h2>B</h2>\n",

		"Paragraph\nHeader\n=\n",
		"<p>Paragraph</p>\n\n<h1>Header</h1>\n",

		"Header\n===\nParagraph\n",
		"<h1>Header</h1>\n\n<p>Paragraph</p>\n",

		"Header\n===\nAnother header\n---\n",
		"<h1>Header</h1>\n\n<h2>Another header</h2>\n",

		"   Header\n======\n",
		"<h1>Header</h1>\n",

		"    Code\n========\n",
		"<pre><code>Code\n</code></pre>\n\n<p>========</p>\n",

		"Header with *inline*\n=====\n",
		"<h1>Header with <em>inline</em></h1>\n",

		"*   List\n    * Sublist\n    Not a header\n    ------\n",
		"<ul>\n<li>List\n\n<ul>\n<li>Sublist\nNot a header\n------</li>\n</ul></li>\n</ul>\n",

		"Paragraph\n\n\n\n\nHeader\n===\n",
		"<p>Paragraph</p>\n\n<h1>Header</h1>\n",

		"Trailing space \n====        \n\n",
		"<h1>Trailing space</h1>\n",

		"Trailing spaces\n====        \n\n",
		"<h1>Trailing spaces</h1>\n",

		"Double underline\n=====\n=====\n",
		"<h1>Double underline</h1>\n\n<p>=====</p>\n",
	}
	doTestsBlock(t, tests, 0)
}

func TestUnderlineHeadersAutoIDs(t *testing.T) {
	var tests = []string{
		"Header 1\n========\n",
		"<h1 id=\"header-1\">Header 1</h1>\n",

		"Header 2\n--------\n",
		"<h2 id=\"header-2\">Header 2</h2>\n",

		"A\n=\n",
		"<h1 id=\"a\">A</h1>\n",

		"B\n-\n",
		"<h2 id=\"b\">B</h2>\n",

		"Paragraph\nHeader\n=\n",
		"<p>Paragraph</p>\n\n<h1 id=\"header\">Header</h1>\n",

		"Header\n===\nParagraph\n",
		"<h1 id=\"header\">Header</h1>\n\n<p>Paragraph</p>\n",

		"Header\n===\nAnother header\n---\n",
		"<h1 id=\"header\">Header</h1>\n\n<h2 id=\"another-header\">Another header</h2>\n",

		"   Header\n======\n",
		"<h1 id=\"header\">Header</h1>\n",

		"Header with *inline*\n=====\n",
		"<h1 id=\"header-with-inline\">Header with <em>inline</em></h1>\n",

		"Paragraph\n\n\n\n\nHeader\n===\n",
		"<p>Paragraph</p>\n\n<h1 id=\"header\">Header</h1>\n",

		"Trailing space \n====        \n\n",
		"<h1 id=\"trailing-space\">Trailing space</h1>\n",

		"Trailing spaces\n====        \n\n",
		"<h1 id=\"trailing-spaces\">Trailing spaces</h1>\n",

		"Double underline\n=====\n=====\n",
		"<h1 id=\"double-underline\">Double underline</h1>\n\n<p>=====</p>\n",

		"Header\n======\n\nHeader\n======\n",
		"<h1 id=\"header\">Header</h1>\n\n<h1 id=\"header-1\">Header</h1>\n",

		"Header 1\n========\n\nHeader 1\n========\n",
		"<h1 id=\"header-1\">Header 1</h1>\n\n<h1 id=\"header-1-1\">Header 1</h1>\n",
	}
	doTestsBlock(t, tests, AutoHeadingIDs)
}

func TestHorizontalRule(t *testing.T) {
	var tests = []string{
		"-\n",
		"<p>-</p>\n",

		"--\n",
		"<p>--</p>\n",

		"---\n",
		"<hr />\n",

		"----\n",
		"<hr />\n",

		"*\n",
		"<p>*</p>\n",

		"**\n",
		"<p>**</p>\n",

		"***\n",
		"<hr />\n",

		"****\n",
		"<hr />\n",

		"_\n",
		"<p>_</p>\n",

		"__\n",
		"<p>__</p>\n",

		"___\n",
		"<hr />\n",

		"____\n",
		"<hr />\n",

		"-*-\n",
		"<p>-*-</p>\n",

		"- - -\n",
		"<hr />\n",

		"* * *\n",
		"<hr />\n",

		"_ _ _\n",
		"<hr />\n",

		"-----*\n",
		"<p>-----*</p>\n",

		"   ------   \n",
		"<hr />\n",

		"Hello\n***\n",
		"<p>Hello</p>\n\n<hr />\n",

		"---\n***\n___\n",
		"<hr />\n\n<hr />\n\n<hr />\n",
	}
	doTestsBlock(t, tests, 0)
}

func TestUnorderedList(t *testing.T) {
	var tests = []string{
		"* Hello\n",
		"<ul>\n<li>Hello</li>\n</ul>\n",

		"* Yin\n* Yang\n",
		"<ul>\n<li>Yin</li>\n<li>Yang</li>\n</ul>\n",

		"* Ting\n* Bong\n* Goo\n",
		"<ul>\n<li>Ting</li>\n<li>Bong</li>\n<li>Goo</li>\n</ul>\n",

		"* Yin\n\n* Yang\n",
		"<ul>\n<li><p>Yin</p></li>\n\n<li><p>Yang</p></li>\n</ul>\n",

		"* Ting\n\n* Bong\n* Goo\n",
		"<ul>\n<li><p>Ting</p></li>\n\n<li><p>Bong</p></li>\n\n<li><p>Goo</p></li>\n</ul>\n",

		"+ Hello\n",
		"<ul>\n<li>Hello</li>\n</ul>\n",

		"+ Yin\n+ Yang\n",
		"<ul>\n<li>Yin</li>\n<li>Yang</li>\n</ul>\n",

		"+ Ting\n+ Bong\n+ Goo\n",
		"<ul>\n<li>Ting</li>\n<li>Bong</li>\n<li>Goo</li>\n</ul>\n",

		"+ Yin\n\n+ Yang\n",
		"<ul>\n<li><p>Yin</p></li>\n\n<li><p>Yang</p></li>\n</ul>\n",

		"+ Ting\n\n+ Bong\n+ Goo\n",
		"<ul>\n<li><p>Ting</p></li>\n\n<li><p>Bong</p></li>\n\n<li><p>Goo</p></li>\n</ul>\n",

		"- Hello\n",
		"<ul>\n<li>Hello</li>\n</ul>\n",

		"- Yin\n- Yang\n",
		"<ul>\n<li>Yin</li>\n<li>Yang</li>\n</ul>\n",

		"- Ting\n- Bong\n- Goo\n",
		"<ul>\n<li>Ting</li>\n<li>Bong</li>\n<li>Goo</li>\n</ul>\n",

		"- Yin\n\n- Yang\n",
		"<ul>\n<li><p>Yin</p></li>\n\n<li><p>Yang</p></li>\n</ul>\n",

		"- Ting\n\n- Bong\n- Goo\n",
		"<ul>\n<li><p>Ting</p></li>\n\n<li><p>Bong</p></li>\n\n<li><p>Goo</p></li>\n</ul>\n",

		"*Hello\n",
		"<p>*Hello</p>\n",

		"*   Hello \n",
		"<ul>\n<li>Hello</li>\n</ul>\n",

		"*   Hello \n    Next line \n",
		"<ul>\n<li>Hello\nNext line</li>\n</ul>\n",

		"Paragraph\n* No linebreak\n",
		"<p>Paragraph\n* No linebreak</p>\n",

		"Paragraph\n\n* Linebreak\n",
		"<p>Paragraph</p>\n\n<ul>\n<li>Linebreak</li>\n</ul>\n",

		"*   List\n    * Nested list\n",
		"<ul>\n<li>List\n\n<ul>\n<li>Nested list</li>\n</ul></li>\n</ul>\n",

		"*   List\n\n    * Nested list\n",
		"<ul>\n<li><p>List</p>\n\n<ul>\n<li>Nested list</li>\n</ul></li>\n</ul>\n",

		"*   List\n    Second line\n\n    + Nested\n",
		"<ul>\n<li><p>List\nSecond line</p>\n\n<ul>\n<li>Nested</li>\n</ul></li>\n</ul>\n",

		"*   List\n    + Nested\n\n    Continued\n",
		"<ul>\n<li><p>List</p>\n\n<ul>\n<li>Nested</li>\n</ul>\n\n<p>Continued</p></li>\n</ul>\n",

		"*   List\n   * shallow indent\n",
		"<ul>\n<li>List\n\n<ul>\n<li>shallow indent</li>\n</ul></li>\n</ul>\n",

		"* List\n" +
			" * shallow indent\n" +
			"  * part of second list\n" +
			"   * still second\n" +
			"    * almost there\n" +
			"     * third level\n",
		"<ul>\n" +
			"<li>List\n\n" +
			"<ul>\n" +
			"<li>shallow indent</li>\n" +
			"<li>part of second list</li>\n" +
			"<li>still second</li>\n" +
			"<li>almost there\n\n" +
			"<ul>\n" +
			"<li>third level</li>\n" +
			"</ul></li>\n" +
			"</ul></li>\n" +
			"</ul>\n",

		"* List\n        extra indent, same paragraph\n",
		"<ul>\n<li>List\n    extra indent, same paragraph</li>\n</ul>\n",

		"* List\n\n        code block\n",
		"<ul>\n<li><p>List</p>\n\n<pre><code>code block\n</code></pre></li>\n</ul>\n",

		"* List\n\n          code block with spaces\n",
		"<ul>\n<li><p>List</p>\n\n<pre><code>  code block with spaces\n</code></pre></li>\n</ul>\n",

		"* List\n\n    * sublist\n\n    normal text\n\n    * another sublist\n",
		"<ul>\n<li><p>List</p>\n\n<ul>\n<li>sublist</li>\n</ul>\n\n<p>normal text</p>\n\n<ul>\n<li>another sublist</li>\n</ul></li>\n</ul>\n",
	}
	doTestsBlock(t, tests, 0)
}

func TestOrderedList(t *testing.T) {
	var tests = []string{
		"1. Hello\n",
		"<ol>\n<li>Hello</li>\n</ol>\n",

		"1. Yin\n2. Yang\n",
		"<ol>\n<li>Yin</li>\n<li>Yang</li>\n</ol>\n",

		"1. Ting\n2. Bong\n3. Goo\n",
		"<ol>\n<li>Ting</li>\n<li>Bong</li>\n<li>Goo</li>\n</ol>\n",

		"1. Yin\n\n2. Yang\n",
		"<ol>\n<li><p>Yin</p></li>\n\n<li><p>Yang</p></li>\n</ol>\n",

		"1. Ting\n\n2. Bong\n3. Goo\n",
		"<ol>\n<li><p>Ting</p></li>\n\n<li><p>Bong</p></li>\n\n<li><p>Goo</p></li>\n</ol>\n",

		"1 Hello\n",
		"<p>1 Hello</p>\n",

		"1.Hello\n",
		"<p>1.Hello</p>\n",

		"1.  Hello \n",
		"<ol>\n<li>Hello</li>\n</ol>\n",

		"1.  Hello \n    Next line \n",
		"<ol>\n<li>Hello\nNext line</li>\n</ol>\n",

		"Paragraph\n1. No linebreak\n",
		"<p>Paragraph\n1. No linebreak</p>\n",

		"Paragraph\n\n1. Linebreak\n",
		"<p>Paragraph</p>\n\n<ol>\n<li>Linebreak</li>\n</ol>\n",

		"1.  List\n    1. Nested list\n",
		"<ol>\n<li>List\n\n<ol>\n<li>Nested list</li>\n</ol></li>\n</ol>\n",

		"1.  List\n\n    1. Nested list\n",
		"<ol>\n<li><p>List</p>\n\n<ol>\n<li>Nested list</li>\n</ol></li>\n</ol>\n",

		"1.  List\n    Second line\n\n    1. Nested\n",
		"<ol>\n<li><p>List\nSecond line</p>\n\n<ol>\n<li>Nested</li>\n</ol></li>\n</ol>\n",

		"1.  List\n    1. Nested\n\n    Continued\n",
		"<ol>\n<li><p>List</p>\n\n<ol>\n<li>Nested</li>\n</ol>\n\n<p>Continued</p></li>\n</ol>\n",

		"1.  List\n   1. shallow indent\n",
		"<ol>\n<li>List\n\n<ol>\n<li>shallow indent</li>\n</ol></li>\n</ol>\n",

		"1. List\n" +
			" 1. shallow indent\n" +
			"  2. part of second list\n" +
			"   3. still second\n" +
			"    4. almost there\n" +
			"     1. third level\n",
		"<ol>\n" +
			"<li>List\n\n" +
			"<ol>\n" +
			"<li>shallow indent</li>\n" +
			"<li>part of second list</li>\n" +
			"<li>still second</li>\n" +
			"<li>almost there\n\n" +
			"<ol>\n" +
			"<li>third level</li>\n" +
			"</ol></li>\n" +
			"</ol></li>\n" +
			"</ol>\n",

		"1. List\n        extra indent, same paragraph\n",
		"<ol>\n<li>List\n    extra indent, same paragraph</li>\n</ol>\n",

		"1. List\n\n        code block\n",
		"<ol>\n<li><p>List</p>\n\n<pre><code>code block\n</code></pre></li>\n</ol>\n",

		"1. List\n\n          code block with spaces\n",
		"<ol>\n<li><p>List</p>\n\n<pre><code>  code block with spaces\n</code></pre></li>\n</ol>\n",

		"1. List\n    * Mixted list\n",
		"<ol>\n<li>List\n\n<ul>\n<li>Mixted list</li>\n</ul></li>\n</ol>\n",

		"1. List\n * Mixed list\n",
		"<ol>\n<li>List\n\n<ul>\n<li>Mixed list</li>\n</ul></li>\n</ol>\n",

		"* Start with unordered\n 1. Ordered\n",
		"<ul>\n<li>Start with unordered\n\n<ol>\n<li>Ordered</li>\n</ol></li>\n</ul>\n",

		"* Start with unordered\n    1. Ordered\n",
		"<ul>\n<li>Start with unordered\n\n<ol>\n<li>Ordered</li>\n</ol></li>\n</ul>\n",

		"1. numbers\n1. are ignored\n",
		"<ol>\n<li>numbers</li>\n<li>are ignored</li>\n</ol>\n",
	}
	doTestsBlock(t, tests, 0)
}

func TestDefinitionList(t *testing.T) {
	var tests = []string{
		"Term 1\n:   Definition a\n",
		"<dl>\n<dt>Term 1</dt>\n<dd>Definition a</dd>\n</dl>\n",

		"Term 1\n:   Definition a \n",
		"<dl>\n<dt>Term 1</dt>\n<dd>Definition a</dd>\n</dl>\n",

		"Term 1\n:   Definition a\n:   Definition b\n",
		"<dl>\n<dt>Term 1</dt>\n<dd>Definition a</dd>\n<dd>Definition b</dd>\n</dl>\n",

		"Term 1\n:   Definition a\n\nTerm 2\n:   Definition b\n",
		"<dl>\n" +
			"<dt>Term 1</dt>\n" +
			"<dd>Definition a</dd>\n" +
			"<dt>Term 2</dt>\n" +
			"<dd>Definition b</dd>\n" +
			"</dl>\n",

		"Term 1\n:   Definition a\n\nTerm 2\n:   Definition b\n\nTerm 3\n:   Definition c\n",
		"<dl>\n" +
			"<dt>Term 1</dt>\n" +
			"<dd>Definition a</dd>\n" +
			"<dt>Term 2</dt>\n" +
			"<dd>Definition b</dd>\n" +
			"<dt>Term 3</dt>\n" +
			"<dd>Definition c</dd>\n" +
			"</dl>\n",

		"Term 1\n:   Definition a\n:   Definition b\n\nTerm 2\n:   Definition c\n",
		"<dl>\n" +
			"<dt>Term 1</dt>\n" +
			"<dd>Definition a</dd>\n" +
			"<dd>Definition b</dd>\n" +
			"<dt>Term 2</dt>\n" +
			"<dd>Definition c</dd>\n" +
			"</dl>\n",

		"Term 1\n\n:   Definition a\n\nTerm 2\n\n:   Definition b\n",
		"<dl>\n" +
			"<dt>Term 1</dt>\n" +
			"<dd><p>Definition a</p></dd>\n" +
			"<dt>Term 2</dt>\n" +
			"<dd><p>Definition b</p></dd>\n" +
			"</dl>\n",

		"Term 1\n\n:   Definition a\n\n:   Definition b\n\nTerm 2\n\n:   Definition c\n",
		"<dl>\n" +
			"<dt>Term 1</dt>\n" +
			"<dd><p>Definition a</p></dd>\n" +
			"<dd><p>Definition b</p></dd>\n" +
			"<dt>Term 2</dt>\n" +
			"<dd><p>Definition c</p></dd>\n" +
			"</dl>\n",

		"Term 1\n:   Definition a\nNext line\n",
		"<dl>\n<dt>Term 1</dt>\n<dd>Definition a\nNext line</dd>\n</dl>\n",

		"Term 1\n:   Definition a\n  Next line\n",
		"<dl>\n<dt>Term 1</dt>\n<dd>Definition a\nNext line</dd>\n</dl>\n",

		"Term 1\n:   Definition a \n  Next line \n",
		"<dl>\n<dt>Term 1</dt>\n<dd>Definition a\nNext line</dd>\n</dl>\n",

		"Term 1\n:   Definition a\nNext line\n\nTerm 2\n:   Definition b",
		"<dl>\n" +
			"<dt>Term 1</dt>\n" +
			"<dd>Definition a\nNext line</dd>\n" +
			"<dt>Term 2</dt>\n" +
			"<dd>Definition b</dd>\n" +
			"</dl>\n",

		"Term 1\n: Definition a\n",
		"<dl>\n<dt>Term 1</dt>\n<dd>Definition a</dd>\n</dl>\n",

		"Term 1\n:Definition a\n",
		"<p>Term 1\n:Definition a</p>\n",

		"Term 1\n\n:   Definition a\n\nTerm 2\n\n:   Definition b\n\nText 1",
		"<dl>\n" +
			"<dt>Term 1</dt>\n" +
			"<dd><p>Definition a</p></dd>\n" +
			"<dt>Term 2</dt>\n" +
			"<dd><p>Definition b</p></dd>\n" +
			"</dl>\n" +
			"\n<p>Text 1</p>\n",

		"Term 1\n\n:   Definition a\n\nText 1\n\nTerm 2\n\n:   Definition b\n\nText 2",
		"<dl>\n" +
			"<dt>Term 1</dt>\n" +
			"<dd><p>Definition a</p></dd>\n" +
			"</dl>\n" +
			"\n<p>Text 1</p>\n" +
			"\n<dl>\n" +
			"<dt>Term 2</dt>\n" +
			"<dd><p>Definition b</p></dd>\n" +
			"</dl>\n" +
			"\n<p>Text 2</p>\n",
	}
	doTestsBlock(t, tests, DefinitionLists)
}

func TestPreformattedHtml(t *testing.T) {
	var tests = []string{
		"<div></div>\n",
		"<div></div>\n",

		"<div>\n</div>\n",
		"<div>\n</div>\n",

		"<div>\n</div>\nParagraph\n",
		"<p><div>\n</div>\nParagraph</p>\n",

		"<div class=\"foo\">\n</div>\n",
		"<div class=\"foo\">\n</div>\n",

		"<div>\nAnything here\n</div>\n",
		"<div>\nAnything here\n</div>\n",

		"<div>\n  Anything here\n</div>\n",
		"<div>\n  Anything here\n</div>\n",

		"<div>\nAnything here\n  </div>\n",
		"<div>\nAnything here\n  </div>\n",

		"<div>\nThis is *not* &proceessed\n</div>\n",
		"<div>\nThis is *not* &proceessed\n</div>\n",

		"<faketag>\n  Something\n</faketag>\n",
		"<p><faketag>\n  Something\n</faketag></p>\n",

		"<div>\n  Something here\n</divv>\n",
		"<p><div>\n  Something here\n</divv></p>\n",

		"Paragraph\n<div>\nHere? >&<\n</div>\n",
		"<p>Paragraph\n<div>\nHere? &gt;&amp;&lt;\n</div></p>\n",

		"Paragraph\n\n<div>\nHow about here? >&<\n</div>\n",
		"<p>Paragraph</p>\n\n<div>\nHow about here? >&<\n</div>\n",

		"Paragraph\n<div>\nHere? >&<\n</div>\nAnd here?\n",
		"<p>Paragraph\n<div>\nHere? &gt;&amp;&lt;\n</div>\nAnd here?</p>\n",

		"Paragraph\n\n<div>\nHow about here? >&<\n</div>\nAnd here?\n",
		"<p>Paragraph</p>\n\n<p><div>\nHow about here? &gt;&amp;&lt;\n</div>\nAnd here?</p>\n",

		"Paragraph\n<div>\nHere? >&<\n</div>\n\nAnd here?\n",
		"<p>Paragraph\n<div>\nHere? &gt;&amp;&lt;\n</div></p>\n\n<p>And here?</p>\n",

		"Paragraph\n\n<div>\nHow about here? >&<\n</div>\n\nAnd here?\n",
		"<p>Paragraph</p>\n\n<div>\nHow about here? >&<\n</div>\n\n<p>And here?</p>\n",
	}
	doTestsBlock(t, tests, 0)
}

func TestPreformattedHtmlLax(t *testing.T) {
	var tests = []string{
		"Paragraph\n<div>\nHere? >&<\n</div>\n",
		"<p>Paragraph</p>\n\n<div>\nHere? >&<\n</div>\n",

		"Paragraph\n\n<div>\nHow about here? >&<\n</div>\n",
		"<p>Paragraph</p>\n\n<div>\nHow about here? >&<\n</div>\n",

		"Paragraph\n<div>\nHere? >&<\n</div>\nAnd here?\n",
		"<p>Paragraph</p>\n\n<div>\nHere? >&<\n</div>\n\n<p>And here?</p>\n",

		"Paragraph\n\n<div>\nHow about here? >&<\n</div>\nAnd here?\n",
		"<p>Paragraph</p>\n\n<div>\nHow about here? >&<\n</div>\n\n<p>And here?</p>\n",

		"Paragraph\n<div>\nHere? >&<\n</div>\n\nAnd here?\n",
		"<p>Paragraph</p>\n\n<div>\nHere? >&<\n</div>\n\n<p>And here?</p>\n",

		"Paragraph\n\n<div>\nHow about here? >&<\n</div>\n\nAnd here?\n",
		"<p>Paragraph</p>\n\n<div>\nHow about here? >&<\n</div>\n\n<p>And here?</p>\n",
	}
	doTestsBlock(t, tests, LaxHTMLBlocks)
}

func TestFencedCodeBlock(t *testing.T) {
	var tests = []string{
		"``` go\nfunc foo() bool {\n\treturn true;\n}\n```\n",
		"<pre><code class=\"language-go\">func foo() bool {\n\treturn true;\n}\n</code></pre>\n",

		"``` c\n/* special & char < > \" escaping */\n```\n",
		"<pre><code class=\"language-c\">/* special &amp; char &lt; &gt; &quot; escaping */\n</code></pre>\n",

		"``` c\nno *inline* processing ~~of text~~\n```\n",
		"<pre><code class=\"language-c\">no *inline* processing ~~of text~~\n</code></pre>\n",

		"```\nNo language\n```\n",
		"<pre><code>No language\n</code></pre>\n",

		"``` {ocaml}\nlanguage in braces\n```\n",
		"<pre><code class=\"language-ocaml\">language in braces\n</code></pre>\n",

		"```    {ocaml}      \nwith extra whitespace\n```\n",
		"<pre><code class=\"language-ocaml\">with extra whitespace\n</code></pre>\n",

		"```{   ocaml   }\nwith extra whitespace\n```\n",
		"<pre><code class=\"language-ocaml\">with extra whitespace\n</code></pre>\n",

		"~ ~~ java\nWith whitespace\n~~~\n",
		"<p>~ ~~ java\nWith whitespace\n~~~</p>\n",

		"~~\nonly two\n~~\n",
		"<p>~~\nonly two\n~~</p>\n",

		"```` python\nextra\n````\n",
		"<pre><code class=\"language-python\">extra\n</code></pre>\n",

		"~~~ perl\nthree to start, four to end\n~~~~\n",
		"<p>~~~ perl\nthree to start, four to end\n~~~~</p>\n",

		"~~~~ perl\nfour to start, three to end\n~~~\n",
		"<p>~~~~ perl\nfour to start, three to end\n~~~</p>\n",

		"~~~ bash\ntildes\n~~~\n",
		"<pre><code class=\"language-bash\">tildes\n</code></pre>\n",

		"``` lisp\nno ending\n",
		"<p>``` lisp\nno ending</p>\n",

		"~~~ lisp\nend with language\n~~~ lisp\n",
		"<p>~~~ lisp\nend with language\n~~~ lisp</p>\n",

		"```\nmismatched begin and end\n~~~\n",
		"<p>```\nmismatched begin and end\n~~~</p>\n",

		"~~~\nmismatched begin and end\n```\n",
		"<p>~~~\nmismatched begin and end\n```</p>\n",

		"   ``` oz\nleading spaces\n```\n",
		"<pre><code class=\"language-oz\">leading spaces\n</code></pre>\n",

		"  ``` oz\nleading spaces\n ```\n",
		"<pre><code class=\"language-oz\">leading spaces\n</code></pre>\n",

		" ``` oz\nleading spaces\n  ```\n",
		"<pre><code class=\"language-oz\">leading spaces\n</code></pre>\n",

		"``` oz\nleading spaces\n   ```\n",
		"<pre><code class=\"language-oz\">leading spaces\n</code></pre>\n",

		"    ``` oz\nleading spaces\n    ```\n",
		"<pre><code>``` oz\n</code></pre>\n\n<p>leading spaces\n    ```</p>\n",

		"Bla bla\n\n``` oz\ncode blocks breakup paragraphs\n```\n\nBla Bla\n",
		"<p>Bla bla</p>\n\n<pre><code class=\"language-oz\">code blocks breakup paragraphs\n</code></pre>\n\n<p>Bla Bla</p>\n",

		"Some text before a fenced code block\n``` oz\ncode blocks breakup paragraphs\n```\nAnd some text after a fenced code block",
		"<p>Some text before a fenced code block</p>\n\n<pre><code class=\"language-oz\">code blocks breakup paragraphs\n</code></pre>\n\n<p>And some text after a fenced code block</p>\n",

		"`",
		"<p>`</p>\n",

		"Bla bla\n\n``` oz\ncode blocks breakup paragraphs\n```\n\nBla Bla\n\n``` oz\nmultiple code blocks work okay\n```\n\nBla Bla\n",
		"<p>Bla bla</p>\n\n<pre><code class=\"language-oz\">code blocks breakup paragraphs\n</code></pre>\n\n<p>Bla Bla</p>\n\n<pre><code class=\"language-oz\">multiple code blocks work okay\n</code></pre>\n\n<p>Bla Bla</p>\n",

		"Some text before a fenced code block\n``` oz\ncode blocks breakup paragraphs\n```\nSome text in between\n``` oz\nmultiple code blocks work okay\n```\nAnd some text after a fenced code block",
		"<p>Some text before a fenced code block</p>\n\n<pre><code class=\"language-oz\">code blocks breakup paragraphs\n</code></pre>\n\n<p>Some text in between</p>\n\n<pre><code class=\"language-oz\">multiple code blocks work okay\n</code></pre>\n\n<p>And some text after a fenced code block</p>\n",

		"```\n[]:()\n```\n",
		"<pre><code>[]:()\n</code></pre>\n",

		"```\n[]:()\n[]:)\n[]:(\n[]:x\n[]:testing\n[:testing\n\n[]:\nlinebreak\n[]()\n\n[]:\n[]()\n```",
		"<pre><code>[]:()\n[]:)\n[]:(\n[]:x\n[]:testing\n[:testing\n\n[]:\nlinebreak\n[]()\n\n[]:\n[]()\n</code></pre>\n",
	}
	doTestsBlock(t, tests, FencedCode)
}

func TestFencedCodeInsideBlockquotes(t *testing.T) {
	cat := func(s ...string) string { return strings.Join(s, "\n") }
	var tests = []string{
		cat("> ```go",
			"package moo",
			"",
			"```",
			""),
		`<blockquote>
<pre><code class="language-go">package moo

</code></pre>
</blockquote>
`,
		// -------------------------------------------
		cat("> foo",
			"> ",
			"> ```go",
			"package moo",
			"```",
			"> ",
			"> goo.",
			""),
		`<blockquote>
<p>foo</p>

<pre><code class="language-go">package moo
</code></pre>

<p>goo.</p>
</blockquote>
`,
		// -------------------------------------------
		cat("> foo",
			"> ",
			"> quote",
			"continues",
			"```",
			""),
		`<blockquote>
<p>foo</p>

<p>quote
continues
` + "```" + `</p>
</blockquote>
`,
		// -------------------------------------------
		cat("> foo",
			"> ",
			"> ```go",
			"package moo",
			"```",
			"> ",
			"> goo.",
			"> ",
			"> ```go",
			"package zoo",
			"```",
			"> ",
			"> woo.",
			""),
		`<blockquote>
<p>foo</p>

<pre><code class="language-go">package moo
</code></pre>

<p>goo.</p>

<pre><code class="language-go">package zoo
</code></pre>

<p>woo.</p>
</blockquote>
`,
	}

	// These 2 alternative forms of blockquoted fenced code blocks should produce same output.
	forms := [2]string{
		cat("> plain quoted text",
			"> ```fenced",
			"code",
			" with leading single space correctly preserved",
			"okay",
			"```",
			"> rest of quoted text"),
		cat("> plain quoted text",
			"> ```fenced",
			"> code",
			">  with leading single space correctly preserved",
			"> okay",
			"> ```",
			"> rest of quoted text"),
	}
	want := `<blockquote>
<p>plain quoted text</p>

<pre><code class="language-fenced">code
 with leading single space correctly preserved
okay
</code></pre>

<p>rest of quoted text</p>
</blockquote>
`
	tests = append(tests, forms[0], want)
	tests = append(tests, forms[1], want)

	doTestsBlock(t, tests, FencedCode)
}

func TestTable(t *testing.T) {
	var tests = []string{
		"a | b\n---|---\nc | d\n",
		"<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n\n" +
			"<tbody>\n<tr>\n<td>c</td>\n<td>d</td>\n</tr>\n</tbody>\n</table>\n",

		"a | b\n---|--\nc | d\n",
		"<p>a | b\n---|--\nc | d</p>\n",

		"|a|b|c|d|\n|----|----|----|---|\n|e|f|g|h|\n",
		"<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n<th>c</th>\n<th>d</th>\n</tr>\n</thead>\n\n" +
			"<tbody>\n<tr>\n<td>e</td>\n<td>f</td>\n<td>g</td>\n<td>h</td>\n</tr>\n</tbody>\n</table>\n",

		"*a*|__b__|[c](C)|d\n---|---|---|---\ne|f|g|h\n",
		"<table>\n<thead>\n<tr>\n<th><em>a</em></th>\n<th><strong>b</strong></th>\n<th><a href=\"C\">c</a></th>\n<th>d</th>\n</tr>\n</thead>\n\n" +
			"<tbody>\n<tr>\n<td>e</td>\n<td>f</td>\n<td>g</td>\n<td>h</td>\n</tr>\n</tbody>\n</table>\n",

		"a|b|c\n---|---|---\nd|e|f\ng|h\ni|j|k|l|m\nn|o|p\n",
		"<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n<th>c</th>\n</tr>\n</thead>\n\n" +
			"<tbody>\n<tr>\n<td>d</td>\n<td>e</td>\n<td>f</td>\n</tr>\n\n" +
			"<tr>\n<td>g</td>\n<td>h</td>\n<td></td>\n</tr>\n\n" +
			"<tr>\n<td>i</td>\n<td>j</td>\n<td>k</td>\n</tr>\n\n" +
			"<tr>\n<td>n</td>\n<td>o</td>\n<td>p</td>\n</tr>\n</tbody>\n</table>\n",

		"a|b|c\n---|---|---\n*d*|__e__|f\n",
		"<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n<th>c</th>\n</tr>\n</thead>\n\n" +
			"<tbody>\n<tr>\n<td><em>d</em></td>\n<td><strong>e</strong></td>\n<td>f</td>\n</tr>\n</tbody>\n</table>\n",

		"a|b|c|d\n:--|--:|:-:|---\ne|f|g|h\n",
		"<table>\n<thead>\n<tr>\n<th align=\"left\">a</th>\n<th align=\"right\">b</th>\n" +
			"<th align=\"center\">c</th>\n<th>d</th>\n</tr>\n</thead>\n\n" +
			"<tbody>\n<tr>\n<td align=\"left\">e</td>\n<td align=\"right\">f</td>\n" +
			"<td align=\"center\">g</td>\n<td>h</td>\n</tr>\n</tbody>\n</table>\n",

		"a|b|c\n---|---|---\n",
		"<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n<th>c</th>\n</tr>\n</thead>\n\n<tbody>\n</tbody>\n</table>\n",

		"a| b|c | d | e\n---|---|---|---|---\nf| g|h | i |j\n",
		"<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n<th>c</th>\n<th>d</th>\n<th>e</th>\n</tr>\n</thead>\n\n" +
			"<tbody>\n<tr>\n<td>f</td>\n<td>g</td>\n<td>h</td>\n<td>i</td>\n<td>j</td>\n</tr>\n</tbody>\n</table>\n",

		"a|b\\|c|d\n---|---|---\nf|g\\|h|i\n",
		"<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b|c</th>\n<th>d</th>\n</tr>\n</thead>\n\n<tbody>\n<tr>\n<td>f</td>\n<td>g|h</td>\n<td>i</td>\n</tr>\n</tbody>\n</table>\n",
	}
	doTestsBlock(t, tests, Tables)
}

func TestUnorderedListWith_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	var tests = []string{
		"* Hello\n",
		"<ul>\n<li>Hello</li>\n</ul>\n",

		"* Yin\n* Yang\n",
		"<ul>\n<li>Yin</li>\n<li>Yang</li>\n</ul>\n",

		"* Ting\n* Bong\n* Goo\n",
		"<ul>\n<li>Ting</li>\n<li>Bong</li>\n<li>Goo</li>\n</ul>\n",

		"* Yin\n\n* Yang\n",
		"<ul>\n<li><p>Yin</p></li>\n\n<li><p>Yang</p></li>\n</ul>\n",

		"* Ting\n\n* Bong\n* Goo\n",
		"<ul>\n<li><p>Ting</p></li>\n\n<li><p>Bong</p></li>\n\n<li><p>Goo</p></li>\n</ul>\n",

		"+ Hello\n",
		"<ul>\n<li>Hello</li>\n</ul>\n",

		"+ Yin\n+ Yang\n",
		"<ul>\n<li>Yin</li>\n<li>Yang</li>\n</ul>\n",

		"+ Ting\n+ Bong\n+ Goo\n",
		"<ul>\n<li>Ting</li>\n<li>Bong</li>\n<li>Goo</li>\n</ul>\n",

		"+ Yin\n\n+ Yang\n",
		"<ul>\n<li><p>Yin</p></li>\n\n<li><p>Yang</p></li>\n</ul>\n",

		"+ Ting\n\n+ Bong\n+ Goo\n",
		"<ul>\n<li><p>Ting</p></li>\n\n<li><p>Bong</p></li>\n\n<li><p>Goo</p></li>\n</ul>\n",

		"- Hello\n",
		"<ul>\n<li>Hello</li>\n</ul>\n",

		"- Yin\n- Yang\n",
		"<ul>\n<li>Yin</li>\n<li>Yang</li>\n</ul>\n",

		"- Ting\n- Bong\n- Goo\n",
		"<ul>\n<li>Ting</li>\n<li>Bong</li>\n<li>Goo</li>\n</ul>\n",

		"- Yin\n\n- Yang\n",
		"<ul>\n<li><p>Yin</p></li>\n\n<li><p>Yang</p></li>\n</ul>\n",

		"- Ting\n\n- Bong\n- Goo\n",
		"<ul>\n<li><p>Ting</p></li>\n\n<li><p>Bong</p></li>\n\n<li><p>Goo</p></li>\n</ul>\n",

		"*Hello\n",
		"<p>*Hello</p>\n",

		"*   Hello \n",
		"<ul>\n<li>Hello</li>\n</ul>\n",

		"*   Hello \n    Next line \n",
		"<ul>\n<li>Hello\nNext line</li>\n</ul>\n",

		"Paragraph\n* No linebreak\n",
		"<p>Paragraph</p>\n\n<ul>\n<li>No linebreak</li>\n</ul>\n",

		"Paragraph\n\n* Linebreak\n",
		"<p>Paragraph</p>\n\n<ul>\n<li>Linebreak</li>\n</ul>\n",

		"*   List\n    * Nested list\n",
		"<ul>\n<li>List\n\n<ul>\n<li>Nested list</li>\n</ul></li>\n</ul>\n",

		"*   List\n\n    * Nested list\n",
		"<ul>\n<li><p>List</p>\n\n<ul>\n<li>Nested list</li>\n</ul></li>\n</ul>\n",

		"*   List\n    Second line\n\n    + Nested\n",
		"<ul>\n<li><p>List\nSecond line</p>\n\n<ul>\n<li>Nested</li>\n</ul></li>\n</ul>\n",

		"*   List\n    + Nested\n\n    Continued\n",
		"<ul>\n<li><p>List</p>\n\n<ul>\n<li>Nested</li>\n</ul>\n\n<p>Continued</p></li>\n</ul>\n",

		"*   List\n   * shallow indent\n",
		"<ul>\n<li>List\n\n<ul>\n<li>shallow indent</li>\n</ul></li>\n</ul>\n",

		"* List\n" +
			" * shallow indent\n" +
			"  * part of second list\n" +
			"   * still second\n" +
			"    * almost there\n" +
			"     * third level\n",
		"<ul>\n" +
			"<li>List\n\n" +
			"<ul>\n" +
			"<li>shallow indent</li>\n" +
			"<li>part of second list</li>\n" +
			"<li>still second</li>\n" +
			"<li>almost there\n\n" +
			"<ul>\n" +
			"<li>third level</li>\n" +
			"</ul></li>\n" +
			"</ul></li>\n" +
			"</ul>\n",

		"* List\n        extra indent, same paragraph\n",
		"<ul>\n<li>List\n    extra indent, same paragraph</li>\n</ul>\n",

		"* List\n\n        code block\n",
		"<ul>\n<li><p>List</p>\n\n<pre><code>code block\n</code></pre></li>\n</ul>\n",

		"* List\n\n          code block with spaces\n",
		"<ul>\n<li><p>List</p>\n\n<pre><code>  code block with spaces\n</code></pre></li>\n</ul>\n",

		"* List\n\n    * sublist\n\n    normal text\n\n    * another sublist\n",
		"<ul>\n<li><p>List</p>\n\n<ul>\n<li>sublist</li>\n</ul>\n\n<p>normal text</p>\n\n<ul>\n<li>another sublist</li>\n</ul></li>\n</ul>\n",
	}
	doTestsBlock(t, tests, NoEmptyLineBeforeBlock)
}

func TestOrderedList_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	var tests = []string{
		"1. Hello\n",
		"<ol>\n<li>Hello</li>\n</ol>\n",

		"1. Yin\n2. Yang\n",
		"<ol>\n<li>Yin</li>\n<li>Yang</li>\n</ol>\n",

		"1. Ting\n2. Bong\n3. Goo\n",
		"<ol>\n<li>Ting</li>\n<li>Bong</li>\n<li>Goo</li>\n</ol>\n",

		"1. Yin\n\n2. Yang\n",
		"<ol>\n<li><p>Yin</p></li>\n\n<li><p>Yang</p></li>\n</ol>\n",

		"1. Ting\n\n2. Bong\n3. Goo\n",
		"<ol>\n<li><p>Ting</p></li>\n\n<li><p>Bong</p></li>\n\n<li><p>Goo</p></li>\n</ol>\n",

		"1 Hello\n",
		"<p>1 Hello</p>\n",

		"1.Hello\n",
		"<p>1.Hello</p>\n",

		"1.  Hello \n",
		"<ol>\n<li>Hello</li>\n</ol>\n",

		"1.  Hello \n    Next line \n",
		"<ol>\n<li>Hello\nNext line</li>\n</ol>\n",

		"Paragraph\n1. No linebreak\n",
		"<p>Paragraph</p>\n\n<ol>\n<li>No linebreak</li>\n</ol>\n",

		"Paragraph\n\n1. Linebreak\n",
		"<p>Paragraph</p>\n\n<ol>\n<li>Linebreak</li>\n</ol>\n",

		"1.  List\n    1. Nested list\n",
		"<ol>\n<li>List\n\n<ol>\n<li>Nested list</li>\n</ol></li>\n</ol>\n",

		"1.  List\n\n    1. Nested list\n",
		"<ol>\n<li><p>List</p>\n\n<ol>\n<li>Nested list</li>\n</ol></li>\n</ol>\n",

		"1.  List\n    Second line\n\n    1. Nested\n",
		"<ol>\n<li><p>List\nSecond line</p>\n\n<ol>\n<li>Nested</li>\n</ol></li>\n</ol>\n",

		"1.  List\n    1. Nested\n\n    Continued\n",
		"<ol>\n<li><p>List</p>\n\n<ol>\n<li>Nested</li>\n</ol>\n\n<p>Continued</p></li>\n</ol>\n",

		"1.  List\n   1. shallow indent\n",
		"<ol>\n<li>List\n\n<ol>\n<li>shallow indent</li>\n</ol></li>\n</ol>\n",

		"1. List\n" +
			" 1. shallow indent\n" +
			"  2. part of second list\n" +
			"   3. still second\n" +
			"    4. almost there\n" +
			"     1. third level\n",
		"<ol>\n" +
			"<li>List\n\n" +
			"<ol>\n" +
			"<li>shallow indent</li>\n" +
			"<li>part of second list</li>\n" +
			"<li>still second</li>\n" +
			"<li>almost there\n\n" +
			"<ol>\n" +
			"<li>third level</li>\n" +
			"</ol></li>\n" +
			"</ol></li>\n" +
			"</ol>\n",

		"1. List\n        extra indent, same paragraph\n",
		"<ol>\n<li>List\n    extra indent, same paragraph</li>\n</ol>\n",

		"1. List\n\n        code block\n",
		"<ol>\n<li><p>List</p>\n\n<pre><code>code block\n</code></pre></li>\n</ol>\n",

		"1. List\n\n          code block with spaces\n",
		"<ol>\n<li><p>List</p>\n\n<pre><code>  code block with spaces\n</code></pre></li>\n</ol>\n",

		"1. List\n    * Mixted list\n",
		"<ol>\n<li>List\n\n<ul>\n<li>Mixted list</li>\n</ul></li>\n</ol>\n",

		"1. List\n * Mixed list\n",
		"<ol>\n<li>List\n\n<ul>\n<li>Mixed list</li>\n</ul></li>\n</ol>\n",

		"* Start with unordered\n 1. Ordered\n",
		"<ul>\n<li>Start with unordered\n\n<ol>\n<li>Ordered</li>\n</ol></li>\n</ul>\n",

		"* Start with unordered\n    1. Ordered\n",
		"<ul>\n<li>Start with unordered\n\n<ol>\n<li>Ordered</li>\n</ol></li>\n</ul>\n",

		"1. numbers\n1. are ignored\n",
		"<ol>\n<li>numbers</li>\n<li>are ignored</li>\n</ol>\n",
	}
	doTestsBlock(t, tests, NoEmptyLineBeforeBlock)
}

func TestFencedCodeBlock_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	var tests = []string{
		"``` go\nfunc foo() bool {\n\treturn true;\n}\n```\n",
		"<pre><code class=\"language-go\">func foo() bool {\n\treturn true;\n}\n</code></pre>\n",

		"``` c\n/* special & char < > \" escaping */\n```\n",
		"<pre><code class=\"language-c\">/* special &amp; char &lt; &gt; &quot; escaping */\n</code></pre>\n",

		"``` c\nno *inline* processing ~~of text~~\n```\n",
		"<pre><code class=\"language-c\">no *inline* processing ~~of text~~\n</code></pre>\n",

		"```\nNo language\n```\n",
		"<pre><code>No language\n</code></pre>\n",

		"``` {ocaml}\nlanguage in braces\n```\n",
		"<pre><code class=\"language-ocaml\">language in braces\n</code></pre>\n",

		"```    {ocaml}      \nwith extra whitespace\n```\n",
		"<pre><code class=\"language-ocaml\">with extra whitespace\n</code></pre>\n",

		"```{   ocaml   }\nwith extra whitespace\n```\n",
		"<pre><code class=\"language-ocaml\">with extra whitespace\n</code></pre>\n",

		"~ ~~ java\nWith whitespace\n~~~\n",
		"<p>~ ~~ java\nWith whitespace\n~~~</p>\n",

		"~~\nonly two\n~~\n",
		"<p>~~\nonly two\n~~</p>\n",

		"```` python\nextra\n````\n",
		"<pre><code class=\"language-python\">extra\n</code></pre>\n",

		"~~~ perl\nthree to start, four to end\n~~~~\n",
		"<p>~~~ perl\nthree to start, four to end\n~~~~</p>\n",

		"~~~~ perl\nfour to start, three to end\n~~~\n",
		"<p>~~~~ perl\nfour to start, three to end\n~~~</p>\n",

		"~~~ bash\ntildes\n~~~\n",
		"<pre><code class=\"language-bash\">tildes\n</code></pre>\n",

		"``` lisp\nno ending\n",
		"<p>``` lisp\nno ending</p>\n",

		"~~~ lisp\nend with language\n~~~ lisp\n",
		"<p>~~~ lisp\nend with language\n~~~ lisp</p>\n",

		"```\nmismatched begin and end\n~~~\n",
		"<p>```\nmismatched begin and end\n~~~</p>\n",

		"~~~\nmismatched begin and end\n```\n",
		"<p>~~~\nmismatched begin and end\n```</p>\n",

		"   ``` oz\nleading spaces\n```\n",
		"<pre><code class=\"language-oz\">leading spaces\n</code></pre>\n",

		"  ``` oz\nleading spaces\n ```\n",
		"<pre><code class=\"language-oz\">leading spaces\n</code></pre>\n",

		" ``` oz\nleading spaces\n  ```\n",
		"<pre><code class=\"language-oz\">leading spaces\n</code></pre>\n",

		"``` oz\nleading spaces\n   ```\n",
		"<pre><code class=\"language-oz\">leading spaces\n</code></pre>\n",

		"    ``` oz\nleading spaces\n    ```\n",
		"<pre><code>``` oz\n</code></pre>\n\n<p>leading spaces</p>\n\n<pre><code>```\n</code></pre>\n",
	}
	doTestsBlock(t, tests, FencedCode|NoEmptyLineBeforeBlock)
}

func TestTitleBlock_EXTENSION_TITLEBLOCK(t *testing.T) {
	var tests = []string{
		"% Some title\n" +
			"% Another title line\n" +
			"% Yep, more here too\n",
		"<h1 class=\"title\">" +
			"Some title\n" +
			"Another title line\n" +
			"Yep, more here too" +
			"</h1>\n",
	}
	doTestsBlock(t, tests, Titleblock)
}

func TestBlockComments(t *testing.T) {
	var tests = []string{
		"Some text\n\n<!-- comment -->\n",
		"<p>Some text</p>\n\n<!-- comment -->\n",

		"Some text\n\n<!--\n\nmultiline\ncomment\n-->\n",
		"<p>Some text</p>\n\n<!--\n\nmultiline\ncomment\n-->\n",

		"Some text\n\n<!--\n\n<div><p>Commented</p>\n<span>html</span></div>\n-->\n",
		"<p>Some text</p>\n\n<!--\n\n<div><p>Commented</p>\n<span>html</span></div>\n-->\n",
	}
	doTestsBlock(t, tests, 0)
}

func TestTOC(t *testing.T) {
	var tests = []string{
		"# Title\n\n##Subtitle1\n\n##Subtitle2",
		//"<nav>\n<ul>\n<li><a href=\"#toc_0\">Title</a>\n<ul>\n<li><a href=\"#toc_1\">Subtitle1</a></li>\n<li><a href=\"#toc_2\">Subtitle2</a></li>\n</ul></li>\n</ul>\n</nav>\n\n<h1 id=\"toc_0\">Title</h1>\n\n<h2 id=\"toc_1\">Subtitle1</h2>\n\n<h2 id=\"toc_2\">Subtitle2</h2>\n",
		`<nav>

<ul>
<li><a href="#toc_0">Title</a>
<ul>
<li><a href="#toc_1">Subtitle1</a></li>

<li><a href="#toc_2">Subtitle2</a></li>
</ul></li>
</ul>

</nav>

<h1 id="toc_0">Title</h1>

<h2 id="toc_1">Subtitle1</h2>

<h2 id="toc_2">Subtitle2</h2>
`,

		"# Title\n\n##Subtitle\n\n#Title2",
		//"<nav>\n<ul>\n<li><a href=\"#toc_0\">Title</a>\n<ul>\n<li><a href=\"#toc_1\">Subtitle</a></li>\n</ul></li>\n<li><a href=\"#toc_2\">Title2</a></li>\n</ul>\n</nav>\n\n<h1 id=\"toc_0\">Title</h1>\n\n<h2 id=\"toc_1\">Subtitle</h2>\n\n<h1 id=\"toc_2\">Title2</h1>\n",
		`<nav>

<ul>
<li><a href="#toc_0">Title</a>
<ul>
<li><a href="#toc_1">Subtitle</a></li>
</ul></li>

<li><a href="#toc_2">Title2</a></li>
</ul>

</nav>

<h1 id="toc_0">Title</h1>

<h2 id="toc_1">Subtitle</h2>

<h1 id="toc_2">Title2</h1>
`,

		"## Subtitle\n\n# Title",
		`<nav>

<ul>
<li>
<ul>
<li><a href="#toc_0">Subtitle</a></li>
</ul></li>

<li><a href="#toc_1">Title</a></li>
</ul>

</nav>

<h2 id="toc_0">Subtitle</h2>

<h1 id="toc_1">Title</h1>
`,

		"# Title 1\n\n## Subtitle 1\n\n### Subsubtitle 1\n\n# Title 2\n\n### Subsubtitle 2",
		`<nav>

<ul>
<li><a href="#toc_0">Title 1</a>
<ul>
<li><a href="#toc_1">Subtitle 1</a>
<ul>
<li><a href="#toc_2">Subsubtitle 1</a></li>
</ul></li>
</ul></li>

<li><a href="#toc_3">Title 2</a>
<ul>
<li>
<ul>
<li><a href="#toc_4">Subsubtitle 2</a></li>
</ul></li>
</ul></li>
</ul>

</nav>

<h1 id="toc_0">Title 1</h1>

<h2 id="toc_1">Subtitle 1</h2>

<h3 id="toc_2">Subsubtitle 1</h3>

<h1 id="toc_3">Title 2</h1>

<h3 id="toc_4">Subsubtitle 2</h3>
`,

		"# Title with `code`",
		`<nav>

<ul>
<li><a href="#toc_0">Title with <code>code</code></a></li>
</ul>

</nav>

<h1 id="toc_0">Title with <code>code</code></h1>
`,

		// Trigger empty TOC
		"#",
		"",
	}
	doTestsParam(t, tests, TestParams{
		HTMLFlags: UseXHTML | TOC,
	})
}

func TestCompletePage(t *testing.T) {
	var tests = []string{
		"*foo*",
		`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title></title>
  <meta name="GENERATOR" content="Blackfriday Markdown Processor v2.0" />
  <meta charset="utf-8" />
</head>
<body>

<p><em>foo</em></p>

</body>
</html>
`,
	}
	doTestsParam(t, tests, TestParams{HTMLFlags: UseXHTML | CompletePage})
}

func TestIsFenceLine(t *testing.T) {
	tests := []struct {
		data            []byte
		syntaxRequested bool
		wantEnd         int
		wantMarker      string
		wantSyntax      string
	}{
		{
			data:       []byte("```"),
			wantEnd:    3,
			wantMarker: "```",
		},
		{
			data:       []byte("```\nstuff here\n"),
			wantEnd:    4,
			wantMarker: "```",
		},
		{
			data:            []byte("```\nstuff here\n"),
			syntaxRequested: true,
			wantEnd:         4,
			wantMarker:      "```",
		},
		{
			data:    []byte("stuff here\n```\n"),
			wantEnd: 0,
		},
		{
			data:            []byte("```"),
			syntaxRequested: true,
			wantEnd:         3,
			wantMarker:      "```",
		},
		{
			data:            []byte("``` go"),
			syntaxRequested: true,
			wantEnd:         6,
			wantMarker:      "```",
			wantSyntax:      "go",
		},
	}

	for _, test := range tests {
		var syntax *string
		if test.syntaxRequested {
			syntax = new(string)
		}
		end, marker := isFenceLine(test.data, syntax, "```")
		if got, want := end, test.wantEnd; got != want {
			t.Errorf("got end %v, want %v", got, want)
		}
		if got, want := marker, test.wantMarker; got != want {
			t.Errorf("got marker %q, want %q", got, want)
		}
		if test.syntaxRequested {
			if got, want := *syntax, test.wantSyntax; got != want {
				t.Errorf("got syntax %q, want %q", got, want)
			}
		}
	}
}
