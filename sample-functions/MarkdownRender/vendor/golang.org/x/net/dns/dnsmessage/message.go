// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dnsmessage provides a mostly RFC 1035 compliant implementation of
// DNS message packing and unpacking.
//
// This implementation is designed to minimize heap allocations and avoid
// unnecessary packing and unpacking as much as possible.
package dnsmessage

import (
	"errors"
)

// Packet formats

// A Type is a type of DNS request and response.
type Type uint16

// A Class is a type of network.
type Class uint16

// An OpCode is a DNS operation code.
type OpCode uint16

// An RCode is a DNS response status code.
type RCode uint16

// Wire constants.
const (
	// ResourceHeader.Type and Question.Type
	TypeA     Type = 1
	TypeNS    Type = 2
	TypeCNAME Type = 5
	TypeSOA   Type = 6
	TypePTR   Type = 12
	TypeMX    Type = 15
	TypeTXT   Type = 16
	TypeAAAA  Type = 28
	TypeSRV   Type = 33

	// Question.Type
	TypeWKS   Type = 11
	TypeHINFO Type = 13
	TypeMINFO Type = 14
	TypeAXFR  Type = 252
	TypeALL   Type = 255

	// ResourceHeader.Class and Question.Class
	ClassINET   Class = 1
	ClassCSNET  Class = 2
	ClassCHAOS  Class = 3
	ClassHESIOD Class = 4

	// Question.Class
	ClassANY Class = 255

	// Message.Rcode
	RCodeSuccess        RCode = 0
	RCodeFormatError    RCode = 1
	RCodeServerFailure  RCode = 2
	RCodeNameError      RCode = 3
	RCodeNotImplemented RCode = 4
	RCodeRefused        RCode = 5
)

var (
	// ErrNotStarted indicates that the prerequisite information isn't
	// available yet because the previous records haven't been appropriately
	// parsed, skipped or finished.
	ErrNotStarted = errors.New("parsing/packing of this type isn't available yet")

	// ErrSectionDone indicated that all records in the section have been
	// parsed or finished.
	ErrSectionDone = errors.New("parsing/packing of this section has completed")

	errBaseLen            = errors.New("insufficient data for base length type")
	errCalcLen            = errors.New("insufficient data for calculated length type")
	errReserved           = errors.New("segment prefix is reserved")
	errTooManyPtr         = errors.New("too many pointers (>10)")
	errInvalidPtr         = errors.New("invalid pointer")
	errNilResouceBody     = errors.New("nil resource body")
	errResourceLen        = errors.New("insufficient data for resource body length")
	errSegTooLong         = errors.New("segment length too long")
	errZeroSegLen         = errors.New("zero length segment")
	errResTooLong         = errors.New("resource length too long")
	errTooManyQuestions   = errors.New("too many Questions to pack (>65535)")
	errTooManyAnswers     = errors.New("too many Answers to pack (>65535)")
	errTooManyAuthorities = errors.New("too many Authorities to pack (>65535)")
	errTooManyAdditionals = errors.New("too many Additionals to pack (>65535)")
	errNonCanonicalName   = errors.New("name is not in canonical format (it must end with a .)")
)

// Internal constants.
const (
	// packStartingCap is the default initial buffer size allocated during
	// packing.
	//
	// The starting capacity doesn't matter too much, but most DNS responses
	// Will be <= 512 bytes as it is the limit for DNS over UDP.
	packStartingCap = 512

	// uint16Len is the length (in bytes) of a uint16.
	uint16Len = 2

	// uint32Len is the length (in bytes) of a uint32.
	uint32Len = 4

	// headerLen is the length (in bytes) of a DNS header.
	//
	// A header is comprised of 6 uint16s and no padding.
	headerLen = 6 * uint16Len
)

type nestedError struct {
	// s is the current level's error message.
	s string

	// err is the nested error.
	err error
}

// nestedError implements error.Error.
func (e *nestedError) Error() string {
	return e.s + ": " + e.err.Error()
}

// Header is a representation of a DNS message header.
type Header struct {
	ID                 uint16
	Response           bool
	OpCode             OpCode
	Authoritative      bool
	Truncated          bool
	RecursionDesired   bool
	RecursionAvailable bool
	RCode              RCode
}

func (m *Header) pack() (id uint16, bits uint16) {
	id = m.ID
	bits = uint16(m.OpCode)<<11 | uint16(m.RCode)
	if m.RecursionAvailable {
		bits |= headerBitRA
	}
	if m.RecursionDesired {
		bits |= headerBitRD
	}
	if m.Truncated {
		bits |= headerBitTC
	}
	if m.Authoritative {
		bits |= headerBitAA
	}
	if m.Response {
		bits |= headerBitQR
	}
	return
}

// Message is a representation of a DNS message.
type Message struct {
	Header
	Questions   []Question
	Answers     []Resource
	Authorities []Resource
	Additionals []Resource
}

type section uint8

const (
	sectionNotStarted section = iota
	sectionHeader
	sectionQuestions
	sectionAnswers
	sectionAuthorities
	sectionAdditionals
	sectionDone

	headerBitQR = 1 << 15 // query/response (response=1)
	headerBitAA = 1 << 10 // authoritative
	headerBitTC = 1 << 9  // truncated
	headerBitRD = 1 << 8  // recursion desired
	headerBitRA = 1 << 7  // recursion available
)

var sectionNames = map[section]string{
	sectionHeader:      "header",
	sectionQuestions:   "Question",
	sectionAnswers:     "Answer",
	sectionAuthorities: "Authority",
	sectionAdditionals: "Additional",
}

// header is the wire format for a DNS message header.
type header struct {
	id          uint16
	bits        uint16
	questions   uint16
	answers     uint16
	authorities uint16
	additionals uint16
}

func (h *header) count(sec section) uint16 {
	switch sec {
	case sectionQuestions:
		return h.questions
	case sectionAnswers:
		return h.answers
	case sectionAuthorities:
		return h.authorities
	case sectionAdditionals:
		return h.additionals
	}
	return 0
}

func (h *header) pack(msg []byte) []byte {
	msg = packUint16(msg, h.id)
	msg = packUint16(msg, h.bits)
	msg = packUint16(msg, h.questions)
	msg = packUint16(msg, h.answers)
	msg = packUint16(msg, h.authorities)
	return packUint16(msg, h.additionals)
}

func (h *header) unpack(msg []byte, off int) (int, error) {
	newOff := off
	var err error
	if h.id, newOff, err = unpackUint16(msg, newOff); err != nil {
		return off, &nestedError{"id", err}
	}
	if h.bits, newOff, err = unpackUint16(msg, newOff); err != nil {
		return off, &nestedError{"bits", err}
	}
	if h.questions, newOff, err = unpackUint16(msg, newOff); err != nil {
		return off, &nestedError{"questions", err}
	}
	if h.answers, newOff, err = unpackUint16(msg, newOff); err != nil {
		return off, &nestedError{"answers", err}
	}
	if h.authorities, newOff, err = unpackUint16(msg, newOff); err != nil {
		return off, &nestedError{"authorities", err}
	}
	if h.additionals, newOff, err = unpackUint16(msg, newOff); err != nil {
		return off, &nestedError{"additionals", err}
	}
	return newOff, nil
}

