// Copyright 2020-2021 The NATS Authors
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
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nats-io/nuid"
)

// Request API subjects for JetStream.
const (
	// defaultAPIPrefix is the default prefix for the JetStream API.
	defaultAPIPrefix = "$JS.API."

	// jsDomainT is used to create JetStream API prefix by specifying only Domain
	jsDomainT = "$JS.%s.API."

	// apiAccountInfo is for obtaining general information about JetStream.
	apiAccountInfo = "INFO"

	// apiConsumerCreateT is used to create consumers.
	apiConsumerCreateT = "CONSUMER.CREATE.%s"

	// apiDurableCreateT is used to create durable consumers.
	apiDurableCreateT = "CONSUMER.DURABLE.CREATE.%s.%s"

	// apiConsumerInfoT is used to create consumers.
	apiConsumerInfoT = "CONSUMER.INFO.%s.%s"

	// apiRequestNextT is the prefix for the request next message(s) for a consumer in worker/pull mode.
	apiRequestNextT = "CONSUMER.MSG.NEXT.%s.%s"

	// apiDeleteConsumerT is used to delete consumers.
	apiConsumerDeleteT = "CONSUMER.DELETE.%s.%s"

	// apiConsumerListT is used to return all detailed consumer information
	apiConsumerListT = "CONSUMER.LIST.%s"

	// apiConsumerNamesT is used to return a list with all consumer names for the stream.
	apiConsumerNamesT = "CONSUMER.NAMES.%s"

	// apiStreams can lookup a stream by subject.
	apiStreams = "STREAM.NAMES"

	// apiStreamCreateT is the endpoint to create new streams.
	apiStreamCreateT = "STREAM.CREATE.%s"

	// apiStreamInfoT is the endpoint to get information on a stream.
	apiStreamInfoT = "STREAM.INFO.%s"

	// apiStreamUpdate is the endpoint to update existing streams.
	apiStreamUpdateT = "STREAM.UPDATE.%s"

	// apiStreamDeleteT is the endpoint to delete streams.
	apiStreamDeleteT = "STREAM.DELETE.%s"

	// apiPurgeStreamT is the endpoint to purge streams.
	apiStreamPurgeT = "STREAM.PURGE.%s"

	// apiStreamListT is the endpoint that will return all detailed stream information
	apiStreamList = "STREAM.LIST"

	// apiMsgGetT is the endpoint to get a message.
	apiMsgGetT = "STREAM.MSG.GET.%s"

	// apiMsgDeleteT is the endpoint to remove a message.
	apiMsgDeleteT = "STREAM.MSG.DELETE.%s"
)

// JetStream allows persistent messaging through JetStream.
type JetStream interface {
	// Publish publishes a message to JetStream.
	Publish(subj string, data []byte, opts ...PubOpt) (*PubAck, error)

	// PublishMsg publishes a Msg to JetStream.
	PublishMsg(m *Msg, opts ...PubOpt) (*PubAck, error)

	// PublishAsync publishes a message to JetStream and returns a PubAckFuture.
	// The data should not be changed until the PubAckFuture has been processed.
	PublishAsync(subj string, data []byte, opts ...PubOpt) (PubAckFuture, error)

	// PublishMsgAsync publishes a Msg to JetStream and returms a PubAckFuture.
	// The message should not be changed until the PubAckFuture has been processed.
	PublishMsgAsync(m *Msg, opts ...PubOpt) (PubAckFuture, error)

	// PublishAsyncPending returns the number of async publishes outstanding for this context.
	PublishAsyncPending() int

	// PublishAsyncComplete returns a channel that will be closed when all outstanding messages are ack'd.
	PublishAsyncComplete() <-chan struct{}

	// Subscribe creates an async Subscription for JetStream.
	Subscribe(subj string, cb MsgHandler, opts ...SubOpt) (*Subscription, error)

	// SubscribeSync creates a Subscription that can be used to process messages synchronously.
	SubscribeSync(subj string, opts ...SubOpt) (*Subscription, error)

	// ChanSubscribe creates channel based Subscription.
	ChanSubscribe(subj string, ch chan *Msg, opts ...SubOpt) (*Subscription, error)

	// ChanQueueSubscribe creates channel based Subscription with a queue group.
	ChanQueueSubscribe(subj, queue string, ch chan *Msg, opts ...SubOpt) (*Subscription, error)

	// QueueSubscribe creates a Subscription with a queue group.
	QueueSubscribe(subj, queue string, cb MsgHandler, opts ...SubOpt) (*Subscription, error)

	// QueueSubscribeSync creates a Subscription with a queue group that can be used to process messages synchronously.
	QueueSubscribeSync(subj, queue string, opts ...SubOpt) (*Subscription, error)

	// PullSubscribe creates a Subscription that can fetch messages.
	PullSubscribe(subj, durable string, opts ...SubOpt) (*Subscription, error)
}

// JetStreamContext allows JetStream messaging and stream management.
type JetStreamContext interface {
	JetStream
	JetStreamManager
}

// js is an internal struct from a JetStreamContext.
type js struct {
	nc   *Conn
	opts *jsOpts

	// For async publish context.
	mu   sync.RWMutex
	rpre string
	rsub *Subscription
	pafs map[string]*pubAckFuture
	stc  chan struct{}
	dch  chan struct{}
	rr   *rand.Rand
}

type jsOpts struct {
	ctx context.Context
	// For importing JetStream from other accounts.
	pre string
	// Amount of time to wait for API requests.
	wait time.Duration
	// For async publish error handling.
	aecb MsgErrHandler
	// Maximum in flight.
	maxap int
}

const (
	defaultRequestWait  = 5 * time.Second
	defaultAccountCheck = 20 * time.Second
)

// JetStream returns a JetStreamContext for messaging and stream management.
// Errors are only returned if inconsistent options are provided.
func (nc *Conn) JetStream(opts ...JSOpt) (JetStreamContext, error) {
	js := &js{
		nc: nc,
		opts: &jsOpts{
			pre:  defaultAPIPrefix,
			wait: defaultRequestWait,
		},
	}

	for _, opt := range opts {
		if err := opt.configureJSContext(js.opts); err != nil {
			return nil, err
		}
	}
	return js, nil
}

// JSOpt configures a JetStreamContext.
type JSOpt interface {
	configureJSContext(opts *jsOpts) error
}

// jsOptFn configures an option for the JetStreamContext.
type jsOptFn func(opts *jsOpts) error

func (opt jsOptFn) configureJSContext(opts *jsOpts) error {
	return opt(opts)
}

// Domain changes the domain part of JetSteam API prefix.
func Domain(domain string) JSOpt {
	return APIPrefix(fmt.Sprintf(jsDomainT, domain))
}

// APIPrefix changes the default prefix used for the JetStream API.
func APIPrefix(pre string) JSOpt {
	return jsOptFn(func(js *jsOpts) error {
		js.pre = pre
		if !strings.HasSuffix(js.pre, ".") {
			js.pre = js.pre + "."
		}
		return nil
	})
}

func (js *js) apiSubj(subj string) string {
	if js.opts.pre == _EMPTY_ {
		return subj
	}
	var b strings.Builder
	b.WriteString(js.opts.pre)
	b.WriteString(subj)
	return b.String()
}

// PubOpt configures options for publishing JetStream messages.
type PubOpt interface {
	configurePublish(opts *pubOpts) error
}

// pubOptFn is a function option used to configure JetStream Publish.
type pubOptFn func(opts *pubOpts) error

func (opt pubOptFn) configurePublish(opts *pubOpts) error {
	return opt(opts)
}

type pubOpts struct {
	ctx context.Context
	ttl time.Duration
	id  string
	lid string // Expected last msgId
	str string // Expected stream name
	seq uint64 // Expected last sequence
}

// pubAckResponse is the ack response from the JetStream API when publishing a message.
type pubAckResponse struct {
	apiResponse
	*PubAck
}

// PubAck is an ack received after successfully publishing a message.
type PubAck struct {
	Stream    string `json:"stream"`
	Sequence  uint64 `json:"seq"`
	Duplicate bool   `json:"duplicate,omitempty"`
}

// Headers for published messages.
const (
	MsgIdHdr             = "Nats-Msg-Id"
	ExpectedStreamHdr    = "Nats-Expected-Stream"
	ExpectedLastSeqHdr   = "Nats-Expected-Last-Sequence"
	ExpectedLastMsgIdHdr = "Nats-Expected-Last-Msg-Id"
)

