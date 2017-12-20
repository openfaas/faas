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
	"net/url"
	"regexp"
	"strings"
)

// Policy encapsulates the whitelist of HTML elements and attributes that will
// be applied to the sanitised HTML.
//
// You should use bluemonday.NewPolicy() to create a blank policy as the
// unexported fields contain maps that need to be initialized.
type Policy struct {

	// Declares whether the maps have been initialized, used as a cheap check to
	// ensure that those using Policy{} directly won't cause nil pointer
	// exceptions
	initialized bool

	// Allows the <!DOCTYPE > tag to exist in the sanitized document
	allowDocType bool

	// If true then we add spaces when stripping tags, specifically the closing
	// tag is replaced by a space character.
	addSpaces bool

	// When true, add rel="nofollow" to HTML anchors
	requireNoFollow bool

	// When true, add rel="nofollow" to HTML anchors
	// Will add for href="http://foo"
	// Will skip for href="/foo" or href="foo"
	requireNoFollowFullyQualifiedLinks bool

	// When true add target="_blank" to fully qualified links
	// Will add for href="http://foo"
	// Will skip for href="/foo" or href="foo"
	addTargetBlankToFullyQualifiedLinks bool

	// When true, URLs must be parseable by "net/url" url.Parse()
	requireParseableURLs bool

	// When true, u, _ := url.Parse("url"); !u.IsAbs() is permitted
	allowRelativeURLs bool

	// map[htmlElementName]map[htmlAttributeName]attrPolicy
	elsAndAttrs map[string]map[string]attrPolicy

	// map[htmlAttributeName]attrPolicy
	globalAttrs map[string]attrPolicy

	// If urlPolicy is nil, all URLs with matching schema are allowed.
	// Otherwise, only the URLs with matching schema and urlPolicy(url)
	// returning true are allowed.
	allowURLSchemes map[string]urlPolicy

	// If an element has had all attributes removed as a result of a policy
	// being applied, then the element would be removed from the output.
	//
	// However some elements are valid and have strong layout meaning without
	// any attributes, i.e. <table>. To prevent those being removed we maintain
	// a list of elements that are allowed to have no attributes and that will
	// be maintained in the output HTML.
	setOfElementsAllowedWithoutAttrs map[string]struct{}

	setOfElementsToSkipContent map[string]struct{}
}

type attrPolicy struct {

	// optional pattern to match, when not nil the regexp needs to match
	// otherwise the attribute is removed
	regexp *regexp.Regexp
}

type attrPolicyBuilder struct {
	p *Policy

	attrNames  []string
	regexp     *regexp.Regexp
	allowEmpty bool
}

type urlPolicy func(url *url.URL) (allowUrl bool)

// init initializes the maps if this has not been done already
func (p *Policy) init() {
	if !p.initialized {
		p.elsAndAttrs = make(map[string]map[string]attrPolicy)
		p.globalAttrs = make(map[string]attrPolicy)
		p.allowURLSchemes = make(map[string]urlPolicy)
		p.setOfElementsAllowedWithoutAttrs = make(map[string]struct{})
		p.setOfElementsToSkipContent = make(map[string]struct{})
		p.initialized = true
	}
}

// NewPolicy returns a blank policy with nothing whitelisted or permitted. This
// is the recommended way to start building a policy and you should now use
// AllowAttrs() and/or AllowElements() to construct the whitelist of HTML
// elements and attributes.
func NewPolicy() *Policy {

	p := Policy{}

	p.addDefaultElementsWithoutAttrs()
	p.addDefaultSkipElementContent()

	return &p
}

// AllowAttrs takes a range of HTML attribute names and returns an
// attribute policy builder that allows you to specify the pattern and scope of
// the whitelisted attribute.
//
// The attribute policy is only added to the core policy when either Globally()
// or OnElements(...) are called.
func (p *Policy) AllowAttrs(attrNames ...string) *attrPolicyBuilder {

	p.init()

	abp := attrPolicyBuilder{
		p:          p,
		allowEmpty: false,
	}

	for _, attrName := range attrNames {
		abp.attrNames = append(abp.attrNames, strings.ToLower(attrName))
	}

	return &abp
}

// AllowNoAttrs says that attributes on element are optional.
//
// The attribute policy is only added to the core policy when OnElements(...)
// are called.
func (p *Policy) AllowNoAttrs() *attrPolicyBuilder {

	p.init()

	abp := attrPolicyBuilder{
		p:          p,
		allowEmpty: true,
	}
	return &abp
}

// AllowNoAttrs says that attributes on element are optional.
//
// The attribute policy is only added to the core policy when OnElements(...)
// are called.
func (abp *attrPolicyBuilder) AllowNoAttrs() *attrPolicyBuilder {

	abp.allowEmpty = true

	return abp
}