func (h *header) header() Header {
	return Header{
		ID:                 h.id,
		Response:           (h.bits & headerBitQR) != 0,
		OpCode:             OpCode(h.bits>>11) & 0xF,
		Authoritative:      (h.bits & headerBitAA) != 0,
		Truncated:          (h.bits & headerBitTC) != 0,
		RecursionDesired:   (h.bits & headerBitRD) != 0,
		RecursionAvailable: (h.bits & headerBitRA) != 0,
		RCode:              RCode(h.bits & 0xF),
	}
}

// A Resource is a DNS resource record.
type Resource struct {
	Header ResourceHeader
	Body   ResourceBody
}

// A ResourceBody is a DNS resource record minus the header.
type ResourceBody interface {
	// pack packs a Resource except for its header.
	pack(msg []byte, compression map[string]int) ([]byte, error)

	// realType returns the actual type of the Resource. This is used to
	// fill in the header Type field.
	realType() Type
}

func (r *Resource) pack(msg []byte, compression map[string]int) ([]byte, error) {
	if r.Body == nil {
		return msg, errNilResouceBody
	}
	oldMsg := msg
	r.Header.Type = r.Body.realType()
	msg, length, err := r.Header.pack(msg, compression)
	if err != nil {
		return msg, &nestedError{"ResourceHeader", err}
	}
	preLen := len(msg)
	msg, err = r.Body.pack(msg, compression)
	if err != nil {
		return msg, &nestedError{"content", err}
	}
	if err := r.Header.fixLen(msg, length, preLen); err != nil {
		return oldMsg, err
	}
	return msg, nil
}

// A Parser allows incrementally parsing a DNS message.
//
// When parsing is started, the Header is parsed. Next, each Question can be
// either parsed or skipped. Alternatively, all Questions can be skipped at
// once. When all Questions have been parsed, attempting to parse Questions
// will return (nil, nil) and attempting to skip Questions will return
// (true, nil). After all Questions have been either parsed or skipped, all
// Answers, Authorities and Additionals can be either parsed or skipped in the
// same way, and each type of Resource must be fully parsed or skipped before
// proceeding to the next type of Resource.
//
// Note that there is no requirement to fully skip or parse the message.
type Parser struct {
	msg    []byte
	header header

	section        section
	off            int
	index          int
	resHeaderValid bool
	resHeader      ResourceHeader
}

// Start parses the header and enables the parsing of Questions.
func (p *Parser) Start(msg []byte) (Header, error) {
	if p.msg != nil {
		*p = Parser{}
	}
	p.msg = msg
	var err error
	if p.off, err = p.header.unpack(msg, 0); err != nil {
		return Header{}, &nestedError{"unpacking header", err}
	}
	p.section = sectionQuestions
	return p.header.header(), nil
}

func (p *Parser) checkAdvance(sec section) error {
	if p.section < sec {
		return ErrNotStarted
	}
	if p.section > sec {
		return ErrSectionDone
	}
	p.resHeaderValid = false
	if p.index == int(p.header.count(sec)) {
		p.index = 0
		p.section++
		return ErrSectionDone
	}
	return nil
}

func (p *Parser) resource(sec section) (Resource, error) {
	var r Resource
	var err error
	r.Header, err = p.resourceHeader(sec)
	if err != nil {
		return r, err
	}
	p.resHeaderValid = false
	r.Body, p.off, err = unpackResourceBody(p.msg, p.off, r.Header)
	if err != nil {
		return Resource{}, &nestedError{"unpacking " + sectionNames[sec], err}
	}
	p.index++
	return r, nil
}

func (p *Parser) resourceHeader(sec section) (ResourceHeader, error) {
	if p.resHeaderValid {
		return p.resHeader, nil
	}
	if err := p.checkAdvance(sec); err != nil {
		return ResourceHeader{}, err
	}
	var hdr ResourceHeader
	off, err := hdr.unpack(p.msg, p.off)
	if err != nil {
		return ResourceHeader{}, err
	}
	p.resHeaderValid = true
	p.resHeader = hdr
	p.off = off
	return hdr, nil
}

func (p *Parser) skipResource(sec section) error {
	if p.resHeaderValid {
		newOff := p.off + int(p.resHeader.Length)
		if newOff > len(p.msg) {
			return errResourceLen
		}
		p.off = newOff
		p.resHeaderValid = false
		p.index++
		return nil
	}
	if err := p.checkAdvance(sec); err != nil {
		return err
	}
	var err error
	p.off, err = skipResource(p.msg, p.off)
	if err != nil {
		return &nestedError{"skipping: " + sectionNames[sec], err}
	}
	p.index++
	return nil
}

// Question parses a single Question.
func (p *Parser) Question() (Question, error) {
	if err := p.checkAdvance(sectionQuestions); err != nil {
		return Question{}, err
	}
	var name Name
	off, err := name.unpack(p.msg, p.off)
	if err != nil {
		return Question{}, &nestedError{"unpacking Question.Name", err}
	}
	typ, off, err := unpackType(p.msg, off)
	if err != nil {
		return Question{}, &nestedError{"unpacking Question.Type", err}
	}
	class, off, err := unpackClass(p.msg, off)
	if err != nil {
		return Question{}, &nestedError{"unpacking Question.Class", err}
	}
	p.off = off
	p.index++
	return Question{name, typ, class}, nil
}

// AllQuestions parses all Questions.
func (p *Parser) AllQuestions() ([]Question, error) {
	qs := make([]Question, 0, p.header.questions)
	for {
		q, err := p.Question()
		if err == ErrSectionDone {
			return qs, nil
		}
		if err != nil {
			return nil, err
		}
		qs = append(qs, q)
	}
}

// SkipQuestion skips a single Question.
func (p *Parser) SkipQuestion() error {
	if err := p.checkAdvance(sectionQuestions); err != nil {
		return err
	}
	off, err := skipName(p.msg, p.off)
	if err != nil {
		return &nestedError{"skipping Question Name", err}
	}
	if off, err = skipType(p.msg, off); err != nil {
		return &nestedError{"skipping Question Type", err}
	}
	if off, err = skipClass(p.msg, off); err != nil {
		return &nestedError{"skipping Question Class", err}
	}
	p.off = off
	p.index++
	return nil
}

// SkipAllQuestions skips all Questions.
func (p *Parser) SkipAllQuestions() error {
	for {
		if err := p.SkipQuestion(); err == ErrSectionDone {
			return nil
		} else if err != nil {
			return err
		}
	}
}

// AnswerHeader parses a single Answer ResourceHeader.
func (p *Parser) AnswerHeader() (ResourceHeader, error) {
	return p.resourceHeader(sectionAnswers)
}

// Answer parses a single Answer Resource.
func (p *Parser) Answer() (Resource, error) {
	return p.resource(sectionAnswers)
}