// PublishMsg publishes a Msg to a stream from JetStream.
func (js *js) PublishMsg(m *Msg, opts ...PubOpt) (*PubAck, error) {
	var o pubOpts
	if len(opts) > 0 {
		if m.Header == nil {
			m.Header = Header{}
		}
		for _, opt := range opts {
			if err := opt.configurePublish(&o); err != nil {
				return nil, err
			}
		}
	}
	// Check for option collisions. Right now just timeout and context.
	if o.ctx != nil && o.ttl != 0 {
		return nil, ErrContextAndTimeout
	}
	if o.ttl == 0 && o.ctx == nil {
		o.ttl = js.opts.wait
	}

	if o.id != _EMPTY_ {
		m.Header.Set(MsgIdHdr, o.id)
	}
	if o.lid != _EMPTY_ {
		m.Header.Set(ExpectedLastMsgIdHdr, o.lid)
	}
	if o.str != _EMPTY_ {
		m.Header.Set(ExpectedStreamHdr, o.str)
	}
	if o.seq > 0 {
		m.Header.Set(ExpectedLastSeqHdr, strconv.FormatUint(o.seq, 10))
	}

	var resp *Msg
	var err error

	if o.ttl > 0 {
		resp, err = js.nc.RequestMsg(m, time.Duration(o.ttl))
	} else {
		resp, err = js.nc.RequestMsgWithContext(o.ctx, m)
	}

	if err != nil {
		if err == ErrNoResponders {
			err = ErrNoStreamResponse
		}
		return nil, err
	}
	var pa pubAckResponse
	if err := json.Unmarshal(resp.Data, &pa); err != nil {
		return nil, ErrInvalidJSAck
	}
	if pa.Error != nil {
		return nil, fmt.Errorf("nats: %s", pa.Error.Description)
	}
	if pa.PubAck == nil || pa.PubAck.Stream == _EMPTY_ {
		return nil, ErrInvalidJSAck
	}
	return pa.PubAck, nil
}

// Publish publishes a message to a stream from JetStream.
func (js *js) Publish(subj string, data []byte, opts ...PubOpt) (*PubAck, error) {
	return js.PublishMsg(&Msg{Subject: subj, Data: data}, opts...)
}

// PubAckFuture is a future for a PubAck.
type PubAckFuture interface {
	// Ok returns a receive only channel that can be used to get a PubAck.
	Ok() <-chan *PubAck

	// Err returns a receive only channel that can be used to get the error from an async publish.
	Err() <-chan error

	// Msg returns the message that was sent to the server.
	Msg() *Msg
}

type pubAckFuture struct {
	js     *js
	msg    *Msg
	pa     *PubAck
	st     time.Time
	err    error
	errCh  chan error
	doneCh chan *PubAck
}

func (paf *pubAckFuture) Ok() <-chan *PubAck {
	paf.js.mu.Lock()
	defer paf.js.mu.Unlock()

	if paf.doneCh == nil {
		paf.doneCh = make(chan *PubAck, 1)
		if paf.pa != nil {
			paf.doneCh <- paf.pa
		}
	}

	return paf.doneCh
}

func (paf *pubAckFuture) Err() <-chan error {
	paf.js.mu.Lock()
	defer paf.js.mu.Unlock()

	if paf.errCh == nil {
		paf.errCh = make(chan error, 1)
		if paf.err != nil {
			paf.errCh <- paf.err
		}
	}

	return paf.errCh
}

func (paf *pubAckFuture) Msg() *Msg {
	paf.js.mu.RLock()
	defer paf.js.mu.RUnlock()
	return paf.msg
}

// pullSubscribe creates the wildcard subscription used per pull subscriber
// to make fetch requests.
func (js *js) pullSubscribe(subj string) (*Subscription, error) {
	jsi := &jsSub{js: js, pull: true}

	// Similar to async request handler we create a wildcard subscription for making requests,
	// though we do not use the token based approach since we cannot match the response to
	// the requestor due to JS subject being remapped on delivery.  Instead, we just use an array
	// of channels similar to how ping/pong interval is handled and send the message to the first
	// available requestor via a channel.
	jsi.rr = rand.New(rand.NewSource(time.Now().UnixNano()))
	jsi.rpre = fmt.Sprintf("%s.", NewInbox())
	sub, err := js.nc.Subscribe(fmt.Sprintf("%s*", jsi.rpre), jsi.handleFetch)
	if err != nil {
		return nil, err
	}
	jsi.psub = sub

	return &Subscription{Subject: subj, conn: js.nc, typ: PullSubscription, jsi: jsi}, nil
}

// For quick token lookup etc.
const aReplyPreLen = 14
const aReplyTokensize = 6