// Matching allows a regular expression to be applied to a nascent attribute
// policy, and returns the attribute policy. Calling this more than once will
// replace the existing regexp.
func (abp *attrPolicyBuilder) Matching(regex *regexp.Regexp) *attrPolicyBuilder {

	abp.regexp = regex

	return abp
}

// OnElements will bind an attribute policy to a given range of HTML elements
// and return the updated policy
func (abp *attrPolicyBuilder) OnElements(elements ...string) *Policy {

	for _, element := range elements {
		element = strings.ToLower(element)

		for _, attr := range abp.attrNames {

			if _, ok := abp.p.elsAndAttrs[element]; !ok {
				abp.p.elsAndAttrs[element] = make(map[string]attrPolicy)
			}

			ap := attrPolicy{}
			if abp.regexp != nil {
				ap.regexp = abp.regexp
			}

			abp.p.elsAndAttrs[element][attr] = ap
		}

		if abp.allowEmpty {
			abp.p.setOfElementsAllowedWithoutAttrs[element] = struct{}{}

			if _, ok := abp.p.elsAndAttrs[element]; !ok {
				abp.p.elsAndAttrs[element] = make(map[string]attrPolicy)
			}
		}
	}

	return abp.p
}

// Globally will bind an attribute policy to all HTML elements and return the
// updated policy
func (abp *attrPolicyBuilder) Globally() *Policy {

	for _, attr := range abp.attrNames {
		if _, ok := abp.p.globalAttrs[attr]; !ok {
			abp.p.globalAttrs[attr] = attrPolicy{}
		}

		ap := attrPolicy{}
		if abp.regexp != nil {
			ap.regexp = abp.regexp
		}

		abp.p.globalAttrs[attr] = ap
	}

	return abp.p
}

// AllowElements will append HTML elements to the whitelist without applying an
// attribute policy to those elements (the elements are permitted
// sans-attributes)
func (p *Policy) AllowElements(names ...string) *Policy {
	p.init()

	for _, element := range names {
		element = strings.ToLower(element)

		if _, ok := p.elsAndAttrs[element]; !ok {
			p.elsAndAttrs[element] = make(map[string]attrPolicy)
		}
	}

	return p
}

// RequireNoFollowOnLinks will result in all <a> tags having a rel="nofollow"
// added to them if one does not already exist
//
// Note: This requires p.RequireParseableURLs(true) and will enable it.
func (p *Policy) RequireNoFollowOnLinks(require bool) *Policy {

	p.requireNoFollow = require
	p.requireParseableURLs = true

	return p
}

// RequireNoFollowOnFullyQualifiedLinks will result in all <a> tags that point
// to a non-local destination (i.e. starts with a protocol and has a host)
// having a rel="nofollow" added to them if one does not already exist
//
// Note: This requires p.RequireParseableURLs(true) and will enable it.
func (p *Policy) RequireNoFollowOnFullyQualifiedLinks(require bool) *Policy {

	p.requireNoFollowFullyQualifiedLinks = require
	p.requireParseableURLs = true

	return p
}

// AddTargetBlankToFullyQualifiedLinks will result in all <a> tags that point
// to a non-local destination (i.e. starts with a protocol and has a host)
// having a target="_blank" added to them if one does not already exist
//
// Note: This requires p.RequireParseableURLs(true) and will enable it.
func (p *Policy) AddTargetBlankToFullyQualifiedLinks(require bool) *Policy {

	p.addTargetBlankToFullyQualifiedLinks = require
	p.requireParseableURLs = true

	return p
}

// RequireParseableURLs will result in all URLs requiring that they be parseable
// by "net/url" url.Parse()
// This applies to:
// - a.href
// - area.href
// - blockquote.cite
// - img.src
// - link.href
// - script.src
func (p *Policy) RequireParseableURLs(require bool) *Policy {

	p.requireParseableURLs = require

	return p
}

// AllowRelativeURLs enables RequireParseableURLs and then permits URLs that
// are parseable, have no schema information and url.IsAbs() returns false
// This permits local URLs
func (p *Policy) AllowRelativeURLs(require bool) *Policy {

	p.RequireParseableURLs(true)
	p.allowRelativeURLs = require

	return p
}

// AllowURLSchemes will append URL schemes to the whitelist
// Example: p.AllowURLSchemes("mailto", "http", "https")
func (p *Policy) AllowURLSchemes(schemes ...string) *Policy {
	p.init()

	p.RequireParseableURLs(true)

	for _, scheme := range schemes {
		scheme = strings.ToLower(scheme)

		// Allow all URLs with matching scheme.
		p.allowURLSchemes[scheme] = nil
	}

	return p
}