// AllAnswers parses all Answer Resources.
func (p *Parser) AllAnswers() ([]Resource, error) {
	as := make([]Resource, 0, p.header.answers)
	for {
		a, err := p.Answer()
		if err == ErrSectionDone {
			return as, nil
		}
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
}

// SkipAnswer skips a single Answer Resource.
func (p *Parser) SkipAnswer() error {
	return p.skipResource(sectionAnswers)
}

// SkipAllAnswers skips all Answer Resources.
func (p *Parser) SkipAllAnswers() error {
	for {
		if err := p.SkipAnswer(); err == ErrSectionDone {
			return nil
		} else if err != nil {
			return err
		}
	}
}

// AuthorityHeader parses a single Authority ResourceHeader.
func (p *Parser) AuthorityHeader() (ResourceHeader, error) {
	return p.resourceHeader(sectionAuthorities)
}

// Authority parses a single Authority Resource.
func (p *Parser) Authority() (Resource, error) {
	return p.resource(sectionAuthorities)
}

// AllAuthorities parses all Authority Resources.
func (p *Parser) AllAuthorities() ([]Resource, error) {
	as := make([]Resource, 0, p.header.authorities)
	for {
		a, err := p.Authority()
		if err == ErrSectionDone {
			return as, nil
		}
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
}

// SkipAuthority skips a single Authority Resource.
func (p *Parser) SkipAuthority() error {
	return p.skipResource(sectionAuthorities)
}

// SkipAllAuthorities skips all Authority Resources.
func (p *Parser) SkipAllAuthorities() error {
	for {
		if err := p.SkipAuthority(); err == ErrSectionDone {
			return nil
		} else if err != nil {
			return err
		}
	}
}

// AdditionalHeader parses a single Additional ResourceHeader.
func (p *Parser) AdditionalHeader() (ResourceHeader, error) {
	return p.resourceHeader(sectionAdditionals)
}

// Additional parses a single Additional Resource.
func (p *Parser) Additional() (Resource, error) {
	return p.resource(sectionAdditionals)
}

// AllAdditionals parses all Additional Resources.
func (p *Parser) AllAdditionals() ([]Resource, error) {
	as := make([]Resource, 0, p.header.additionals)
	for {
		a, err := p.Additional()
		if err == ErrSectionDone {
			return as, nil
		}
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
}

// SkipAdditional skips a single Additional Resource.
func (p *Parser) SkipAdditional() error {
	return p.skipResource(sectionAdditionals)
}

// SkipAllAdditionals skips all Additional Resources.
func (p *Parser) SkipAllAdditionals() error {
	for {
		if err := p.SkipAdditional(); err == ErrSectionDone {
			return nil
		} else if err != nil {
			return err
		}
	}
}

// CNAMEResource parses a single CNAMEResource.
//
// One of the XXXHeader methods must have been called before calling this
// method.
func (p *Parser) CNAMEResource() (CNAMEResource, error) {
	if !p.resHeaderValid || p.resHeader.Type != TypeCNAME {
		return CNAMEResource{}, ErrNotStarted
	}
	r, err := unpackCNAMEResource(p.msg, p.off)
	if err != nil {
		return CNAMEResource{}, err
	}
	p.off += int(p.resHeader.Length)
	p.resHeaderValid = false
	p.index++
	return r, nil
}

// MXResource parses a single MXResource.
//
// One of the XXXHeader methods must have been called before calling this
// method.
func (p *Parser) MXResource() (MXResource, error) {
	if !p.resHeaderValid || p.resHeader.Type != TypeMX {
		return MXResource{}, ErrNotStarted
	}
	r, err := unpackMXResource(p.msg, p.off)
	if err != nil {
		return MXResource{}, err
	}
	p.off += int(p.resHeader.Length)
	p.resHeaderValid = false
	p.index++
	return r, nil
}

// NSResource parses a single NSResource.
//
// One of the XXXHeader methods must have been called before calling this
// method.
func (p *Parser) NSResource() (NSResource, error) {
	if !p.resHeaderValid || p.resHeader.Type != TypeNS {
		return NSResource{}, ErrNotStarted
	}
	r, err := unpackNSResource(p.msg, p.off)
	if err != nil {
		return NSResource{}, err
	}
	p.off += int(p.resHeader.Length)
	p.resHeaderValid = false
	p.index++
	return r, nil
}

// PTRResource parses a single PTRResource.
//
// One of the XXXHeader methods must have been called before calling this
// method.
func (p *Parser) PTRResource() (PTRResource, error) {
	if !p.resHeaderValid || p.resHeader.Type != TypePTR {
		return PTRResource{}, ErrNotStarted
	}
	r, err := unpackPTRResource(p.msg, p.off)
	if err != nil {
		return PTRResource{}, err
	}
	p.off += int(p.resHeader.Length)
	p.resHeaderValid = false
	p.index++
	return r, nil
}

// SOAResource parses a single SOAResource.
//
// One of the XXXHeader methods must have been called before calling this
// method.
func (p *Parser) SOAResource() (SOAResource, error) {
	if !p.resHeaderValid || p.resHeader.Type != TypeSOA {
		return SOAResource{}, ErrNotStarted
	}
	r, err := unpackSOAResource(p.msg, p.off)
	if err != nil {
		return SOAResource{}, err
	}
	p.off += int(p.resHeader.Length)
	p.resHeaderValid = false
	p.index++
	return r, nil
}

// TXTResource parses a single TXTResource.
//
// One of the XXXHeader methods must have been called before calling this
// method.
func (p *Parser) TXTResource() (TXTResource, error) {
	if !p.resHeaderValid || p.resHeader.Type != TypeTXT {
		return TXTResource{}, ErrNotStarted
	}
	r, err := unpackTXTResource(p.msg, p.off, p.resHeader.Length)
	if err != nil {
		return TXTResource{}, err
	}
	p.off += int(p.resHeader.Length)
	p.resHeaderValid = false
	p.index++
	return r, nil
}

// SRVResource parses a single SRVResource.
//
// One of the XXXHeader methods must have been called before calling this
// method.
func (p *Parser) SRVResource() (SRVResource, error) {
	if !p.resHeaderValid || p.resHeader.Type != TypeSRV {
		return SRVResource{}, ErrNotStarted
	}
	r, err := unpackSRVResource(p.msg, p.off)
	if err != nil {
		return SRVResource{}, err
	}
	p.off += int(p.resHeader.Length)
	p.resHeaderValid = false
	p.index++
	return r, nil
}

// AResource parses a single AResource.
//
// One of the XXXHeader methods must have been called before calling this
// method.
func (p *Parser) AResource() (AResource, error) {
	if !p.resHeaderValid || p.resHeader.Type != TypeA {
		return AResource{}, ErrNotStarted
	}
	r, err := unpackAResource(p.msg, p.off)
	if err != nil {
		return AResource{}, err
	}
	p.off += int(p.resHeader.Length)
	p.resHeaderValid = false
	p.index++
	return r, nil
}

// AAAAResource parses a single AAAAResource.
//
// One of the XXXHeader methods must have been called before calling this
// method.
func (p *Parser) AAAAResource() (AAAAResource, error) {
	if !p.resHeaderValid || p.resHeader.Type != TypeAAAA {
		return AAAAResource{}, ErrNotStarted
	}
	r, err := unpackAAAAResource(p.msg, p.off)
	if err != nil {
		return AAAAResource{}, err
	}
	p.off += int(p.resHeader.Length)
	p.resHeaderValid = false
	p.index++
	return r, nil
}

// Unpack parses a full Message.
func (m *Message) Unpack(msg []byte) error {
	var p Parser
	var err error
	if m.Header, err = p.Start(msg); err != nil {
		return err
	}
	if m.Questions, err = p.AllQuestions(); err != nil {
		return err
	}
	if m.Answers, err = p.AllAnswers(); err != nil {
		return err
	}
	if m.Authorities, err = p.AllAuthorities(); err != nil {
		return err
	}
	if m.Additionals, err = p.AllAdditionals(); err != nil {
		return err
	}
	return nil
}

// Pack packs a full Message.
func (m *Message) Pack() ([]byte, error) {
	return m.AppendPack(make([]byte, 0, packStartingCap))
}

// AppendPack is like Pack but appends the full Message to b and returns the
// extended buffer.
func (m *Message) AppendPack(b []byte) ([]byte, error) {
	// Validate the lengths. It is very unlikely that anyone will try to
	// pack more than 65535 of any particular type, but it is possible and
	// we should fail gracefully.
	if len(m.Questions) > int(^uint16(0)) {
		return nil, errTooManyQuestions
	}
	if len(m.Answers) > int(^uint16(0)) {
		return nil, errTooManyAnswers
	}
	if len(m.Authorities) > int(^uint16(0)) {
		return nil, errTooManyAuthorities
	}
	if len(m.Additionals) > int(^uint16(0)) {
		return nil, errTooManyAdditionals
	}

	var h header
	h.id, h.bits = m.Header.pack()

	h.questions = uint16(len(m.Questions))
	h.answers = uint16(len(m.Answers))
	h.authorities = uint16(len(m.Authorities))
	h.additionals = uint16(len(m.Additionals))

	msg := h.pack(b)

	// RFC 1035 allows (but does not require) compression for packing. RFC
	// 1035 requires unpacking implementations to support compression, so
	// unconditionally enabling it is fine.
	//
	// DNS lookups are typically done over UDP, and RFC 1035 states that UDP
	// DNS packets can be a maximum of 512 bytes long. Without compression,
	// many DNS response packets are over this limit, so enabling
	// compression will help ensure compliance.
	compression := map[string]int{}

	for i := range m.Questions {
		var err error
		if msg, err = m.Questions[i].pack(msg, compression); err != nil {
			return nil, &nestedError{"packing Question", err}
		}
	}
	for i := range m.Answers {
		var err error
		if msg, err = m.Answers[i].pack(msg, compression); err != nil {
			return nil, &nestedError{"packing Answer", err}
		}
	}
	for i := range m.Authorities {
		var err error
		if msg, err = m.Authorities[i].pack(msg, compression); err != nil {
			return nil, &nestedError{"packing Authority", err}
		}
	}
	for i := range m.Additionals {
		var err error
		if msg, err = m.Additionals[i].pack(msg, compression); err != nil {
			return nil, &nestedError{"packing Additional", err}
		}
	}

	return msg, nil
}

// A Builder allows incrementally packing a DNS message.
type Builder struct {
	msg         []byte
	header      header
	section     section
	compression map[string]int
}

// Start initializes the builder.
//
// buf is optional (nil is fine), but if provided, Start takes ownership of buf.
func (b *Builder) Start(buf []byte, h Header) {
	b.StartWithoutCompression(buf, h)
	b.compression = map[string]int{}
}

// StartWithoutCompression initializes the builder with compression disabled.
//
// This avoids compression related allocations, but can result in larger message
// sizes. Be careful with this mode as it can cause messages to exceed the UDP
// size limit.
//
// buf is optional (nil is fine), but if provided, Start takes ownership of buf.
func (b *Builder) StartWithoutCompression(buf []byte, h Header) {
	*b = Builder{msg: buf}
	b.header.id, b.header.bits = h.pack()
	if cap(b.msg) < headerLen {
		b.msg = make([]byte, 0, packStartingCap)
	}
	b.msg = b.msg[:headerLen]
	b.section = sectionHeader
}

func (b *Builder) startCheck(s section) error {
	if b.section <= sectionNotStarted {
		return ErrNotStarted
	}
	if b.section > s {
		return ErrSectionDone
	}
	return nil
}

// StartQuestions prepares the builder for packing Questions.
func (b *Builder) StartQuestions() error {
	if err := b.startCheck(sectionQuestions); err != nil {
		return err
	}
	b.section = sectionQuestions
	return nil
}

// StartAnswers prepares the builder for packing Answers.
func (b *Builder) StartAnswers() error {
	if err := b.startCheck(sectionAnswers); err != nil {
		return err
	}
	b.section = sectionAnswers
	return nil
}

// StartAuthorities prepares the builder for packing Authorities.
func (b *Builder) StartAuthorities() error {
	if err := b.startCheck(sectionAuthorities); err != nil {
		return err
	}
	b.section = sectionAuthorities
	return nil
}

// StartAdditionals prepares the builder for packing Additionals.
func (b *Builder) StartAdditionals() error {
	if err := b.startCheck(sectionAdditionals); err != nil {
		return err
	}
	b.section = sectionAdditionals
	return nil
}

func (b *Builder) incrementSectionCount() error {
	var count *uint16
	var err error
	switch b.section {
	case sectionQuestions:
		count = &b.header.questions
		err = errTooManyQuestions
	case sectionAnswers:
		count = &b.header.answers
		err = errTooManyAnswers
	case sectionAuthorities:
		count = &b.header.authorities
		err = errTooManyAuthorities
	case sectionAdditionals:
		count = &b.header.additionals
		err = errTooManyAdditionals
	}
	if *count == ^uint16(0) {
		return err
	}
	*count++
	return nil
}

// Question adds a single Question.
func (b *Builder) Question(q Question) error {
	if b.section < sectionQuestions {
		return ErrNotStarted
	}
	if b.section > sectionQuestions {
		return ErrSectionDone
	}
	msg, err := q.pack(b.msg, b.compression)
	if err != nil {
		return err
	}
	if err := b.incrementSectionCount(); err != nil {
		return err
	}
	b.msg = msg
	return nil
}

func (b *Builder) checkResourceSection() error {
	if b.section < sectionAnswers {
		return ErrNotStarted
	}
	if b.section > sectionAdditionals {
		return ErrSectionDone
	}
	return nil
}

// CNAMEResource adds a single CNAMEResource.
func (b *Builder) CNAMEResource(h ResourceHeader, r CNAMEResource) error {
	if err := b.checkResourceSection(); err != nil {
		return err
	}
	h.Type = r.realType()
	msg, length, err := h.pack(b.msg, b.compression)
	if err != nil {
		return &nestedError{"ResourceHeader", err}
	}
	preLen := len(msg)
	if msg, err = r.pack(msg, b.compression); err != nil {
		return &nestedError{"CNAMEResource body", err}
	}
	if err := h.fixLen(msg, length, preLen); err != nil {
		return err
	}
	if err := b.incrementSectionCount(); err != nil {
		return err
	}
	b.msg = msg
	return nil
}

// MXResource adds a single MXResource.
func (b *Builder) MXResource(h ResourceHeader, r MXResource) error {
	if err := b.checkResourceSection(); err != nil {
		return err
	}
	h.Type = r.realType()
	msg, length, err := h.pack(b.msg, b.compression)
	if err != nil {
		return &nestedError{"ResourceHeader", err}
	}
	preLen := len(msg)
	if msg, err = r.pack(msg, b.compression); err != nil {
		return &nestedError{"MXResource body", err}
	}
	if err := h.fixLen(msg, length, preLen); err != nil {
		return err
	}
	if err := b.incrementSectionCount(); err != nil {
		return err
	}
	b.msg = msg
	return nil
}

// NSResource adds a single NSResource.
func (b *Builder) NSResource(h ResourceHeader, r NSResource) error {
	if err := b.checkResourceSection(); err != nil {
		return err
	}
	h.Type = r.realType()
	msg, length, err := h.pack(b.msg, b.compression)
	if err != nil {
		return &nestedError{"ResourceHeader", err}
	}
	preLen := len(msg)
	if msg, err = r.pack(msg, b.compression); err != nil {
		return &nestedError{"NSResource body", err}
	}
	if err := h.fixLen(msg, length, preLen); err != nil {
		return err
	}
	if err := b.incrementSectionCount(); err != nil {
		return err
	}
	b.msg = msg
	return nil
}

// PTRResource adds a single PTRResource.
func (b *Builder) PTRResource(h ResourceHeader, r PTRResource) error {
	if err := b.checkResourceSection(); err != nil {
		return err
	}
	h.Type = r.realType()
	msg, length, err := h.pack(b.msg, b.compression)
	if err != nil {
		return &nestedError{"ResourceHeader", err}
	}
	preLen := len(msg)
	if msg, err = r.pack(msg, b.compression); err != nil {
		return &nestedError{"PTRResource body", err}
	}
	if err := h.fixLen(msg, length, preLen); err != nil {
		return err
	}
	if err := b.incrementSectionCount(); err != nil {
		return err
	}
	b.msg = msg
	return nil
}

// SOAResource adds a single SOAResource.
func (b *Builder) SOAResource(h ResourceHeader, r SOAResource) error {
	if err := b.checkResourceSection(); err != nil {
		return err
	}
	h.Type = r.realType()
	msg, length, err := h.pack(b.msg, b.compression)
	if err != nil {
		return &nestedError{"ResourceHeader", err}
	}
	preLen := len(msg)
	if msg, err = r.pack(msg, b.compression); err != nil {
		return &nestedError{"SOAResource body", err}
	}
	if err := h.fixLen(msg, length, preLen); err != nil {
		return err
	}
	if err := b.incrementSectionCount(); err != nil {
		return err
	}
	b.msg = msg
	return nil
}

// TXTResource adds a single TXTResource.
func (b *Builder) TXTResource(h ResourceHeader, r TXTResource) error {
	if err := b.checkResourceSection(); err != nil {
		return err
	}
	h.Type = r.realType()
	msg, length, err := h.pack(b.msg, b.compression)
	if err != nil {
		return &nestedError{"ResourceHeader", err}
	}
	preLen := len(msg)
	if msg, err = r.pack(msg, b.compression); err != nil {
		return &nestedError{"TXTResource body", err}
	}
	if err := h.fixLen(msg, length, preLen); err != nil {
		return err
	}
	if err := b.incrementSectionCount(); err != nil {
		return err
	}
	b.msg = msg
	return nil
}

// SRVResource adds a single SRVResource.
func (b *Builder) SRVResource(h ResourceHeader, r SRVResource) error {
	if err := b.checkResourceSection(); err != nil {
		return err
	}
	h.Type = r.realType()
	msg, length, err := h.pack(b.msg, b.compression)
	if err != nil {
		return &nestedError{"ResourceHeader", err}
	}
	preLen := len(msg)
	if msg, err = r.pack(msg, b.compression); err != nil {
		return &nestedError{"SRVResource body", err}
	}
	if err := h.fixLen(msg, length, preLen); err != nil {
		return err
	}
	if err := b.incrementSectionCount(); err != nil {
		return err
	}
	b.msg = msg
	return nil
}

// AResource adds a single AResource.
func (b *Builder) AResource(h ResourceHeader, r AResource) error {
	if err := b.checkResourceSection(); err != nil {
		return err
	}
	h.Type = r.realType()
	msg, length, err := h.pack(b.msg, b.compression)
	if err != nil {
		return &nestedError{"ResourceHeader", err}
	}
	preLen := len(msg)
	if msg, err = r.pack(msg, b.compression); err != nil {
		return &nestedError{"AResource body", err}
	}
	if err := h.fixLen(msg, length, preLen); err != nil {
		return err
	}
	if err := b.incrementSectionCount(); err != nil {
		return err
	}
	b.msg = msg
	return nil
}

// AAAAResource adds a single AAAAResource.
func (b *Builder) AAAAResource(h ResourceHeader, r AAAAResource) error {
	if err := b.checkResourceSection(); err != nil {
		return err
	}
	h.Type = r.realType()
	msg, length, err := h.pack(b.msg, b.compression)
	if err != nil {
		return &nestedError{"ResourceHeader", err}
	}
	preLen := len(msg)
	if msg, err = r.pack(msg, b.compression); err != nil {
		return &nestedError{"AAAAResource body", err}
	}
	if err := h.fixLen(msg, length, preLen); err != nil {
		return err
	}
	if err := b.incrementSectionCount(); err != nil {
		return err
	}
	b.msg = msg
	return nil
}

// Finish ends message building and generates a binary packet.
func (b *Builder) Finish() ([]byte, error) {
	if b.section < sectionHeader {
		return nil, ErrNotStarted
	}
	b.section = sectionDone
	b.header.pack(b.msg[:0])
	return b.msg, nil
}

// A ResourceHeader is the header of a DNS resource record. There are
// many types of DNS resource records, but they all share the same header.
type ResourceHeader struct {
	// Name is the domain name for which this resource record pertains.
	Name Name

	// Type is the type of DNS resource record.
	//
	// This field will be set automatically during packing.
	Type Type

	// Class is the class of network to which this DNS resource record
	// pertains.
	Class Class

	// TTL is the length of time (measured in seconds) which this resource
	// record is valid for (time to live). All Resources in a set should
	// have the same TTL (RFC 2181 Section 5.2).
	TTL uint32

	// Length is the length of data in the resource record after the header.
	//
	// This field will be set automatically during packing.
	Length uint16
}

// pack packs all of the fields in a ResourceHeader except for the length. The
// length bytes are returned as a slice so they can be filled in after the rest
// of the Resource has been packed.
func (h *ResourceHeader) pack(oldMsg []byte, compression map[string]int) (msg []byte, length []byte, err error) {
	msg = oldMsg
	if msg, err = h.Name.pack(msg, compression); err != nil {
		return oldMsg, nil, &nestedError{"Name", err}
	}
	msg = packType(msg, h.Type)
	msg = packClass(msg, h.Class)
	msg = packUint32(msg, h.TTL)
	lenBegin := len(msg)
	msg = packUint16(msg, h.Length)
	return msg, msg[lenBegin : lenBegin+uint16Len], nil
}

func (h *ResourceHeader) unpack(msg []byte, off int) (int, error) {
	newOff := off
	var err error
	if newOff, err = h.Name.unpack(msg, newOff); err != nil {
		return off, &nestedError{"Name", err}
	}
	if h.Type, newOff, err = unpackType(msg, newOff); err != nil {
		return off, &nestedError{"Type", err}
	}
	if h.Class, newOff, err = unpackClass(msg, newOff); err != nil {
		return off, &nestedError{"Class", err}
	}
	if h.TTL, newOff, err = unpackUint32(msg, newOff); err != nil {
		return off, &nestedError{"TTL", err}
	}
	if h.Length, newOff, err = unpackUint16(msg, newOff); err != nil {
		return off, &nestedError{"Length", err}
	}
	return newOff, nil
}

func (h *ResourceHeader) fixLen(msg []byte, length []byte, preLen int) error {
	conLen := len(msg) - preLen
	if conLen > int(^uint16(0)) {
		return errResTooLong
	}

	// Fill in the length now that we know how long the content is.
	packUint16(length[:0], uint16(conLen))
	h.Length = uint16(conLen)

	return nil
}

func skipResource(msg []byte, off int) (int, error) {
	newOff, err := skipName(msg, off)
	if err != nil {
		return off, &nestedError{"Name", err}
	}
	if newOff, err = skipType(msg, newOff); err != nil {
		return off, &nestedError{"Type", err}
	}
	if newOff, err = skipClass(msg, newOff); err != nil {
		return off, &nestedError{"Class", err}
	}
	if newOff, err = skipUint32(msg, newOff); err != nil {
		return off, &nestedError{"TTL", err}
	}
	length, newOff, err := unpackUint16(msg, newOff)
	if err != nil {
		return off, &nestedError{"Length", err}
	}
	if newOff += int(length); newOff > len(msg) {
		return off, errResourceLen
	}
	return newOff, nil
}

func packUint16(msg []byte, field uint16) []byte {
	return append(msg, byte(field>>8), byte(field))
}

func unpackUint16(msg []byte, off int) (uint16, int, error) {
	if off+uint16Len > len(msg) {
		return 0, off, errBaseLen
	}
	return uint16(msg[off])<<8 | uint16(msg[off+1]), off + uint16Len, nil
}

func skipUint16(msg []byte, off int) (int, error) {
	if off+uint16Len > len(msg) {
		return off, errBaseLen
	}
	return off + uint16Len, nil
}

func packType(msg []byte, field Type) []byte {
	return packUint16(msg, uint16(field))
}

func unpackType(msg []byte, off int) (Type, int, error) {
	t, o, err := unpackUint16(msg, off)
	return Type(t), o, err
}

func skipType(msg []byte, off int) (int, error) {
	return skipUint16(msg, off)
}

func packClass(msg []byte, field Class) []byte {
	return packUint16(msg, uint16(field))
}

func unpackClass(msg []byte, off int) (Class, int, error) {
	c, o, err := unpackUint16(msg, off)
	return Class(c), o, err
}

func skipClass(msg []byte, off int) (int, error) {
	return skipUint16(msg, off)
}

func packUint32(msg []byte, field uint32) []byte {
	return append(
		msg,
		byte(field>>24),
		byte(field>>16),
		byte(field>>8),
		byte(field),
	)
}

func unpackUint32(msg []byte, off int) (uint32, int, error) {
	if off+uint32Len > len(msg) {
		return 0, off, errBaseLen
	}
	v := uint32(msg[off])<<24 | uint32(msg[off+1])<<16 | uint32(msg[off+2])<<8 | uint32(msg[off+3])
	return v, off + uint32Len, nil
}

func skipUint32(msg []byte, off int) (int, error) {
	if off+uint32Len > len(msg) {
		return off, errBaseLen
	}
	return off + uint32Len, nil
}

func packText(msg []byte, field string) []byte {
	for len(field) > 0 {
		l := len(field)
		if l > 255 {
			l = 255
		}
		msg = append(msg, byte(l))
		msg = append(msg, field[:l]...)
		field = field[l:]
	}
	return msg
}

func unpackText(msg []byte, off int) (string, int, error) {
	if off >= len(msg) {
		return "", off, errBaseLen
	}
	beginOff := off + 1
	endOff := beginOff + int(msg[off])
	if endOff > len(msg) {
		return "", off, errCalcLen
	}
	return string(msg[beginOff:endOff]), endOff, nil
}

func skipText(msg []byte, off int) (int, error) {
	if off >= len(msg) {
		return off, errBaseLen
	}
	endOff := off + 1 + int(msg[off])
	if endOff > len(msg) {
		return off, errCalcLen
	}
	return endOff, nil
}

func packBytes(msg []byte, field []byte) []byte {
	return append(msg, field...)
}

func unpackBytes(msg []byte, off int, field []byte) (int, error) {
	newOff := off + len(field)
	if newOff > len(msg) {
		return off, errBaseLen
	}
	copy(field, msg[off:newOff])
	return newOff, nil
}

func skipBytes(msg []byte, off int, field []byte) (int, error) {
	newOff := off + len(field)
	if newOff > len(msg) {
		return off, errBaseLen
	}
	return newOff, nil
}

const nameLen = 255

// A Name is a non-encoded domain name. It is used instead of strings to avoid
// allocations.
type Name struct {
	Data   [nameLen]byte
	Length uint8
}

// NewName creates a new Name from a string.
func NewName(name string) (Name, error) {
	if len([]byte(name)) > nameLen {
		return Name{}, errCalcLen
	}
	n := Name{Length: uint8(len(name))}
	copy(n.Data[:], []byte(name))
	return n, nil
}

func (n Name) String() string {
	return string(n.Data[:n.Length])
}

// pack packs a domain name.
//
// Domain names are a sequence of counted strings split at the dots. They end
// with a zero-length string. Compression can be used to reuse domain suffixes.
//
// The compression map will be updated with new domain suffixes. If compression
// is nil, compression will not be used.
func (n *Name) pack(msg []byte, compression map[string]int) ([]byte, error) {
	oldMsg := msg

	// Add a trailing dot to canonicalize name.
	if n.Length == 0 || n.Data[n.Length-1] != '.' {
		return oldMsg, errNonCanonicalName
	}

	// Allow root domain.
	if n.Data[0] == '.' && n.Length == 1 {
		return append(msg, 0), nil
	}

	// Emit sequence of counted strings, chopping at dots.
	for i, begin := 0, 0; i < int(n.Length); i++ {
		// Check for the end of the segment.
		if n.Data[i] == '.' {
			// The two most significant bits have special meaning.
			// It isn't allowed for segments to be long enough to
			// need them.
			if i-begin >= 1<<6 {
				return oldMsg, errSegTooLong
			}

			// Segments must have a non-zero length.
			if i-begin == 0 {
				return oldMsg, errZeroSegLen
			}

			msg = append(msg, byte(i-begin))

			for j := begin; j < i; j++ {
				msg = append(msg, n.Data[j])
			}

			begin = i + 1
			continue
		}

		// We can only compress domain suffixes starting with a new
		// segment. A pointer is two bytes with the two most significant
		// bits set to 1 to indicate that it is a pointer.
		if (i == 0 || n.Data[i-1] == '.') && compression != nil {
			if ptr, ok := compression[string(n.Data[i:])]; ok {
				// Hit. Emit a pointer instead of the rest of
				// the domain.
				return append(msg, byte(ptr>>8|0xC0), byte(ptr)), nil
			}

			// Miss. Add the suffix to the compression table if the
			// offset can be stored in the available 14 bytes.
			if len(msg) <= int(^uint16(0)>>2) {
				compression[string(n.Data[i:])] = len(msg)
			}
		}
	}
	return append(msg, 0), nil
}

// unpack unpacks a domain name.
func (n *Name) unpack(msg []byte, off int) (int, error) {
	// currOff is the current working offset.
	currOff := off

	// newOff is the offset where the next record will start. Pointers lead
	// to data that belongs to other names and thus doesn't count towards to
	// the usage of this name.
	newOff := off

	// ptr is the number of pointers followed.
	var ptr int

	// Name is a slice representation of the name data.
	name := n.Data[:0]

Loop:
	for {
		if currOff >= len(msg) {
			return off, errBaseLen
		}
		c := int(msg[currOff])
		currOff++
		switch c & 0xC0 {
		case 0x00: // String segment
			if c == 0x00 {
				// A zero length signals the end of the name.
				break Loop
			}
			endOff := currOff + c
			if endOff > len(msg) {
				return off, errCalcLen
			}
			name = append(name, msg[currOff:endOff]...)
			name = append(name, '.')
			currOff = endOff
		case 0xC0: // Pointer
			if currOff >= len(msg) {
				return off, errInvalidPtr
			}
			c1 := msg[currOff]
			currOff++
			if ptr == 0 {
				newOff = currOff
			}
			// Don't follow too many pointers, maybe there's a loop.
			if ptr++; ptr > 10 {
				return off, errTooManyPtr
			}
			currOff = (c^0xC0)<<8 | int(c1)
		default:
			// Prefixes 0x80 and 0x40 are reserved.
			return off, errReserved
		}
	}
	if len(name) == 0 {
		name = append(name, '.')
	}
	if len(name) > len(n.Data) {
		return off, errCalcLen
	}
	n.Length = uint8(len(name))
	if ptr == 0 {
		newOff = currOff
	}
	return newOff, nil
}

func skipName(msg []byte, off int) (int, error) {
	// newOff is the offset where the next record will start. Pointers lead
	// to data that belongs to other names and thus doesn't count towards to
	// the usage of this name.
	newOff := off

Loop:
	for {
		if newOff >= len(msg) {
			return off, errBaseLen
		}
		c := int(msg[newOff])
		newOff++
		switch c & 0xC0 {
		case 0x00:
			if c == 0x00 {
				// A zero length signals the end of the name.
				break Loop
			}
			// literal string
			newOff += c
			if newOff > len(msg) {
				return off, errCalcLen
			}
		case 0xC0:
			// Pointer to somewhere else in msg.

			// Pointers are two bytes.
			newOff++

			// Don't follow the pointer as the data here has ended.
			break Loop
		default:
			// Prefixes 0x80 and 0x40 are reserved.
			return off, errReserved
		}
	}

	return newOff, nil
}

// A Question is a DNS query.
type Question struct {
	Name  Name
	Type  Type
	Class Class
}

func (q *Question) pack(msg []byte, compression map[string]int) ([]byte, error) {
	msg, err := q.Name.pack(msg, compression)
	if err != nil {
		return msg, &nestedError{"Name", err}
	}
	msg = packType(msg, q.Type)
	return packClass(msg, q.Class), nil
}

func unpackResourceBody(msg []byte, off int, hdr ResourceHeader) (ResourceBody, int, error) {
	var (
		r    ResourceBody
		err  error
		name string
	)
	switch hdr.Type {
	case TypeA:
		var rb AResource
		rb, err = unpackAResource(msg, off)
		r = &rb
		name = "A"
	case TypeNS:
		var rb NSResource
		rb, err = unpackNSResource(msg, off)
		r = &rb
		name = "NS"
	case TypeCNAME:
		var rb CNAMEResource
		rb, err = unpackCNAMEResource(msg, off)
		r = &rb
		name = "CNAME"
	case TypeSOA:
		var rb SOAResource
		rb, err = unpackSOAResource(msg, off)
		r = &rb
		name = "SOA"
	case TypePTR:
		var rb PTRResource
		rb, err = unpackPTRResource(msg, off)
		r = &rb
		name = "PTR"
	case TypeMX:
		var rb MXResource
		rb, err = unpackMXResource(msg, off)
		r = &rb
		name = "MX"
	case TypeTXT:
		var rb TXTResource
		rb, err = unpackTXTResource(msg, off, hdr.Length)
		r = &rb
		name = "TXT"
	case TypeAAAA:
		var rb AAAAResource
		rb, err = unpackAAAAResource(msg, off)
		r = &rb
		name = "AAAA"
	case TypeSRV:
		var rb SRVResource
		rb, err = unpackSRVResource(msg, off)
		r = &rb
		name = "SRV"
	}
	if err != nil {
		return nil, off, &nestedError{name + " record", err}
	}
	if r == nil {
		return nil, off, errors.New("invalid resource type: " + string(hdr.Type+'0'))
	}
	return r, off + int(hdr.Length), nil
}

// A CNAMEResource is a CNAME Resource record.
type CNAMEResource struct {
	CNAME Name
}

func (r *CNAMEResource) realType() Type {
	return TypeCNAME
}

func (r *CNAMEResource) pack(msg []byte, compression map[string]int) ([]byte, error) {
	return r.CNAME.pack(msg, compression)
}

func unpackCNAMEResource(msg []byte, off int) (CNAMEResource, error) {
	var cname Name
	if _, err := cname.unpack(msg, off); err != nil {
		return CNAMEResource{}, err
	}
	return CNAMEResource{cname}, nil
}

// An MXResource is an MX Resource record.
type MXResource struct {
	Pref uint16
	MX   Name
}

func (r *MXResource) realType() Type {
	return TypeMX
}

func (r *MXResource) pack(msg []byte, compression map[string]int) ([]byte, error) {
	oldMsg := msg
	msg = packUint16(msg, r.Pref)
	msg, err := r.MX.pack(msg, compression)
	if err != nil {
		return oldMsg, &nestedError{"MXResource.MX", err}
	}
	return msg, nil
}

func unpackMXResource(msg []byte, off int) (MXResource, error) {
	pref, off, err := unpackUint16(msg, off)
	if err != nil {
		return MXResource{}, &nestedError{"Pref", err}
	}
	var mx Name
	if _, err := mx.unpack(msg, off); err != nil {
		return MXResource{}, &nestedError{"MX", err}
	}
	return MXResource{pref, mx}, nil
}

// An NSResource is an NS Resource record.
type NSResource struct {
	NS Name
}

func (r *NSResource) realType() Type {
	return TypeNS
}

func (r *NSResource) pack(msg []byte, compression map[string]int) ([]byte, error) {
	return r.NS.pack(msg, compression)
}

func unpackNSResource(msg []byte, off int) (NSResource, error) {
	var ns Name
	if _, err := ns.unpack(msg, off); err != nil {
		return NSResource{}, err
	}
	return NSResource{ns}, nil
}

// A PTRResource is a PTR Resource record.
type PTRResource struct {
	PTR Name
}

func (r *PTRResource) realType() Type {
	return TypePTR
}

func (r *PTRResource) pack(msg []byte, compression map[string]int) ([]byte, error) {
	return r.PTR.pack(msg, compression)
}

func unpackPTRResource(msg []byte, off int) (PTRResource, error) {
	var ptr Name
	if _, err := ptr.unpack(msg, off); err != nil {
		return PTRResource{}, err
	}
	return PTRResource{ptr}, nil
}

// An SOAResource is an SOA Resource record.
type SOAResource struct {
	NS      Name
	MBox    Name
	Serial  uint32
	Refresh uint32
	Retry   uint32
	Expire  uint32

	// MinTTL the is the default TTL of Resources records which did not
	// contain a TTL value and the TTL of negative responses. (RFC 2308
	// Section 4)
	MinTTL uint32
}

func (r *SOAResource) realType() Type {
	return TypeSOA
}

func (r *SOAResource) pack(msg []byte, compression map[string]int) ([]byte, error) {
	oldMsg := msg
	msg, err := r.NS.pack(msg, compression)
	if err != nil {
		return oldMsg, &nestedError{"SOAResource.NS", err}
	}
	msg, err = r.MBox.pack(msg, compression)
	if err != nil {
		return oldMsg, &nestedError{"SOAResource.MBox", err}
	}
	msg = packUint32(msg, r.Serial)
	msg = packUint32(msg, r.Refresh)
	msg = packUint32(msg, r.Retry)
	msg = packUint32(msg, r.Expire)
	return packUint32(msg, r.MinTTL), nil
}

func unpackSOAResource(msg []byte, off int) (SOAResource, error) {
	var ns Name
	off, err := ns.unpack(msg, off)
	if err != nil {
		return SOAResource{}, &nestedError{"NS", err}
	}
	var mbox Name
	if off, err = mbox.unpack(msg, off); err != nil {
		return SOAResource{}, &nestedError{"MBox", err}
	}
	serial, off, err := unpackUint32(msg, off)
	if err != nil {
		return SOAResource{}, &nestedError{"Serial", err}
	}
	refresh, off, err := unpackUint32(msg, off)
	if err != nil {
		return SOAResource{}, &nestedError{"Refresh", err}
	}
	retry, off, err := unpackUint32(msg, off)
	if err != nil {
		return SOAResource{}, &nestedError{"Retry", err}
	}
	expire, off, err := unpackUint32(msg, off)
	if err != nil {
		return SOAResource{}, &nestedError{"Expire", err}
	}
	minTTL, _, err := unpackUint32(msg, off)
	if err != nil {
		return SOAResource{}, &nestedError{"MinTTL", err}
	}
	return SOAResource{ns, mbox, serial, refresh, retry, expire, minTTL}, nil
}

// A TXTResource is a TXT Resource record.
type TXTResource struct {
	Txt string // Not a domain name.
}

func (r *TXTResource) realType() Type {
	return TypeTXT
}

func (r *TXTResource) pack(msg []byte, compression map[string]int) ([]byte, error) {
	return packText(msg, r.Txt), nil
}

func unpackTXTResource(msg []byte, off int, length uint16) (TXTResource, error) {
	var txt string
	for n := uint16(0); n < length; {
		var t string
		var err error
		if t, off, err = unpackText(msg, off); err != nil {
			return TXTResource{}, &nestedError{"text", err}
		}
		// Check if we got too many bytes.
		if length-n < uint16(len(t))+1 {
			return TXTResource{}, errCalcLen
		}
		n += uint16(len(t)) + 1
		txt += t
	}
	return TXTResource{txt}, nil
}

// An SRVResource is an SRV Resource record.
type SRVResource struct {
	Priority uint16
	Weight   uint16
	Port     uint16
	Target   Name // Not compressed as per RFC 2782.
}

func (r *SRVResource) realType() Type {
	return TypeSRV
}

func (r *SRVResource) pack(msg []byte, compression map[string]int) ([]byte, error) {
	oldMsg := msg
	msg = packUint16(msg, r.Priority)
	msg = packUint16(msg, r.Weight)
	msg = packUint16(msg, r.Port)
	msg, err := r.Target.pack(msg, nil)
	if err != nil {
		return oldMsg, &nestedError{"SRVResource.Target", err}
	}
	return msg, nil
}

func unpackSRVResource(msg []byte, off int) (SRVResource, error) {
	priority, off, err := unpackUint16(msg, off)
	if err != nil {
		return SRVResource{}, &nestedError{"Priority", err}
	}
	weight, off, err := unpackUint16(msg, off)
	if err != nil {
		return SRVResource{}, &nestedError{"Weight", err}
	}
	port, off, err := unpackUint16(msg, off)
	if err != nil {
		return SRVResource{}, &nestedError{"Port", err}
	}
	var target Name
	if _, err := target.unpack(msg, off); err != nil {
		return SRVResource{}, &nestedError{"Target", err}
	}
	return SRVResource{priority, weight, port, target}, nil
}

// An AResource is an A Resource record.
type AResource struct {
	A [4]byte
}

func (r *AResource) realType() Type {
	return TypeA
}

func (r *AResource) pack(msg []byte, compression map[string]int) ([]byte, error) {
	return packBytes(msg, r.A[:]), nil
}

func unpackAResource(msg []byte, off int) (AResource, error) {
	var a [4]byte
	if _, err := unpackBytes(msg, off, a[:]); err != nil {
		return AResource{}, err
	}
	return AResource{a}, nil
}

// An AAAAResource is an AAAA Resource record.
type AAAAResource struct {
	AAAA [16]byte
}

func (r *AAAAResource) realType() Type {
	return TypeAAAA
}

func (r *AAAAResource) pack(msg []byte, compression map[string]int) ([]byte, error) {
	return packBytes(msg, r.AAAA[:]), nil
}

func unpackAAAAResource(msg []byte, off int) (AAAAResource, error) {
	var aaaa [16]byte
	if _, err := unpackBytes(msg, off, aaaa[:]); err != nil {
		return AAAAResource{}, err
	}
	return AAAAResource{aaaa}, nil
}