func (js *js) newAsyncReply() string {
	js.mu.Lock()
	if js.rsub == nil {
		// Create our wildcard reply subject.
		sha := sha256.New()
		sha.Write([]byte(nuid.Next()))
		b := sha.Sum(nil)
		for i := 0; i < aReplyTokensize; i++ {
			b[i] = rdigits[int(b[i]%base)]
		}
		js.rpre = fmt.Sprintf("%s%s.", InboxPrefix, b[:aReplyTokensize])
		sub, err := js.nc.Subscribe(fmt.Sprintf("%s*", js.rpre), js.handleAsyncReply)
		if err != nil {
			js.mu.Unlock()
			return _EMPTY_
		}
		js.rsub = sub
		js.rr = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	var sb strings.Builder
	sb.WriteString(js.rpre)
	rn := js.rr.Int63()
	var b [aReplyTokensize]byte
	for i, l := 0, rn; i < len(b); i++ {
		b[i] = rdigits[l%base]
		l /= base
	}
	sb.Write(b[:])
	js.mu.Unlock()
	return sb.String()
}

// registerPAF will register for a PubAckFuture.
func (js *js) registerPAF(id string, paf *pubAckFuture) (int, int) {
	js.mu.Lock()
	if js.pafs == nil {
		js.pafs = make(map[string]*pubAckFuture)
	}
	paf.js = js
	js.pafs[id] = paf
	np := len(js.pafs)
	maxap := js.opts.maxap
	js.mu.Unlock()
	return np, maxap
}

// Lock should be held.
func (js *js) getPAF(id string) *pubAckFuture {
	if js.pafs == nil {
		return nil
	}
	return js.pafs[id]
}

// clearPAF will remove a PubAckFuture that was registered.
func (js *js) clearPAF(id string) {
	js.mu.Lock()
	delete(js.pafs, id)
	js.mu.Unlock()
}

// PublishAsyncPending returns how many PubAckFutures are pending.
func (js *js) PublishAsyncPending() int {
	js.mu.RLock()
	defer js.mu.RUnlock()
	return len(js.pafs)
}

func (js *js) asyncStall() <-chan struct{} {
	js.mu.Lock()
	if js.stc == nil {
		js.stc = make(chan struct{})
	}
	stc := js.stc
	js.mu.Unlock()
	return stc
}

// Handle an async reply from PublishAsync.
func (js *js) handleAsyncReply(m *Msg) {
	if len(m.Subject) <= aReplyPreLen {
		return
	}
	id := m.Subject[aReplyPreLen:]

	js.mu.Lock()
	paf := js.getPAF(id)
	if paf == nil {
		js.mu.Unlock()
		return
	}
	// Remove
	delete(js.pafs, id)

	// Check on anyone stalled and waiting.
	if js.stc != nil && len(js.pafs) < js.opts.maxap {
		close(js.stc)
		js.stc = nil
	}
	// Check on anyone one waiting on done status.
	if js.dch != nil && len(js.pafs) == 0 {
		dch := js.dch
		js.dch = nil
		// Defer here so error is processed and can be checked.
		defer close(dch)
	}

	doErr := func(err error) {
		paf.err = err
		if paf.errCh != nil {
			paf.errCh <- paf.err
		}
		cb := js.opts.aecb
		js.mu.Unlock()
		if cb != nil {
			cb(paf.js, paf.msg, err)
		}
	}

	// Process no responders etc.
	if len(m.Data) == 0 && m.Header.Get(statusHdr) == noResponders {
		doErr(ErrNoResponders)
		return
	}

	var pa pubAckResponse
	if err := json.Unmarshal(m.Data, &pa); err != nil {
		doErr(ErrInvalidJSAck)
		return
	}
	if pa.Error != nil {
		doErr(fmt.Errorf("nats: %s", pa.Error.Description))
		return
	}
	if pa.PubAck == nil || pa.PubAck.Stream == _EMPTY_ {
		doErr(ErrInvalidJSAck)
		return
	}

	// So here we have received a proper puback.
	paf.pa = pa.PubAck
	if paf.doneCh != nil {
		paf.doneCh <- paf.pa
	}
	js.mu.Unlock()
}

// MsgErrHandler is used to process asynchronous errors from
// JetStream PublishAsync and PublishAsynMsg. It will return the original
// message sent to the server for possible retransmitting and the error encountered.
type MsgErrHandler func(JetStream, *Msg, error)

// PublishAsyncErrHandler sets the error handler for async publishes in JetStream.
func PublishAsyncErrHandler(cb MsgErrHandler) JSOpt {
	return jsOptFn(func(js *jsOpts) error {
		js.aecb = cb
		return nil
	})
}

// PublishAsyncMaxPending sets the maximum outstanding async publishes that can be inflight at one time.
func PublishAsyncMaxPending(max int) JSOpt {
	return jsOptFn(func(js *jsOpts) error {
		if max < 1 {
			return errors.New("nats: max ack pending should be >= 1")
		}
		js.maxap = max
		return nil
	})
}

// PublishAsync publishes a message to JetStream and returns a PubAckFuture
func (js *js) PublishAsync(subj string, data []byte, opts ...PubOpt) (PubAckFuture, error) {
	return js.PublishMsgAsync(&Msg{Subject: subj, Data: data}, opts...)
}

func (js *js) PublishMsgAsync(m *Msg, opts ...PubOpt) (PubAckFuture, error) {
	var o pubOpts
	if len(opts) > 0 {
		if m.Header == nil {
			m.Header = Header{}
		}
		for _, opt := range opts {
			if err := opt.configurePublish(&o); err != nil {
				return nil, err
			}
		}
	}

	// Timeouts and contexts do not make sense for these.
	if o.ttl != 0 || o.ctx != nil {
		return nil, ErrContextAndTimeout
	}

	// FIXME(dlc) - Make common.
	if o.id != _EMPTY_ {
		m.Header.Set(MsgIdHdr, o.id)
	}
	if o.lid != _EMPTY_ {
		m.Header.Set(ExpectedLastMsgIdHdr, o.lid)
	}
	if o.str != _EMPTY_ {
		m.Header.Set(ExpectedStreamHdr, o.str)
	}
	if o.seq > 0 {
		m.Header.Set(ExpectedLastSeqHdr, strconv.FormatUint(o.seq, 10))
	}

	// Reply
	if m.Reply != _EMPTY_ {
		return nil, errors.New("nats: reply subject should be empty")
	}
	m.Reply = js.newAsyncReply()
	if m.Reply == _EMPTY_ {
		return nil, errors.New("nats: error creating async reply handler")
	}
	id := m.Reply[aReplyPreLen:]
	paf := &pubAckFuture{msg: m, st: time.Now()}
	numPending, maxPending := js.registerPAF(id, paf)

	if maxPending > 0 && numPending >= maxPending {
		select {
		case <-js.asyncStall():
		case <-time.After(200 * time.Millisecond):
			js.clearPAF(id)
			return nil, errors.New("nats: stalled with too many outstanding async published messages")
		}
	}

	if err := js.nc.PublishMsg(m); err != nil {
		js.clearPAF(id)
		return nil, err
	}

	return paf, nil
}

// PublishAsyncComplete returns a channel that will be closed when all outstanding messages have been ack'd.
func (js *js) PublishAsyncComplete() <-chan struct{} {
	js.mu.Lock()
	defer js.mu.Unlock()
	if js.dch == nil {
		js.dch = make(chan struct{})
	}
	dch := js.dch
	if len(js.pafs) == 0 {
		close(js.dch)
		js.dch = nil
	}
	return dch
}

// MsgId sets the message ID used for de-duplication.
func MsgId(id string) PubOpt {
	return pubOptFn(func(opts *pubOpts) error {
		opts.id = id
		return nil
	})
}

// ExpectStream sets the expected stream to respond from the publish.
func ExpectStream(stream string) PubOpt {
	return pubOptFn(func(opts *pubOpts) error {
		opts.str = stream
		return nil
	})
}

// ExpectLastSequence sets the expected sequence in the response from the publish.
func ExpectLastSequence(seq uint64) PubOpt {
	return pubOptFn(func(opts *pubOpts) error {
		opts.seq = seq
		return nil
	})
}

// ExpectLastMsgId sets the expected sequence in the response from the publish.
func ExpectLastMsgId(id string) PubOpt {
	return pubOptFn(func(opts *pubOpts) error {
		opts.lid = id
		return nil
	})
}

type ackOpts struct {
	ttl time.Duration
	ctx context.Context
}

// AckOpt are the options that can be passed when acknowledge a message.
type AckOpt interface {
	configureAck(opts *ackOpts) error
}

// MaxWait sets the maximum amount of time we will wait for a response.
type MaxWait time.Duration

func (ttl MaxWait) configureJSContext(js *jsOpts) error {
	js.wait = time.Duration(ttl)
	return nil
}

func (ttl MaxWait) configurePull(opts *pullOpts) error {
	opts.ttl = time.Duration(ttl)
	return nil
}

// AckWait sets the maximum amount of time we will wait for an ack.
type AckWait time.Duration

func (ttl AckWait) configurePublish(opts *pubOpts) error {
	opts.ttl = time.Duration(ttl)
	return nil
}

func (ttl AckWait) configureSubscribe(opts *subOpts) error {
	opts.cfg.AckWait = time.Duration(ttl)
	return nil
}

func (ttl AckWait) configureAck(opts *ackOpts) error {
	opts.ttl = time.Duration(ttl)
	return nil
}

// ContextOpt is an option used to set a context.Context.
type ContextOpt struct {
	context.Context
}

func (ctx ContextOpt) configureJSContext(opts *jsOpts) error {
	opts.ctx = ctx
	return nil
}

func (ctx ContextOpt) configurePublish(opts *pubOpts) error {
	opts.ctx = ctx
	return nil
}

func (ctx ContextOpt) configurePull(opts *pullOpts) error {
	opts.ctx = ctx
	return nil
}

func (ctx ContextOpt) configureAck(opts *ackOpts) error {
	opts.ctx = ctx
	return nil
}

// Context returns an option that can be used to configure a context for APIs
// that are context aware such as those part of the JetStream interface.
func Context(ctx context.Context) ContextOpt {
	return ContextOpt{ctx}
}

// Subscribe

// ConsumerConfig is the configuration of a JetStream consumer.
type ConsumerConfig struct {
	Durable         string        `json:"durable_name,omitempty"`
	DeliverSubject  string        `json:"deliver_subject,omitempty"`
	DeliverPolicy   DeliverPolicy `json:"deliver_policy"`
	OptStartSeq     uint64        `json:"opt_start_seq,omitempty"`
	OptStartTime    *time.Time    `json:"opt_start_time,omitempty"`
	AckPolicy       AckPolicy     `json:"ack_policy"`
	AckWait         time.Duration `json:"ack_wait,omitempty"`
	MaxDeliver      int           `json:"max_deliver,omitempty"`
	FilterSubject   string        `json:"filter_subject,omitempty"`
	ReplayPolicy    ReplayPolicy  `json:"replay_policy"`
	RateLimit       uint64        `json:"rate_limit_bps,omitempty"` // Bits per sec
	SampleFrequency string        `json:"sample_freq,omitempty"`
	MaxWaiting      int           `json:"max_waiting,omitempty"`
	MaxAckPending   int           `json:"max_ack_pending,omitempty"`
	FlowControl     bool          `json:"flow_control,omitempty"`
	Heartbeat       time.Duration `json:"idle_heartbeat,omitempty"`
}

// ConsumerInfo is the info from a JetStream consumer.
type ConsumerInfo struct {
	Stream         string         `json:"stream_name"`
	Name           string         `json:"name"`
	Created        time.Time      `json:"created"`
	Config         ConsumerConfig `json:"config"`
	Delivered      SequencePair   `json:"delivered"`
	AckFloor       SequencePair   `json:"ack_floor"`
	NumAckPending  int            `json:"num_ack_pending"`
	NumRedelivered int            `json:"num_redelivered"`
	NumWaiting     int            `json:"num_waiting"`
	NumPending     uint64         `json:"num_pending"`
	Cluster        *ClusterInfo   `json:"cluster,omitempty"`
}

// SequencePair includes the consumer and stream sequence info from a JetStream consumer.
type SequencePair struct {
	Consumer uint64 `json:"consumer_seq"`
	Stream   uint64 `json:"stream_seq"`
}

// nextRequest is for getting next messages for pull based consumers from JetStream.
type nextRequest struct {
	Expires time.Duration `json:"expires,omitempty"`
	Batch   int           `json:"batch,omitempty"`
	NoWait  bool          `json:"no_wait,omitempty"`
}

// jsSub includes JetStream subscription info.
type jsSub struct {
	js *js

	// To setup request mux handler for pull subscribers.
	mu    sync.RWMutex
	psub  *Subscription
	rpre  string
	rr    *rand.Rand
	freqs []chan *Msg

	consumer string
	stream   string
	deliver  string
	pull     bool
	durable  bool
	attached bool

	// Heartbeats and Flow Control handling from push consumers.
	hbs   bool
	fc    bool
	cmeta string
	fcs   map[uint64]string
}

// newFetchReply generates a unique inbox used for a fetch request.
func (jsi *jsSub) newFetchReply() string {
	jsi.mu.Lock()
	rpre := jsi.rpre
	rn := jsi.rr.Int63()
	jsi.mu.Unlock()
	var sb strings.Builder
	sb.WriteString(rpre)
	var b [aReplyTokensize]byte
	for i, l := 0, rn; i < len(b); i++ {
		b[i] = rdigits[l%base]
		l /= base
	}
	sb.Write(b[:])
	return sb.String()
}

// handleFetch is delivered a message requested by pull subscribers
// when calling Fetch.
func (jsi *jsSub) handleFetch(m *Msg) {
	jsi.mu.Lock()
	if len(jsi.freqs) == 0 {
		nc := jsi.js.nc
		sub := jsi.psub
		nc.mu.Lock()
		errCB := nc.Opts.AsyncErrorCB
		err := fmt.Errorf("nats: fetch response delivered but requestor has gone away")
		if errCB != nil {
			nc.ach.push(func() { errCB(nc, sub, err) })
		}
		nc.mu.Unlock()
		jsi.mu.Unlock()
		return
	}
	mch := jsi.freqs[0]
	if len(jsi.freqs) > 1 {
		jsi.freqs = append(jsi.freqs[:0], jsi.freqs[1:]...)
	} else {
		jsi.freqs = jsi.freqs[:0]
	}
	jsi.mu.Unlock()
	mch <- m
}

// fetchNoWait makes a request to get a single message using no wait.
func (jsi *jsSub) fetchNoWait(ctx context.Context, subj string, payload []byte) (*Msg, error) {
	nc := jsi.js.nc
	m := NewMsg(subj)
	m.Reply = jsi.newFetchReply()
	m.Data = payload

	mch := make(chan *Msg, 1)
	jsi.mu.Lock()
	jsi.freqs = append(jsi.freqs, mch)
	jsi.mu.Unlock()
	if err := nc.PublishMsg(m); err != nil {
		return nil, err
	}

	var ok bool
	var msg *Msg

	select {
	case msg, ok = <-mch:
		if !ok {
			return nil, ErrConnectionClosed
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return msg, nil
}

func (jsi *jsSub) unsubscribe(drainMode bool) error {
	if drainMode && (jsi.durable || jsi.attached) {
		// Skip deleting consumer for durables/attached
		// consumers when using drain mode.
		return nil
	}
	// Clear the extra async pull subscription used for fetch requests.
	if jsi.psub != nil {
		jsi.psub.Drain()
	}

	js := jsi.js
	return js.DeleteConsumer(jsi.stream, jsi.consumer)
}

// SubOpt configures options for subscribing to JetStream consumers.
type SubOpt interface {
	configureSubscribe(opts *subOpts) error
}

// subOptFn is a function option used to configure a JetStream Subscribe.
type subOptFn func(opts *subOpts) error

func (opt subOptFn) configureSubscribe(opts *subOpts) error {
	return opt(opts)
}

// Subscribe will create a subscription to the appropriate stream and consumer.
func (js *js) Subscribe(subj string, cb MsgHandler, opts ...SubOpt) (*Subscription, error) {
	if cb == nil {
		return nil, ErrBadSubscription
	}
	return js.subscribe(subj, _EMPTY_, cb, nil, false, opts)
}

// SubscribeSync will create a sync subscription to the appropriate stream and consumer.
func (js *js) SubscribeSync(subj string, opts ...SubOpt) (*Subscription, error) {
	mch := make(chan *Msg, js.nc.Opts.SubChanLen)
	return js.subscribe(subj, _EMPTY_, nil, mch, true, opts)
}

// QueueSubscribe will create a subscription to the appropriate stream and consumer with queue semantics.
func (js *js) QueueSubscribe(subj, queue string, cb MsgHandler, opts ...SubOpt) (*Subscription, error) {
	if cb == nil {
		return nil, ErrBadSubscription
	}
	return js.subscribe(subj, queue, cb, nil, false, opts)
}

// QueueSubscribeSync will create a sync subscription to the appropriate stream and consumer with queue semantics.
func (js *js) QueueSubscribeSync(subj, queue string, opts ...SubOpt) (*Subscription, error) {
	mch := make(chan *Msg, js.nc.Opts.SubChanLen)
	return js.subscribe(subj, queue, nil, mch, true, opts)
}

// ChanSubscribe will create a subscription to the appropriate stream and consumer using a channel.
func (js *js) ChanSubscribe(subj string, ch chan *Msg, opts ...SubOpt) (*Subscription, error) {
	return js.subscribe(subj, _EMPTY_, nil, ch, false, opts)
}

// ChanQueueSubscribe will create a subscription to the appropriate stream and consumer using a channel.
func (js *js) ChanQueueSubscribe(subj, queue string, ch chan *Msg, opts ...SubOpt) (*Subscription, error) {
	return js.subscribe(subj, queue, nil, ch, false, opts)
}

// PullSubscribe creates a pull subscriber.
func (js *js) PullSubscribe(subj, durable string, opts ...SubOpt) (*Subscription, error) {
	return js.subscribe(subj, _EMPTY_, nil, nil, false, append(opts, Durable(durable)))
}

func (js *js) subscribe(subj, queue string, cb MsgHandler, ch chan *Msg, isSync bool, opts []SubOpt) (*Subscription, error) {
	cfg := ConsumerConfig{AckPolicy: ackPolicyNotSet}
	o := subOpts{cfg: &cfg}
	if len(opts) > 0 {
		for _, opt := range opts {
			if err := opt.configureSubscribe(&o); err != nil {
				return nil, err
			}
		}
	}

	isPullMode := ch == nil && cb == nil && !isSync
	badPullAck := o.cfg.AckPolicy == AckNonePolicy || o.cfg.AckPolicy == AckAllPolicy
	hasHeartbeats := o.cfg.Heartbeat > 0
	hasFC := o.cfg.FlowControl
	if isPullMode && badPullAck {
		return nil, fmt.Errorf("nats: invalid ack mode for pull consumers: %s", o.cfg.AckPolicy)
	}

	var (
		err           error
		shouldCreate  bool
		ccfg          *ConsumerConfig
		info          *ConsumerInfo
		deliver       string
		attached      bool
		stream        = o.stream
		consumer      = o.consumer
		isDurable     = o.cfg.Durable != _EMPTY_
		consumerBound = o.bound
		notFoundErr   bool
		lookupErr     bool
	)

	// In case a consumer has not been set explicitly, then the
	// durable name will be used as the consumer name.
	if consumer == _EMPTY_ {
		consumer = o.cfg.Durable
	}

	// Find the stream mapped to the subject if not bound to a stream already.
	if o.stream == _EMPTY_ {
		stream, err = js.lookupStreamBySubject(subj)
		if err != nil {
			return nil, err
		}
	} else {
		stream = o.stream
	}

	// With an explicit durable name, then can lookup the consumer first
	// to which it should be attaching to.
	if consumer != _EMPTY_ {
		info, err = js.ConsumerInfo(stream, consumer)
		notFoundErr = err != nil && strings.Contains(err.Error(), "consumer not found")
		lookupErr = err == ErrJetStreamNotEnabled || err == ErrTimeout || err == context.DeadlineExceeded
	}

	switch {
	case info != nil:
		// Attach using the found consumer config.
		ccfg = &info.Config
		attached = true

		// Make sure this new subject matches or is a subset.
		if ccfg.FilterSubject != _EMPTY_ && subj != ccfg.FilterSubject {
			return nil, ErrSubjectMismatch
		}

		// Prevent binding a subscription against incompatible consumer types.
		if isPullMode && ccfg.DeliverSubject != _EMPTY_ {
			return nil, ErrPullSubscribeToPushConsumer
		} else if !isPullMode && ccfg.DeliverSubject == _EMPTY_ {
			return nil, ErrPullSubscribeRequired
		}

		if ccfg.DeliverSubject != _EMPTY_ {
			deliver = ccfg.DeliverSubject
		} else if !isPullMode {
			deliver = NewInbox()
		}
	case (err != nil && !notFoundErr) || (notFoundErr && consumerBound):
		// If the consumer is being bound got an error on pull subscribe then allow the error.
		if !(isPullMode && lookupErr && consumerBound) {
			return nil, err
		}
	default:
		// Attempt to create consumer if not found nor using Bind.
		shouldCreate = true
		deliver = NewInbox()
		if !isPullMode {
			cfg.DeliverSubject = deliver
		}
		// Do filtering always, server will clear as needed.
		cfg.FilterSubject = subj
	}

	var sub *Subscription

	// Check if we are manual ack.
	if cb != nil && !o.mack {
		ocb := cb
		cb = func(m *Msg) { ocb(m); m.Ack() }
	}

	if isPullMode {
		sub, err = js.pullSubscribe(subj)
	} else {
		sub, err = js.nc.subscribe(deliver, queue, cb, ch, isSync, &jsSub{js: js, hbs: hasHeartbeats, fc: hasFC})
	}
	if err != nil {
		return nil, err
	}

	// With flow control enabled async subscriptions we will disable msgs
	// limits, and set a larger pending bytes limit by default.
	if !isPullMode && cb != nil && hasFC {
		sub.SetPendingLimits(DefaultSubPendingMsgsLimit*16, DefaultSubPendingBytesLimit)
	}

	// If we are creating or updating let's process that request.
	if shouldCreate {
		// If not set default to ack explicit.
		if cfg.AckPolicy == ackPolicyNotSet {
			cfg.AckPolicy = AckExplicitPolicy
		}
		// If we have acks at all and the MaxAckPending is not set go ahead
		// and set to the internal max.
		// TODO(dlc) - We should be able to update this if client updates PendingLimits.
		if cfg.MaxAckPending == 0 && cfg.AckPolicy != AckNonePolicy {
			maxMsgs, _, _ := sub.PendingLimits()
			cfg.MaxAckPending = maxMsgs
		}

		req := &createConsumerRequest{
			Stream: stream,
			Config: &cfg,
		}

		j, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}

		var ccSubj string
		if isDurable {
			ccSubj = fmt.Sprintf(apiDurableCreateT, stream, cfg.Durable)
		} else {
			ccSubj = fmt.Sprintf(apiConsumerCreateT, stream)
		}

		resp, err := js.nc.Request(js.apiSubj(ccSubj), j, js.opts.wait)
		if err != nil {
			sub.Drain()
			if err == ErrNoResponders {
				err = ErrJetStreamNotEnabled
			}
			return nil, err
		}
		var cinfo consumerResponse
		err = json.Unmarshal(resp.Data, &cinfo)
		if err != nil {
			sub.Drain()
			return nil, err
		}
		info = cinfo.ConsumerInfo
		if cinfo.Error != nil {
			// Remove interest from previous subscribe since it
			// may have an incorrect delivery subject.
			sub.Drain()

			// Multiple subscribers could compete in creating the first consumer
			// that will be shared using the same durable name. If this happens, then
			// do a lookup of the consumer info and resubscribe using the latest info.
			if consumer != _EMPTY_ && (strings.Contains(cinfo.Error.Description, `consumer already exists`) || strings.Contains(cinfo.Error.Description, `consumer name already in use`)) {
				info, err = js.ConsumerInfo(stream, consumer)
				if err != nil {
					return nil, err
				}
				ccfg = &info.Config

				// Validate that the original subject does still match.
				if ccfg.FilterSubject != _EMPTY_ && subj != ccfg.FilterSubject {
					return nil, ErrSubjectMismatch
				}

				// Use the deliver subject from latest consumer config to attach.
				if ccfg.DeliverSubject != _EMPTY_ {
					// We can't reuse the channel, so if one was passed, we need to create a new one.
					if ch != nil {
						ch = make(chan *Msg, cap(ch))
					}
					sub, err = js.nc.subscribe(ccfg.DeliverSubject, queue, cb, ch, isSync,
						&jsSub{js: js, hbs: hasHeartbeats, fc: hasFC})
					if err != nil {
						return nil, err
					}
				}
				attached = true
			} else {
				return nil, fmt.Errorf("nats: %s", cinfo.Error.Description)
			}
		}
		stream = info.Stream
		consumer = info.Name
		deliver = info.Config.DeliverSubject
	}
	sub.mu.Lock()
	sub.jsi.stream = stream
	sub.jsi.consumer = consumer
	sub.jsi.durable = isDurable
	sub.jsi.attached = attached
	sub.jsi.deliver = deliver
	sub.mu.Unlock()

	return sub, nil
}

// ErrConsumerSequenceMismatch represents an error from a consumer
// that received a Heartbeat including sequence different to the
// one expected from the view of the client.
type ErrConsumerSequenceMismatch struct {
	// StreamResumeSequence is the stream sequence from where the consumer
	// should resume consuming from the stream.
	StreamResumeSequence uint64

	// ConsumerSequence is the sequence of the consumer that is behind.
	ConsumerSequence uint64

	// LastConsumerSequence is the sequence of the consumer when the heartbeat
	// was received.
	LastConsumerSequence uint64
}

func (ecs *ErrConsumerSequenceMismatch) Error() string {
	return fmt.Sprintf("nats: sequence mismatch for consumer at sequence %d (%d sequences behind), should restart consumer from stream sequence %d",
		ecs.ConsumerSequence,
		ecs.LastConsumerSequence-ecs.ConsumerSequence,
		ecs.StreamResumeSequence,
	)
}

// isControlMessage will return true if this is an empty control status message.
func isControlMessage(msg *Msg) bool {
	return len(msg.Data) == 0 && msg.Header.Get(statusHdr) == controlMsg
}

func (jsi *jsSub) trackSequences(reply string) {
	jsi.mu.Lock()
	jsi.cmeta = reply
	jsi.mu.Unlock()
}

// checkForFlowControlResponse will check to see if we should send a flow control response
// based on the delivered index.
// Lock should be held.
func (sub *Subscription) checkForFlowControlResponse(delivered uint64) {
	jsi, nc := sub.jsi, sub.conn
	if jsi == nil {
		return
	}

	jsi.mu.Lock()
	defer jsi.mu.Unlock()

	if len(jsi.fcs) == 0 {
		return
	}

	if reply := jsi.fcs[delivered]; reply != _EMPTY_ {
		delete(jsi.fcs, delivered)
		nc.Publish(reply, nil)
	}
}

// Record an inbound flow control message.
func (jsi *jsSub) scheduleFlowControlResponse(dfuture uint64, reply string) {
	jsi.mu.Lock()
	if jsi.fcs == nil {
		jsi.fcs = make(map[uint64]string)
	}
	jsi.fcs[dfuture] = reply
	jsi.mu.Unlock()
}

// handleConsumerSequenceMismatch will send an async error that can be used to restart a push based consumer.
func (nc *Conn) handleConsumerSequenceMismatch(sub *Subscription, err error) {
	nc.mu.Lock()
	errCB := nc.Opts.AsyncErrorCB
	if errCB != nil {
		nc.ach.push(func() { errCB(nc, sub, err) })
	}
	nc.mu.Unlock()
}

// processControlFlow will automatically respond to control messages sent by the server.
func (nc *Conn) processSequenceMismatch(msg *Msg, s *Subscription, jsi *jsSub) {
	// Process heartbeat received, get latest control metadata if present.
	jsi.mu.RLock()
	ctrl := jsi.cmeta
	jsi.mu.RUnlock()

	if ctrl == _EMPTY_ {
		return
	}

	tokens, err := getMetadataFields(ctrl)
	if err != nil {
		return
	}

	// Consumer sequence.
	var ldseq string
	dseq := tokens[6]
	hdr := msg.Header[lastConsumerSeqHdr]
	if len(hdr) == 1 {
		ldseq = hdr[0]
	}

	// Detect consumer sequence mismatch and whether
	// should restart the consumer.
	if ldseq != dseq {
		// Dispatch async error including details such as
		// from where the consumer could be restarted.
		sseq := parseNum(tokens[5])
		ecs := &ErrConsumerSequenceMismatch{
			StreamResumeSequence: uint64(sseq),
			ConsumerSequence:     uint64(parseNum(dseq)),
			LastConsumerSequence: uint64(parseNum(ldseq)),
		}
		nc.handleConsumerSequenceMismatch(s, ecs)
	}
}

type streamRequest struct {
	Subject string `json:"subject,omitempty"`
}

type streamNamesResponse struct {
	apiResponse
	apiPaged
	Streams []string `json:"streams"`
}

func (js *js) lookupStreamBySubject(subj string) (string, error) {
	var slr streamNamesResponse
	req := &streamRequest{subj}
	j, err := json.Marshal(req)
	if err != nil {
		return _EMPTY_, err
	}
	resp, err := js.nc.Request(js.apiSubj(apiStreams), j, js.opts.wait)
	if err != nil {
		if err == ErrNoResponders {
			err = ErrJetStreamNotEnabled
		}
		return _EMPTY_, err
	}
	if err := json.Unmarshal(resp.Data, &slr); err != nil {
		return _EMPTY_, err
	}
	if slr.Error != nil || len(slr.Streams) != 1 {
		return _EMPTY_, ErrNoMatchingStream
	}
	return slr.Streams[0], nil
}

type subOpts struct {
	// For attaching.
	stream, consumer string
	// For manual ack
	mack bool
	// For creating or updating.
	cfg *ConsumerConfig
	// For binding a subscription to a consumer without creating it.
	bound bool
}

// ManualAck disables auto ack functionality for async subscriptions.
func ManualAck() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.mack = true
		return nil
	})
}

