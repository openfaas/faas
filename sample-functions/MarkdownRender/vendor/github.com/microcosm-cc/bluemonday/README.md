# bluemonday [![Build Status](https://travis-ci.org/microcosm-cc/bluemonday.svg?branch=master)](https://travis-ci.org/microcosm-cc/bluemonday) [![GoDoc](https://godoc.org/github.com/microcosm-cc/bluemonday?status.png)](https://godoc.org/github.com/microcosm-cc/bluemonday) [![Sourcegraph](https://sourcegraph.com/github.com/microcosm-cc/bluemonday/-/badge.svg)](https://sourcegraph.com/github.com/microcosm-cc/bluemonday?badge)

bluemonday is a HTML sanitizer implemented in Go. It is fast and highly configurable.

bluemonday takes untrusted user generated content as an input, and will return HTML that has been sanitised against a whitelist of approved HTML elements and attributes so that you can safely include the content in your web page.

If you accept user generated content, and your server uses Go, you **need** bluemonday.

The default policy for user generated content (`bluemonday.UGCPolicy().Sanitize()`) turns this:
```html
Hello <STYLE>.XSS{background-image:url("javascript:alert('XSS')");}</STYLE><A CLASS=XSS></A>World
```

Into a harmless:
```html
Hello World
```

And it turns this:
```html
<a href="javascript:alert('XSS1')" onmouseover="alert('XSS2')">XSS<a>
```

Into this:
```html
XSS
```

Whilst still allowing this:
```html
<a href="http://www.google.com/">
  <img src="https://ssl.gstatic.com/accounts/ui/logo_2x.png"/>
</a>
```

To pass through mostly unaltered (it gained a rel="nofollow" which is a good thing for user generated content):
```html
<a href="http://www.google.com/" rel="nofollow">
  <img src="https://ssl.gstatic.com/accounts/ui/logo_2x.png"/>
</a>
```