// AllowURLSchemeWithCustomPolicy will append URL schemes with
// a custom URL policy to the whitelist.
// Only the URLs with matching schema and urlPolicy(url)
// returning true will be allowed.
func (p *Policy) AllowURLSchemeWithCustomPolicy(
	scheme string,
	urlPolicy func(url *url.URL) (allowUrl bool),
) *Policy {

	p.init()

	p.RequireParseableURLs(true)

	scheme = strings.ToLower(scheme)

	p.allowURLSchemes[scheme] = urlPolicy

	return p
}

// AllowDocType states whether the HTML sanitised by the sanitizer is allowed to
// contain the HTML DocType tag: <!DOCTYPE HTML> or one of it's variants.
//
// The HTML spec only permits one doctype per document, and as you know how you
// are using the output of this, you know best as to whether we should ignore it
// (default) or not.
//
// If you are sanitizing a HTML fragment the default (false) is fine.
func (p *Policy) AllowDocType(allow bool) *Policy {

	p.allowDocType = allow

	return p
}

// AddSpaceWhenStrippingTag states whether to add a single space " " when
// removing tags that are not whitelisted by the policy.
//
// This is useful if you expect to strip tags in dense markup and may lose the
// value of whitespace.
//
// For example: "<p>Hello</p><p>World</p>"" would be sanitized to "HelloWorld"
// with the default value of false, but you may wish to sanitize this to
// " Hello  World " by setting AddSpaceWhenStrippingTag to true as this would
// retain the intent of the text.
func (p *Policy) AddSpaceWhenStrippingTag(allow bool) *Policy {

	p.addSpaces = allow

	return p
}

// SkipElementsContent adds the HTML elements whose tags is needed to be removed
// with it's content.
func (p *Policy) SkipElementsContent(names ...string) *Policy {

	p.init()

	for _, element := range names {
		element = strings.ToLower(element)

		if _, ok := p.setOfElementsToSkipContent[element]; !ok {
			p.setOfElementsToSkipContent[element] = struct{}{}
		}
	}

	return p
}

// AllowElementsContent marks the HTML elements whose content should be
// retained after removing the tag.
func (p *Policy) AllowElementsContent(names ...string) *Policy {

	p.init()

	for _, element := range names {
		delete(p.setOfElementsToSkipContent, strings.ToLower(element))
	}

	return p
}

// addDefaultElementsWithoutAttrs adds the HTML elements that we know are valid
// without any attributes to an internal map.
// i.e. we know that <table> is valid, but <bdo> isn't valid as the "dir" attr
// is mandatory
func (p *Policy) addDefaultElementsWithoutAttrs() {
	p.init()

	p.setOfElementsAllowedWithoutAttrs["abbr"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["acronym"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["article"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["aside"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["audio"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["b"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["bdi"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["blockquote"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["body"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["br"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["button"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["canvas"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["caption"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["center"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["cite"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["code"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["col"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["colgroup"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["datalist"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["dd"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["del"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["details"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["dfn"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["div"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["dl"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["dt"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["em"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["fieldset"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["figcaption"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["figure"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["footer"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["h1"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["h2"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["h3"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["h4"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["h5"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["h6"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["head"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["header"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["hgroup"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["hr"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["html"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["i"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["ins"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["kbd"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["li"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["mark"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["marquee"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["nav"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["ol"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["optgroup"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["option"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["p"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["pre"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["q"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["rp"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["rt"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["ruby"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["s"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["samp"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["section"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["select"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["small"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["span"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["strike"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["strong"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["style"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["sub"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["summary"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["sup"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["svg"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["table"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["tbody"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["td"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["textarea"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["tfoot"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["th"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["thead"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["title"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["time"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["tr"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["tt"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["u"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["ul"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["var"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["video"] = struct{}{}
	p.setOfElementsAllowedWithoutAttrs["wbr"] = struct{}{}

}

// addDefaultSkipElementContent adds the HTML elements that we should skip
// rendering the character content of, if the element itself is not allowed.
// This is all character data that the end user would not normally see.
// i.e. if we exclude a <script> tag then we shouldn't render the JavaScript or
// anything else until we encounter the closing </script> tag.
func (p *Policy) addDefaultSkipElementContent() {
	p.init()

	p.setOfElementsToSkipContent["frame"] = struct{}{}
	p.setOfElementsToSkipContent["frameset"] = struct{}{}
	p.setOfElementsToSkipContent["iframe"] = struct{}{}
	p.setOfElementsToSkipContent["noembed"] = struct{}{}
	p.setOfElementsToSkipContent["noframes"] = struct{}{}
	p.setOfElementsToSkipContent["noscript"] = struct{}{}
	p.setOfElementsToSkipContent["nostyle"] = struct{}{}
	p.setOfElementsToSkipContent["object"] = struct{}{}
	p.setOfElementsToSkipContent["script"] = struct{}{}
	p.setOfElementsToSkipContent["style"] = struct{}{}
	p.setOfElementsToSkipContent["title"] = struct{}{}
}