// Durable defines the consumer name for JetStream durable subscribers.
func Durable(consumer string) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		if opts.cfg.Durable != _EMPTY_ {
			return fmt.Errorf("nats: option Durable set more than once")
		}
		if opts.consumer != _EMPTY_ && opts.consumer != consumer {
			return fmt.Errorf("nats: duplicate consumer names (%s and %s)", opts.consumer, consumer)
		}
		if strings.Contains(consumer, ".") {
			return ErrInvalidDurableName
		}

		opts.cfg.Durable = consumer
		return nil
	})
}

// DeliverAll will configure a Consumer to receive all the
// messages from a Stream.
func DeliverAll() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.DeliverPolicy = DeliverAllPolicy
		return nil
	})
}

// DeliverLast configures a Consumer to receive messages
// starting with the latest one.
func DeliverLast() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.DeliverPolicy = DeliverLastPolicy
		return nil
	})
}

// DeliverNew configures a Consumer to receive messages
// published after the subscription.
func DeliverNew() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.DeliverPolicy = DeliverNewPolicy
		return nil
	})
}

// StartSequence configures a Consumer to receive
// messages from a start sequence.
func StartSequence(seq uint64) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.DeliverPolicy = DeliverByStartSequencePolicy
		opts.cfg.OptStartSeq = seq
		return nil
	})
}