It protects sites from [XSS](http://en.wikipedia.org/wiki/Cross-site_scripting) attacks. There are many [vectors for an XSS attack](https://www.owasp.org/index.php/XSS_Filter_Evasion_Cheat_Sheet) and the best way to mitigate the risk is to sanitize user input against a known safe list of HTML elements and attributes.

You should **always** run bluemonday **after** any other processing.

If you use [blackfriday](https://github.com/russross/blackfriday) or [Pandoc](http://johnmacfarlane.net/pandoc/) then bluemonday should be run after these steps. This ensures that no insecure HTML is introduced later in your process.

bluemonday is heavily inspired by both the [OWASP Java HTML Sanitizer](https://code.google.com/p/owasp-java-html-sanitizer/) and the [HTML Purifier](http://htmlpurifier.org/).

## Technical Summary

Whitelist based, you need to either build a policy describing the HTML elements and attributes to permit (and the `regexp` patterns of attributes), or use one of the supplied policies representing good defaults.

The policy containing the whitelist is applied using a fast non-validating, forward only, token-based parser implemented in the [Go net/html library](https://godoc.org/golang.org/x/net/html) by the core Go team.

We expect to be supplied with well-formatted HTML (closing elements for every applicable open element, nested correctly) and so we do not focus on repairing badly nested or incomplete HTML. We focus on simply ensuring that whatever elements do exist are described in the policy whitelist and that attributes and links are safe for use on your web page. [GIGO](http://en.wikipedia.org/wiki/Garbage_in,_garbage_out) does apply and if you feed it bad HTML bluemonday is not tasked with figuring out how to make it good again.

### Supported Go Versions

bluemonday is tested against Go 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 1.9, and tip.

We do not support Go 1.0 as we depend on `golang.org/x/net/html` which includes a reference to `io.ErrNoProgress` which did not exist in Go 1.0.

## Is it production ready?

*Yes*

We are using bluemonday in production having migrated from the widely used and heavily field tested OWASP Java HTML Sanitizer.

We are passing our extensive test suite (including AntiSamy tests as well as tests for any issues raised). Check for any [unresolved issues](https://github.com/microcosm-cc/bluemonday/issues?page=1&state=open) to see whether anything may be a blocker for you.

We invite pull requests and issues to help us ensure we are offering comprehensive protection against various attacks via user generated content.

## Usage

Install in your `${GOPATH}` using `go get -u github.com/microcosm-cc/bluemonday`

Then call it:
```go
package main

import (
	"fmt"

	"github.com/microcosm-cc/bluemonday"
)

func main() {
	p := bluemonday.UGCPolicy()
	html := p.Sanitize(
		`<a onblur="alert(secret)" href="http://www.google.com">Google</a>`,
	)

	// Output:
	// <a href="http://www.google.com" rel="nofollow">Google</a>
	fmt.Println(html)
}
```

We offer three ways to call Sanitize:
```go
p.Sanitize(string) string
p.SanitizeBytes([]byte) []byte
p.SanitizeReader(io.Reader) bytes.Buffer
```

If you are obsessed about performance, `p.SanitizeReader(r).Bytes()` will return a `[]byte` without performing any unnecessary casting of the inputs or outputs. Though the difference is so negligible you should never need to care.

You can build your own policies:
```go
package main

import (
	"fmt"

	"github.com/microcosm-cc/bluemonday"
)

func main() {
	p := bluemonday.NewPolicy()

	// Require URLs to be parseable by net/url.Parse and either:
	//   mailto: http:// or https://
	p.AllowStandardURLs()

	// We only allow <p> and <a href="">
	p.AllowAttrs("href").OnElements("a")
	p.AllowElements("p")

	html := p.Sanitize(
		`<a onblur="alert(secret)" href="http://www.google.com">Google</a>`,
	)

	// Output:
	// <a href="http://www.google.com">Google</a>
	fmt.Println(html)
}
```

We ship two default policies:

1. `bluemonday.StrictPolicy()` which can be thought of as equivalent to stripping all HTML elements and their attributes as it has nothing on it's whitelist. An example usage scenario would be blog post titles where HTML tags are not expected at all and if they are then the elements *and* the content of the elements should be stripped. This is a *very* strict policy.
2. `bluemonday.UGCPolicy()` which allows a broad selection of HTML elements and attributes that are safe for user generated content. Note that this policy does *not* whitelist iframes, object, embed, styles, script, etc. An example usage scenario would be blog post bodies where a variety of formatting is expected along with the potential for TABLEs and IMGs.

## Policy Building

The essence of building a policy is to determine which HTML elements and attributes are considered safe for your scenario. OWASP provide an [XSS prevention cheat sheet](https://www.owasp.org/index.php/XSS_(Cross_Site_Scripting)_Prevention_Cheat_Sheet) to help explain the risks, but essentially:

1. Avoid anything other than the standard HTML elements
1. Avoid `script`, `style`, `iframe`, `object`, `embed`, `base` elements that allow code to be executed by the client or third party content to be included that can execute code
1. Avoid anything other than plain HTML attributes with values matched to a regexp

Basically, you should be able to describe what HTML is fine for your scenario. If you do not have confidence that you can describe your policy please consider using one of the shipped policies such as `bluemonday.UGCPolicy()`.

To create a new policy:
```go
p := bluemonday.NewPolicy()
```

To add elements to a policy either add just the elements:
```go
p.AllowElements("b", "strong")
```

Or add elements as a virtue of adding an attribute:
```go
// Not the recommended pattern, see the recommendation on using .Matching() below
p.AllowAttrs("nowrap").OnElements("td", "th")
```

Attributes can either be added to all elements:
```go
p.AllowAttrs("dir").Matching(regexp.MustCompile("(?i)rtl|ltr")).Globally()
```

Or attributes can be added to specific elements:
```go
// Not the recommended pattern, see the recommendation on using .Matching() below
p.AllowAttrs("value").OnElements("li")
```

It is **always** recommended that an attribute be made to match a pattern. XSS in HTML attributes is very easy otherwise:
```go
// \p{L} matches unicode letters, \p{N} matches unicode numbers
p.AllowAttrs("title").Matching(regexp.MustCompile(`[\p{L}\p{N}\s\-_',:\[\]!\./\\\(\)&]*`)).Globally()
```

You can stop at any time and call .Sanitize():
```go
// string htmlIn passed in from a HTTP POST
htmlOut := p.Sanitize(htmlIn)
```

And you can take any existing policy and extend it:
```go
p := bluemonday.UGCPolicy()
p.AllowElements("fieldset", "select", "option")
```

### Links

Links are difficult beasts to sanitise safely and also one of the biggest attack vectors for malicious content.

It is possible to do this:
```go
p.AllowAttrs("href").Matching(regexp.MustCompile(`(?i)mailto|https?`)).OnElements("a")
```

But that will not protect you as the regular expression is insufficient in this case to have prevented a malformed value doing something unexpected.

We provide some additional global options for safely working with links.

`RequireParseableURLs` will ensure that URLs are parseable by Go's `net/url` package:
```go
p.RequireParseableURLs(true)
```

If you have enabled parseable URLs then the following option will `AllowRelativeURLs`. By default this is disabled (bluemonday is a whitelist tool... you need to explicitly tell us to permit things) and when disabled it will prevent all local and scheme relative URLs (i.e. `href="localpage.html"`, `href="../home.html"` and even `href="//www.google.com"` are relative):
```go
p.AllowRelativeURLs(true)
```

If you have enabled parseable URLs then you can whitelist the schemes (commonly called protocol when thinking of `http` and `https`) that are permitted. Bear in mind that allowing relative URLs in the above option will allow for a blank scheme:
```go
p.AllowURLSchemes("mailto", "http", "https")
```

Regardless of whether you have enabled parseable URLs, you can force all URLs to have a rel="nofollow" attribute. This will be added if it does not exist, but only when the `href` is valid:
```go
// This applies to "a" "area" "link" elements that have a "href" attribute
p.RequireNoFollowOnLinks(true)
```

We provide a convenience method that applies all of the above, but you will still need to whitelist the linkable elements for the URL rules to be applied to:
```go
p.AllowStandardURLs()
p.AllowAttrs("cite").OnElements("blockquote", "q")
p.AllowAttrs("href").OnElements("a", "area")
p.AllowAttrs("src").OnElements("img")
```

An additional complexity regarding links is the data URI as defined in [RFC2397](http://tools.ietf.org/html/rfc2397). The data URI allows for images to be served inline using this format:

```html
<img src="data:image/webp;base64,UklGRh4AAABXRUJQVlA4TBEAAAAvAAAAAAfQ//73v/+BiOh/AAA=">
```

We have provided a helper to verify the mimetype followed by base64 content of data URIs links:

```go
p.AllowDataURIImages()
```

That helper will enable GIF, JPEG, PNG and WEBP images.

It should be noted that there is a potential [security](http://palizine.plynt.com/issues/2010Oct/bypass-xss-filters/) [risk](https://capec.mitre.org/data/definitions/244.html) with the use of data URI links. You should only enable data URI links if you already trust the content.

We also have some features to help deal with user generated content:
```go
p.AddTargetBlankToFullyQualifiedLinks(true)
```

This will ensure that anchor `<a href="" />` links that are fully qualified (the href destination includes a host name) will get `target="_blank"` added to them.

Additionally any link that has `target="_blank"` after the policy has been applied will also have the `rel` attribute adjusted to add `noopener`. This means a link may start like `<a href="//host/path"/>` and will end up as `<a href="//host/path" rel="noopener" target="_blank">`. It is important to note that the addition of `noopener` is a security feature and not an issue. There is an unfortunate feature to browsers that a browser window opened as a result of `target="_blank"` can still control the opener (your web page) and this protects against that. The background to this can be found here: [https://dev.to/ben/the-targetblank-vulnerability-by-example](https://dev.to/ben/the-targetblank-vulnerability-by-example)

### Policy Building Helpers

We also bundle some helpers to simplify policy building:
```go

// Permits the "dir", "id", "lang", "title" attributes globally
p.AllowStandardAttributes()

// Permits the "img" element and it's standard attributes
p.AllowImages()

// Permits ordered and unordered lists, and also definition lists
p.AllowLists()

// Permits HTML tables and all applicable elements and non-styling attributes
p.AllowTables()
```

### Invalid Instructions

The following are invalid:
```go
// This does not say where the attributes are allowed, you need to add
// .Globally() or .OnElements(...)
// This will be ignored without error.
p.AllowAttrs("value")

// This does not say where the attributes are allowed, you need to add
// .Globally() or .OnElements(...)
// This will be ignored without error.
p.AllowAttrs(
	"type",
).Matching(
	regexp.MustCompile("(?i)^(circle|disc|square|a|A|i|I|1)$"),
)
```

Both examples exhibit the same issue, they declare attributes but do not then specify whether they are whitelisted globally or only on specific elements (and which elements). Attributes belong to one or more elements, and the policy needs to declare this.

## Limitations

We are not yet including any tools to help whitelist and sanitize CSS. Which means that unless you wish to do the heavy lifting in a single regular expression (inadvisable), **you should not allow the "style" attribute anywhere**.

It is not the job of bluemonday to fix your bad HTML, it is merely the job of bluemonday to prevent malicious HTML getting through. If you have mismatched HTML elements, or non-conforming nesting of elements, those will remain. But if you have well-structured HTML bluemonday will not break it.

## TODO

* Add support for CSS sanitisation to allow some CSS properties based on a whitelist, possibly using the [Gorilla CSS3 scanner](http://www.gorillatoolkit.org/pkg/css/scanner)
* Investigate whether devs want to blacklist elements and attributes. This would allow devs to take an existing policy (such as the `bluemonday.UGCPolicy()` ) that encapsulates 90% of what they're looking for but does more than they need, and to remove the extra things they do not want to make it 100% what they want
* Investigate whether devs want a validating HTML mode, in which the HTML elements are not just transformed into a balanced tree (every start tag has a closing tag at the correct depth) but also that elements and character data appear only in their allowed context (i.e. that a `table` element isn't a descendent of a `caption`, that `colgroup`, `thead`, `tbody`, `tfoot` and `tr` are permitted, and that character data is not permitted)

## Development

If you have cloned this repo you will probably need the dependency:

`go get golang.org/x/net/html`

Gophers can use their familiar tools:

`go build`

`go test`

I personally use a Makefile as it spares typing the same args over and over whilst providing consistency for those of us who jump from language to language and enjoy just typing `make` in a project directory and watch magic happen.

`make` will build, vet, test and install the library.

`make clean` will remove the library from a *single* `${GOPATH}/pkg` directory tree

`make test` will run the tests

`make cover` will run the tests and *open a browser window* with the coverage report

`make lint` will run golint (install via `go get github.com/golang/lint/golint`)

## Long term goals

1. Open the code to adversarial peer review similar to the [Attack Review Ground Rules](https://code.google.com/p/owasp-java-html-sanitizer/wiki/AttackReviewGroundRules)
1. Raise funds and pay for an external security review