// StartTime configures a Consumer to receive
// messages from a start time.
func StartTime(startTime time.Time) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.DeliverPolicy = DeliverByStartTimePolicy
		opts.cfg.OptStartTime = &startTime
		return nil
	})
}

// AckNone requires no acks for delivered messages.
func AckNone() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.AckPolicy = AckNonePolicy
		return nil
	})
}

// AckAll when acking a sequence number, this implicitly acks all sequences
// below this one as well.
func AckAll() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.AckPolicy = AckAllPolicy
		return nil
	})
}

// AckExplicit requires ack or nack for all messages.
func AckExplicit() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.AckPolicy = AckExplicitPolicy
		return nil
	})
}

// MaxDeliver sets the number of redeliveries for a message.
func MaxDeliver(n int) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.MaxDeliver = n
		return nil
	})
}

// MaxAckPending sets the number of outstanding acks that are allowed before
// message delivery is halted.
func MaxAckPending(n int) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.MaxAckPending = n
		return nil
	})
}

// ReplayOriginal replays the messages at the original speed.
func ReplayOriginal() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.ReplayPolicy = ReplayOriginalPolicy
		return nil
	})
}

// RateLimit is the Bits per sec rate limit applied to a push consumer.
func RateLimit(n uint64) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.RateLimit = n
		return nil
	})
}

// BindStream binds a consumer to a stream explicitly based on a name.
func BindStream(stream string) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		if opts.stream != _EMPTY_ && opts.stream != stream {
			return fmt.Errorf("nats: duplicate stream name (%s and %s)", opts.stream, stream)
		}

		opts.stream = stream
		return nil
	})
}

// Bind binds a subscription to an existing consumer from a stream without attempting to create.
// The first argument is the stream name and the second argument will be the consumer name.
func Bind(stream, consumer string) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		if stream == _EMPTY_ {
			return ErrStreamNameRequired
		}
		if consumer == _EMPTY_ {
			return ErrConsumerNameRequired
		}

		// In case of pull subscribers, the durable name is a required parameter
		// so check that they are not different.
		if opts.cfg.Durable != _EMPTY_ && opts.cfg.Durable != consumer {
			return fmt.Errorf("nats: duplicate consumer names (%s and %s)", opts.cfg.Durable, consumer)
		}
		if opts.stream != _EMPTY_ && opts.stream != stream {
			return fmt.Errorf("nats: duplicate stream name (%s and %s)", opts.stream, stream)
		}
		opts.stream = stream
		opts.consumer = consumer
		opts.bound = true
		return nil
	})
}

// EnableFlowControl enables flow control for a push based consumer.
func EnableFlowControl() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.FlowControl = true
		return nil
	})
}

// IdleHeartbeat enables push based consumers to have idle heartbeats delivered.
func IdleHeartbeat(duration time.Duration) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.Heartbeat = duration
		return nil
	})
}

func (sub *Subscription) ConsumerInfo() (*ConsumerInfo, error) {
	sub.mu.Lock()
	// TODO(dlc) - Better way to mark especially if we attach.
	if sub.jsi.consumer == _EMPTY_ {
		sub.mu.Unlock()
		return nil, ErrTypeSubscription
	}

	// Consumer info lookup should fail if in direct mode.
	js := sub.jsi.js
	stream, consumer := sub.jsi.stream, sub.jsi.consumer
	sub.mu.Unlock()

	return js.getConsumerInfo(stream, consumer)
}

type pullOpts struct {
	ttl time.Duration
	ctx context.Context
}

// PullOpt are the options that can be passed when pulling a batch of messages.
type PullOpt interface {
	configurePull(opts *pullOpts) error
}

// PullMaxWaiting defines the max inflight pull requests.
func PullMaxWaiting(n int) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.MaxWaiting = n
		return nil
	})
}

var errNoMessages = errors.New("nats: no messages")

// Fetch pulls a batch of messages from a stream for a pull consumer.
func (sub *Subscription) Fetch(batch int, opts ...PullOpt) ([]*Msg, error) {
	if sub == nil {
		return nil, ErrBadSubscription
	}

	var o pullOpts
	for _, opt := range opts {
		if err := opt.configurePull(&o); err != nil {
			return nil, err
		}
	}
	if o.ctx != nil && o.ttl != 0 {
		return nil, ErrContextAndTimeout
	}

	sub.mu.Lock()
	jsi := sub.jsi
	if jsi == nil || sub.typ != PullSubscription {
		sub.mu.Unlock()
		return nil, ErrTypeSubscription
	}

	nc, _ := sub.conn, sub.Subject
	stream, consumer := sub.jsi.stream, sub.jsi.consumer
	js := sub.jsi.js

	ttl := o.ttl
	if ttl == 0 {
		ttl = js.opts.wait
	}
	sub.mu.Unlock()

	// Use the given context or setup a default one for the span
	// of the pull batch request.
	var (
		ctx    = o.ctx
		err    error
		cancel context.CancelFunc
	)
	if o.ctx == nil {
		ctx, cancel = context.WithTimeout(context.Background(), ttl)
		defer cancel()
	}

	// Check if context not done already before making the request.
	select {
	case <-ctx.Done():
		if ctx.Err() == context.Canceled {
			err = ctx.Err()
		} else {
			err = ErrTimeout
		}
	default:
	}
	if err != nil {
		return nil, err
	}

	// Check for empty payload message and process synchronously
	// any status messages.
	checkMsg := func(msg *Msg) error {
		if len(msg.Data) == 0 {
			switch msg.Header.Get(statusHdr) {
			case noResponders:
				return ErrNoResponders
			case noMessages:
				return errNoMessages
			case "400", "408", "409":
				return fmt.Errorf("nats: %s", msg.Header.Get(descrHdr))
			}
		}
		return nil
	}

	checkCtxErr := func(err error) error {
		if o.ctx == nil && err == context.DeadlineExceeded {
			return ErrTimeout
		}
		return err
	}

	var (
		gotNoMessages bool
		nr            = &nextRequest{Batch: batch, NoWait: true}
		req, _        = json.Marshal(nr)
		reqNext       = js.apiSubj(fmt.Sprintf(apiRequestNextT, stream, consumer))
		expires       = ttl - 10*time.Millisecond
		msgs          = make([]*Msg, 0)
	)

	if batch == 1 {
		// To optimize single message no wait fetch, we use a shared wildcard
		// subscription per pull subscriber to wait for the response.
		resp, err := jsi.fetchNoWait(ctx, reqNext, req)
		if err != nil {
			return nil, checkCtxErr(err)
		}

		// In case of a no messages instant error, then fallback
		// into longer version of pull batch request.
		err = checkMsg(resp)
		if err != nil {
			if err == errNoMessages {
				// Use old request style for the retry of the pull request
				// in order to use auto UNSUB 1 to prevent the server
				// from delivering a message when there is no more interest.
				nr.NoWait = false
				nr.Expires = expires
				req, _ = json.Marshal(nr)
				resp, err = nc.oldRequestWithContext(ctx, reqNext, nil, req)
				if err != nil {
					return nil, checkCtxErr(err)
				}

				// This next message, could also be an error
				// (e.g. 408 due to request timeout).
				err = checkMsg(resp)
				if err != nil {
					return nil, err
				}
				return []*Msg{resp}, nil
			} else {
				// Hard error
				return nil, checkCtxErr(err)
			}
		}
		return []*Msg{resp}, nil
	}

	// Setup a request where we will wait for the first response
	// in case of errors, then dispatch the rest of the replies
	// to the channel.
	inbox := NewInbox()

	mch := make(chan *Msg, batch)
	s, err := nc.subscribe(inbox, _EMPTY_, nil, mch, true, nil)
	if err != nil {
		return nil, err
	}

	// Remove interest in the subscription at the end so that the
	// this inbox does not get delivered the results intended
	// for another request.
	defer s.Unsubscribe()

	// Make a publish request to get results of the pull.
	err = nc.publish(reqNext, inbox, nil, req)
	if err != nil {
		s.Unsubscribe()
		return nil, err
	}

	// Try to get the first message or error with NoWait.
	var (
		firstMsg *Msg
		ok       bool
	)
	select {
	case firstMsg, ok = <-mch:
		if !ok {
			err = s.getNextMsgErr()
		} else {
			err = s.processNextMsgDelivered(firstMsg)
			if err == nil {
				err = checkMsg(firstMsg)
			}
		}
	case <-ctx.Done():
		err = checkCtxErr(ctx.Err())
	}

	// If the first error is 'no more messages', then switch into
	// longer form version of the request that waits for messages.
	if err == errNoMessages {
		gotNoMessages = true
	} else if err != nil {
		// We should be getting the response from the server
		// in case we got a poll error, so stop and cleanup.
		s.Unsubscribe()
		return nil, err
	}

	if gotNoMessages {
		// We started with a 404 response right away, so fallback into
		// second request that waits longer for messages to delivered.
		nr.NoWait = false
		nr.Expires = expires
		req, _ = json.Marshal(nr)

		// Since first message was an error we UNSUB (batch+1)
		// since we are counting it as the first message.
		err = s.AutoUnsubscribe(batch + 1)
		if err != nil {
			return nil, err
		}

		// Make another request and wait for the messages...
		err = nc.publish(reqNext, inbox, nil, req)
		if err != nil {
			s.Unsubscribe()
			return nil, err
		}

		// Try to get the first result again or return the error.
		select {
		case firstMsg, ok = <-mch:
			if !ok {
				err = s.getNextMsgErr()
			} else {
				err = s.processNextMsgDelivered(firstMsg)
				if err == nil {
					err = checkMsg(firstMsg)
				}
			}
		case <-ctx.Done():
			err = checkCtxErr(ctx.Err())
		}
		if err != nil {
			s.Unsubscribe()
			return nil, err
		}
		// Check again if the delivered next message is a status error.
		err = checkMsg(firstMsg)
		if err != nil {
			s.Unsubscribe()
			return nil, err
		}
	} else {
		// We are receiving messages at this point. Send UNSUB to let
		// the server clear interest once enough replies are delivered.
		err = s.AutoUnsubscribe(batch)
		if err != nil {
			return nil, err
		}
	}

	msgs = append(msgs, firstMsg)
	for {
		var (
			msg *Msg
			ok  bool
		)
		select {
		case msg, ok = <-mch:
			if !ok {
				err = s.getNextMsgErr()
			} else {
				err = s.processNextMsgDelivered(msg)
				if err == nil {
					err = checkMsg(msg)
				}
			}
		case <-ctx.Done():
			return msgs, checkCtxErr(err)
		}
		if err != nil {
			// Discard the error which may have been a timeout
			// or 408 request timeout status from the server,
			// and just the return delivered messages.
			break
		}
		if msg != nil {
			msgs = append(msgs, msg)
		}

		if len(msgs) == batch {
			// Done!
			break
		}
	}

	return msgs, nil
}

func (js *js) getConsumerInfo(stream, consumer string) (*ConsumerInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), js.opts.wait)
	defer cancel()
	return js.getConsumerInfoContext(ctx, stream, consumer)
}

func (js *js) getConsumerInfoContext(ctx context.Context, stream, consumer string) (*ConsumerInfo, error) {
	ccInfoSubj := fmt.Sprintf(apiConsumerInfoT, stream, consumer)
	resp, err := js.nc.RequestWithContext(ctx, js.apiSubj(ccInfoSubj), nil)
	if err != nil {
		if err == ErrNoResponders {
			err = ErrJetStreamNotEnabled
		}
		return nil, err
	}

	var info consumerResponse
	if err := json.Unmarshal(resp.Data, &info); err != nil {
		return nil, err
	}
	if info.Error != nil {
		return nil, fmt.Errorf("nats: %s", info.Error.Description)
	}
	return info.ConsumerInfo, nil
}

func (m *Msg) checkReply() (*js, *jsSub, error) {
	if m == nil || m.Sub == nil {
		return nil, nil, ErrMsgNotBound
	}
	if m.Reply == "" {
		return nil, nil, ErrMsgNoReply
	}
	sub := m.Sub
	sub.mu.Lock()
	if sub.jsi == nil {
		sub.mu.Unlock()

		// Not using a JS context.
		return nil, nil, nil
	}
	js := sub.jsi.js
	jsi := sub.jsi
	sub.mu.Unlock()

	return js, jsi, nil
}

// ackReply handles all acks. Will do the right thing for pull and sync mode.
// It ensures that an ack is only sent a single time, regardless of
// how many times it is being called to avoid duplicated acks.
func (m *Msg) ackReply(ackType []byte, sync bool, opts ...AckOpt) error {
	var o ackOpts
	for _, opt := range opts {
		if err := opt.configureAck(&o); err != nil {
			return err
		}
	}

	js, _, err := m.checkReply()
	if err != nil {
		return err
	}

	// Skip if already acked.
	if atomic.LoadUint32(&m.ackd) == 1 {
		return ErrInvalidJSAck
	}

	m.Sub.mu.Lock()
	nc := m.Sub.conn
	m.Sub.mu.Unlock()

	usesCtx := o.ctx != nil
	usesWait := o.ttl > 0
	sync = sync || usesCtx || usesWait
	ctx := o.ctx
	wait := defaultRequestWait
	if usesWait {
		wait = o.ttl
	} else if js != nil {
		wait = js.opts.wait
	}

	if sync {
		if usesCtx {
			_, err = nc.RequestWithContext(ctx, m.Reply, ackType)
		} else {
			_, err = nc.Request(m.Reply, ackType, wait)
		}
	} else {
		err = nc.Publish(m.Reply, ackType)
	}

	// Mark that the message has been acked unless it is AckProgress
	// which can be sent many times.
	if err == nil && !bytes.Equal(ackType, ackProgress) {
		atomic.StoreUint32(&m.ackd, 1)
	}

	return err
}

// Ack acknowledges a message. This tells the server that the message was
// successfully processed and it can move on to the next message.
func (m *Msg) Ack(opts ...AckOpt) error {
	return m.ackReply(ackAck, false, opts...)
}

// AckSync is the synchronous version of Ack. This indicates successful message
// processing.
func (m *Msg) AckSync(opts ...AckOpt) error {
	return m.ackReply(ackAck, true, opts...)
}

// Nak negatively acknowledges a message. This tells the server to redeliver
// the message. You can configure the number of redeliveries by passing
// nats.MaxDeliver when you Subscribe. The default is infinite redeliveries.
func (m *Msg) Nak(opts ...AckOpt) error {
	return m.ackReply(ackNak, false, opts...)
}

// Term tells the server to not redeliver this message, regardless of the value
// of nats.MaxDeliver.
func (m *Msg) Term(opts ...AckOpt) error {
	return m.ackReply(ackTerm, false, opts...)
}

// InProgress tells the server that this message is being worked on. It resets
// the redelivery timer on the server.
func (m *Msg) InProgress(opts ...AckOpt) error {
	return m.ackReply(ackProgress, false, opts...)
}

// MsgMetadata is the JetStream metadata associated with received messages.
type MsgMetadata struct {
	Sequence     SequencePair
	NumDelivered uint64
	NumPending   uint64
	Timestamp    time.Time
	Stream       string
	Consumer     string
}

func getMetadataFields(subject string) ([]string, error) {
	const expectedTokens = 9
	const btsep = '.'

	tsa := [expectedTokens]string{}
	start, tokens := 0, tsa[:0]
	for i := 0; i < len(subject); i++ {
		if subject[i] == btsep {
			tokens = append(tokens, subject[start:i])
			start = i + 1
		}
	}
	tokens = append(tokens, subject[start:])
	if len(tokens) != expectedTokens || tokens[0] != "$JS" || tokens[1] != "ACK" {
		return nil, ErrNotJSMessage
	}
	return tokens, nil
}

// Metadata retrieves the metadata from a JetStream message. This method will
// return an error for non-JetStream Msgs.
func (m *Msg) Metadata() (*MsgMetadata, error) {
	if _, _, err := m.checkReply(); err != nil {
		return nil, err
	}

	tokens, err := getMetadataFields(m.Reply)
	if err != nil {
		return nil, err
	}

	meta := &MsgMetadata{
		NumDelivered: uint64(parseNum(tokens[4])),
		NumPending:   uint64(parseNum(tokens[8])),
		Timestamp:    time.Unix(0, parseNum(tokens[7])),
		Stream:       tokens[2],
		Consumer:     tokens[3],
	}
	meta.Sequence.Stream = uint64(parseNum(tokens[5]))
	meta.Sequence.Consumer = uint64(parseNum(tokens[6]))
	return meta, nil
}

// Quick parser for positive numbers in ack reply encoding.
func parseNum(d string) (n int64) {
	if len(d) == 0 {
		return -1
	}

	// Ascii numbers 0-9
	const (
		asciiZero = 48
		asciiNine = 57
	)

	for _, dec := range d {
		if dec < asciiZero || dec > asciiNine {
			return -1
		}
		n = n*10 + (int64(dec) - asciiZero)
	}
	return n
}

// AckPolicy determines how the consumer should acknowledge delivered messages.
type AckPolicy int

const (
	// AckNonePolicy requires no acks for delivered messages.
	AckNonePolicy AckPolicy = iota

	// AckAllPolicy when acking a sequence number, this implicitly acks all
	// sequences below this one as well.
	AckAllPolicy

	// AckExplicitPolicy requires ack or nack for all messages.
	AckExplicitPolicy

	// For setting
	ackPolicyNotSet = 99
)

func jsonString(s string) string {
	return "\"" + s + "\""
}

func (p *AckPolicy) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case jsonString("none"):
		*p = AckNonePolicy
	case jsonString("all"):
		*p = AckAllPolicy
	case jsonString("explicit"):
		*p = AckExplicitPolicy
	default:
		return fmt.Errorf("nats: can not unmarshal %q", data)
	}

	return nil
}

func (p AckPolicy) MarshalJSON() ([]byte, error) {
	switch p {
	case AckNonePolicy:
		return json.Marshal("none")
	case AckAllPolicy:
		return json.Marshal("all")
	case AckExplicitPolicy:
		return json.Marshal("explicit")
	default:
		return nil, fmt.Errorf("nats: unknown acknowlegement policy %v", p)
	}
}

func (p AckPolicy) String() string {
	switch p {
	case AckNonePolicy:
		return "AckNone"
	case AckAllPolicy:
		return "AckAll"
	case AckExplicitPolicy:
		return "AckExplicit"
	case ackPolicyNotSet:
		return "Not Initialized"
	default:
		return "Unknown AckPolicy"
	}
}

// ReplayPolicy determines how the consumer should replay messages it already has queued in the stream.
type ReplayPolicy int

const (
	// ReplayInstantPolicy will replay messages as fast as possible.
	ReplayInstantPolicy ReplayPolicy = iota

	// ReplayOriginalPolicy will maintain the same timing as the messages were received.
	ReplayOriginalPolicy
)

func (p *ReplayPolicy) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case jsonString("instant"):
		*p = ReplayInstantPolicy
	case jsonString("original"):
		*p = ReplayOriginalPolicy
	default:
		return fmt.Errorf("nats: can not unmarshal %q", data)
	}

	return nil
}

func (p ReplayPolicy) MarshalJSON() ([]byte, error) {
	switch p {
	case ReplayOriginalPolicy:
		return json.Marshal("original")
	case ReplayInstantPolicy:
		return json.Marshal("instant")
	default:
		return nil, fmt.Errorf("nats: unknown replay policy %v", p)
	}
}

var (
	ackAck      = []byte("+ACK")
	ackNak      = []byte("-NAK")
	ackProgress = []byte("+WPI")
	ackTerm     = []byte("+TERM")
)

// DeliverPolicy determines how the consumer should select the first message to deliver.
type DeliverPolicy int

const (
	// DeliverAllPolicy starts delivering messages from the very beginning of a
	// stream. This is the default.
	DeliverAllPolicy DeliverPolicy = iota

	// DeliverLastPolicy will start the consumer with the last sequence
	// received.
	DeliverLastPolicy

	// DeliverNewPolicy will only deliver new messages that are sent after the
	// consumer is created.
	DeliverNewPolicy

	// DeliverByStartSequencePolicy will deliver messages starting from a given
	// sequence.
	DeliverByStartSequencePolicy

	// DeliverByStartTimePolicy will deliver messages starting from a given
	// time.
	DeliverByStartTimePolicy
)

func (p *DeliverPolicy) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case jsonString("all"), jsonString("undefined"):
		*p = DeliverAllPolicy
	case jsonString("last"):
		*p = DeliverLastPolicy
	case jsonString("new"):
		*p = DeliverNewPolicy
	case jsonString("by_start_sequence"):
		*p = DeliverByStartSequencePolicy
	case jsonString("by_start_time"):
		*p = DeliverByStartTimePolicy
	}

	return nil
}

func (p DeliverPolicy) MarshalJSON() ([]byte, error) {
	switch p {
	case DeliverAllPolicy:
		return json.Marshal("all")
	case DeliverLastPolicy:
		return json.Marshal("last")
	case DeliverNewPolicy:
		return json.Marshal("new")
	case DeliverByStartSequencePolicy:
		return json.Marshal("by_start_sequence")
	case DeliverByStartTimePolicy:
		return json.Marshal("by_start_time")
	default:
		return nil, fmt.Errorf("nats: unknown deliver policy %v", p)
	}
}

// RetentionPolicy determines how messages in a set are retained.
type RetentionPolicy int

const (
	// LimitsPolicy (default) means that messages are retained until any given limit is reached.
	// This could be one of MaxMsgs, MaxBytes, or MaxAge.
	LimitsPolicy RetentionPolicy = iota
	// InterestPolicy specifies that when all known observables have acknowledged a message it can be removed.
	InterestPolicy
	// WorkQueuePolicy specifies that when the first worker or subscriber acknowledges the message it can be removed.
	WorkQueuePolicy
)

// DiscardPolicy determines how to proceed when limits of messages or bytes are
// reached.
type DiscardPolicy int

const (
	// DiscardOld will remove older messages to return to the limits. This is
	// the default.
	DiscardOld DiscardPolicy = iota
	//DiscardNew will fail to store new messages.
	DiscardNew
)

const (
	limitsPolicyString    = "limits"
	interestPolicyString  = "interest"
	workQueuePolicyString = "workqueue"
)

func (rp RetentionPolicy) String() string {
	switch rp {
	case LimitsPolicy:
		return "Limits"
	case InterestPolicy:
		return "Interest"
	case WorkQueuePolicy:
		return "WorkQueue"
	default:
		return "Unknown Retention Policy"
	}
}

func (rp RetentionPolicy) MarshalJSON() ([]byte, error) {
	switch rp {
	case LimitsPolicy:
		return json.Marshal(limitsPolicyString)
	case InterestPolicy:
		return json.Marshal(interestPolicyString)
	case WorkQueuePolicy:
		return json.Marshal(workQueuePolicyString)
	default:
		return nil, fmt.Errorf("nats: can not marshal %v", rp)
	}
}

func (rp *RetentionPolicy) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case jsonString(limitsPolicyString):
		*rp = LimitsPolicy
	case jsonString(interestPolicyString):
		*rp = InterestPolicy
	case jsonString(workQueuePolicyString):
		*rp = WorkQueuePolicy
	default:
		return fmt.Errorf("nats: can not unmarshal %q", data)
	}
	return nil
}

func (dp DiscardPolicy) String() string {
	switch dp {
	case DiscardOld:
		return "DiscardOld"
	case DiscardNew:
		return "DiscardNew"
	default:
		return "Unknown Discard Policy"
	}
}

func (dp DiscardPolicy) MarshalJSON() ([]byte, error) {
	switch dp {
	case DiscardOld:
		return json.Marshal("old")
	case DiscardNew:
		return json.Marshal("new")
	default:
		return nil, fmt.Errorf("nats: can not marshal %v", dp)
	}
}

func (dp *DiscardPolicy) UnmarshalJSON(data []byte) error {
	switch strings.ToLower(string(data)) {
	case jsonString("old"):
		*dp = DiscardOld
	case jsonString("new"):
		*dp = DiscardNew
	default:
		return fmt.Errorf("nats: can not unmarshal %q", data)
	}
	return nil
}

// StorageType determines how messages are stored for retention.
type StorageType int

const (
	// FileStorage specifies on disk storage. It's the default.
	FileStorage StorageType = iota
	// MemoryStorage specifies in memory only.
	MemoryStorage
)

const (
	memoryStorageString = "memory"
	fileStorageString   = "file"
)

func (st StorageType) String() string {
	switch st {
	case MemoryStorage:
		return strings.Title(memoryStorageString)
	case FileStorage:
		return strings.Title(fileStorageString)
	default:
		return "Unknown Storage Type"
	}
}

func (st StorageType) MarshalJSON() ([]byte, error) {
	switch st {
	case MemoryStorage:
		return json.Marshal(memoryStorageString)
	case FileStorage:
		return json.Marshal(fileStorageString)
	default:
		return nil, fmt.Errorf("nats: can not marshal %v", st)
	}
}

func (st *StorageType) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case jsonString(memoryStorageString):
		*st = MemoryStorage
	case jsonString(fileStorageString):
		*st = FileStorage
	default:
		return fmt.Errorf("nats: can not unmarshal %q", data)
	}
	return nil
}
