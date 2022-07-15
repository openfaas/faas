// Copyright 2020-2022 The NATS Authors
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

// JetStream allows persistent messaging through JetStream.
type JetStream interface {
	// Publish publishes a message to JetStream.
	Publish(subj string, data []byte, opts ...PubOpt) (*PubAck, error)

	// PublishMsg publishes a Msg to JetStream.
	PublishMsg(m *Msg, opts ...PubOpt) (*PubAck, error)

	// PublishAsync publishes a message to JetStream and returns a PubAckFuture.
	// The data should not be changed until the PubAckFuture has been processed.
	PublishAsync(subj string, data []byte, opts ...PubOpt) (PubAckFuture, error)

	// PublishMsgAsync publishes a Msg to JetStream and returns a PubAckFuture.
	// The message should not be changed until the PubAckFuture has been processed.
	PublishMsgAsync(m *Msg, opts ...PubOpt) (PubAckFuture, error)

	// PublishAsyncPending returns the number of async publishes outstanding for this context.
	PublishAsyncPending() int

	// PublishAsyncComplete returns a channel that will be closed when all outstanding messages are ack'd.
	PublishAsyncComplete() <-chan struct{}

	// Subscribe creates an async Subscription for JetStream.
	// The stream and consumer names can be provided with the nats.Bind() option.
	// For creating an ephemeral (where the consumer name is picked by the server),
	// you can provide the stream name with nats.BindStream().
	// If no stream name is specified, the library will attempt to figure out which
	// stream the subscription is for. See important notes below for more details.
	//
	// IMPORTANT NOTES:
	// * If none of the options Bind() nor Durable() are specified, the library will
	// send a request to the server to create an ephemeral JetStream consumer,
	// which will be deleted after an Unsubscribe() or Drain(), or automatically
	// by the server after a short period of time after the NATS subscription is
	// gone.
	// * If Durable() option is specified, the library will attempt to lookup a JetStream
	// consumer with this name, and if found, will bind to it and not attempt to
	// delete it. However, if not found, the library will send a request to create
	// such durable JetStream consumer. The library will delete the JetStream consumer
	// after an Unsubscribe() or Drain().
	// * If Bind() option is provided, the library will attempt to lookup the
	// consumer with the given name, and if successful, bind to it. If the lookup fails,
	// then the Subscribe() call will return an error.
	Subscribe(subj string, cb MsgHandler, opts ...SubOpt) (*Subscription, error)

	// SubscribeSync creates a Subscription that can be used to process messages synchronously.
	// See important note in Subscribe()
	SubscribeSync(subj string, opts ...SubOpt) (*Subscription, error)

	// ChanSubscribe creates channel based Subscription.
	// See important note in Subscribe()
	ChanSubscribe(subj string, ch chan *Msg, opts ...SubOpt) (*Subscription, error)

	// ChanQueueSubscribe creates channel based Subscription with a queue group.
	// See important note in QueueSubscribe()
	ChanQueueSubscribe(subj, queue string, ch chan *Msg, opts ...SubOpt) (*Subscription, error)

	// QueueSubscribe creates a Subscription with a queue group.
	// If no optional durable name nor binding options are specified, the queue name will be used as a durable name.
	// See important note in Subscribe()
	QueueSubscribe(subj, queue string, cb MsgHandler, opts ...SubOpt) (*Subscription, error)

	// QueueSubscribeSync creates a Subscription with a queue group that can be used to process messages synchronously.
	// See important note in QueueSubscribe()
	QueueSubscribeSync(subj, queue string, opts ...SubOpt) (*Subscription, error)

	// PullSubscribe creates a Subscription that can fetch messages.
	// See important note in Subscribe()
	PullSubscribe(subj, durable string, opts ...SubOpt) (*Subscription, error)
}

// JetStreamContext allows JetStream messaging and stream management.
type JetStreamContext interface {
	JetStream
	JetStreamManager
	KeyValueManager
	ObjectStoreManager
}

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

	// apiConsumerDeleteT is used to delete consumers.
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

	// apiStreamUpdateT is the endpoint to update existing streams.
	apiStreamUpdateT = "STREAM.UPDATE.%s"

	// apiStreamDeleteT is the endpoint to delete streams.
	apiStreamDeleteT = "STREAM.DELETE.%s"

	// apiStreamPurgeT is the endpoint to purge streams.
	apiStreamPurgeT = "STREAM.PURGE.%s"

	// apiStreamListT is the endpoint that will return all detailed stream information
	apiStreamListT = "STREAM.LIST"

	// apiMsgGetT is the endpoint to get a message.
	apiMsgGetT = "STREAM.MSG.GET.%s"

	// apiMsgDeleteT is the endpoint to remove a message.
	apiMsgDeleteT = "STREAM.MSG.DELETE.%s"

	// orderedHeartbeatsInterval is how fast we want HBs from the server during idle.
	orderedHeartbeatsInterval = 5 * time.Second

	// Scale for threshold of missed HBs or lack of activity.
	hbcThresh = 2

	// For ChanSubscription, we can't update sub.delivered as we do for other
	// type of subscriptions, since the channel is user provided.
	// With flow control in play, we will check for flow control on incoming
	// messages (as opposed to when they are delivered), but also from a go
	// routine. Without this, the subscription would possibly stall until
	// a new message or heartbeat/fc are received.
	chanSubFCCheckInterval = 250 * time.Millisecond

	// Default time wait between retries on Publish iff err is NoResponders.
	DefaultPubRetryWait = 250 * time.Millisecond

	// Default number of retries
	DefaultPubRetryAttempts = 2

	// defaultAsyncPubAckInflight is the number of async pub acks inflight.
	defaultAsyncPubAckInflight = 4000
)

// Types of control messages, so far heartbeat and flow control
const (
	jsCtrlHB = 1
	jsCtrlFC = 2
)

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
	// Max async pub ack in flight
	maxpa int
	// the domain that produced the pre
	domain string
	// enables protocol tracing
	ctrace      ClientTrace
	shouldTrace bool
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
			pre:   defaultAPIPrefix,
			wait:  defaultRequestWait,
			maxpa: defaultAsyncPubAckInflight,
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

// ClientTrace can be used to trace API interactions for the JetStream Context.
type ClientTrace struct {
	RequestSent      func(subj string, payload []byte)
	ResponseReceived func(subj string, payload []byte, hdr Header)
}

func (ct ClientTrace) configureJSContext(js *jsOpts) error {
	js.ctrace = ct
	js.shouldTrace = true
	return nil
}

// Domain changes the domain part of JetStream API prefix.
func Domain(domain string) JSOpt {
	if domain == _EMPTY_ {
		return APIPrefix(_EMPTY_)
	}

	return jsOptFn(func(js *jsOpts) error {
		js.domain = domain
		js.pre = fmt.Sprintf(jsDomainT, domain)

		return nil
	})

}

// APIPrefix changes the default prefix used for the JetStream API.
func APIPrefix(pre string) JSOpt {
	return jsOptFn(func(js *jsOpts) error {
		if pre == _EMPTY_ {
			return nil
		}

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
	lid string  // Expected last msgId
	str string  // Expected stream name
	seq *uint64 // Expected last sequence
	lss *uint64 // Expected last sequence per subject

	// Publish retries for NoResponders err.
	rwait time.Duration // Retry wait between attempts
	rnum  int           // Retry attempts

	// stallWait is the max wait of a async pub ack.
	stallWait time.Duration
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
	Domain    string `json:"domain,omitempty"`
}

// Headers for published messages.
const (
	MsgIdHdr               = "Nats-Msg-Id"
	ExpectedStreamHdr      = "Nats-Expected-Stream"
	ExpectedLastSeqHdr     = "Nats-Expected-Last-Sequence"
	ExpectedLastSubjSeqHdr = "Nats-Expected-Last-Subject-Sequence"
	ExpectedLastMsgIdHdr   = "Nats-Expected-Last-Msg-Id"
	MsgRollup              = "Nats-Rollup"
)

// Headers for republished messages.
const (
	JSStream       = "Nats-Stream"
	JSSequence     = "Nats-Sequence"
	JSLastSequence = "Nats-Last-Sequence"
)

// MsgSize is a header that will be part of a consumer's delivered message if HeadersOnly requested.
const MsgSize = "Nats-Msg-Size"

// Rollups, can be subject only or all messages.
const (
	MsgRollupSubject = "sub"
	MsgRollupAll     = "all"
)

// PublishMsg publishes a Msg to a stream from JetStream.
func (js *js) PublishMsg(m *Msg, opts ...PubOpt) (*PubAck, error) {
	var o = pubOpts{rwait: DefaultPubRetryWait, rnum: DefaultPubRetryAttempts}
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
	if o.stallWait > 0 {
		return nil, fmt.Errorf("nats: stall wait cannot be set to sync publish")
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
	if o.seq != nil {
		m.Header.Set(ExpectedLastSeqHdr, strconv.FormatUint(*o.seq, 10))
	}
	if o.lss != nil {
		m.Header.Set(ExpectedLastSubjSeqHdr, strconv.FormatUint(*o.lss, 10))
	}

	var resp *Msg
	var err error

	if o.ttl > 0 {
		resp, err = js.nc.RequestMsg(m, time.Duration(o.ttl))
	} else {
		resp, err = js.nc.RequestMsgWithContext(o.ctx, m)
	}

	if err != nil {
		for r, ttl := 0, o.ttl; err == ErrNoResponders && (r < o.rnum || o.rnum < 0); r++ {
			// To protect against small blips in leadership changes etc, if we get a no responders here retry.
			if o.ctx != nil {
				select {
				case <-o.ctx.Done():
				case <-time.After(o.rwait):
				}
			} else {
				time.Sleep(o.rwait)
			}
			if o.ttl > 0 {
				ttl -= o.rwait
				if ttl <= 0 {
					err = ErrTimeout
					break
				}
				resp, err = js.nc.RequestMsg(m, time.Duration(ttl))
			} else {
				resp, err = js.nc.RequestMsgWithContext(o.ctx, m)
			}
		}
		if err != nil {
			if err == ErrNoResponders {
				err = ErrNoStreamResponse
			}
			return nil, err
		}
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
	maxpa := js.opts.maxpa
	js.mu.Unlock()
	return np, maxpa
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
	if js.stc != nil && len(js.pafs) < js.opts.maxpa {
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
// JetStream PublishAsync. It will return the original
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
		js.maxpa = max
		return nil
	})
}

// PublishAsync publishes a message to JetStream and returns a PubAckFuture
func (js *js) PublishAsync(subj string, data []byte, opts ...PubOpt) (PubAckFuture, error) {
	return js.PublishMsgAsync(&Msg{Subject: subj, Data: data}, opts...)
}

const defaultStallWait = 200 * time.Millisecond

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
	stallWait := defaultStallWait
	if o.stallWait > 0 {
		stallWait = o.stallWait
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
	if o.seq != nil {
		m.Header.Set(ExpectedLastSeqHdr, strconv.FormatUint(*o.seq, 10))
	}
	if o.lss != nil {
		m.Header.Set(ExpectedLastSubjSeqHdr, strconv.FormatUint(*o.lss, 10))
	}

	// Reply
	if m.Reply != _EMPTY_ {
		return nil, errors.New("nats: reply subject should be empty")
	}
	reply := m.Reply
	m.Reply = js.newAsyncReply()
	defer func() { m.Reply = reply }()

	if m.Reply == _EMPTY_ {
		return nil, errors.New("nats: error creating async reply handler")
	}

	id := m.Reply[aReplyPreLen:]
	paf := &pubAckFuture{msg: m, st: time.Now()}
	numPending, maxPending := js.registerPAF(id, paf)

	if maxPending > 0 && numPending >= maxPending {
		select {
		case <-js.asyncStall():
		case <-time.After(stallWait):
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

// MsgId sets the message ID used for deduplication.
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
		opts.seq = &seq
		return nil
	})
}

// ExpectLastSequencePerSubject sets the expected sequence per subject in the response from the publish.
func ExpectLastSequencePerSubject(seq uint64) PubOpt {
	return pubOptFn(func(opts *pubOpts) error {
		opts.lss = &seq
		return nil
	})
}

// ExpectLastMsgId sets the expected last msgId in the response from the publish.
func ExpectLastMsgId(id string) PubOpt {
	return pubOptFn(func(opts *pubOpts) error {
		opts.lid = id
		return nil
	})
}

// RetryWait sets the retry wait time when ErrNoResponders is encountered.
func RetryWait(dur time.Duration) PubOpt {
	return pubOptFn(func(opts *pubOpts) error {
		opts.rwait = dur
		return nil
	})
}

// RetryAttempts sets the retry number of attempts when ErrNoResponders is encountered.
func RetryAttempts(num int) PubOpt {
	return pubOptFn(func(opts *pubOpts) error {
		opts.rnum = num
		return nil
	})
}

// StallWait sets the max wait when the producer becomes stall producing messages.
func StallWait(ttl time.Duration) PubOpt {
	return pubOptFn(func(opts *pubOpts) error {
		if ttl <= 0 {
			return fmt.Errorf("nats: stall wait should be more than 0")
		}
		opts.stallWait = ttl
		return nil
	})
}

type ackOpts struct {
	ttl      time.Duration
	ctx      context.Context
	nakDelay time.Duration
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

func (ctx ContextOpt) configureSubscribe(opts *subOpts) error {
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

type nakDelay time.Duration

func (d nakDelay) configureAck(opts *ackOpts) error {
	opts.nakDelay = time.Duration(d)
	return nil
}

// Subscribe

// ConsumerConfig is the configuration of a JetStream consumer.
type ConsumerConfig struct {
	Durable         string          `json:"durable_name,omitempty"`
	Description     string          `json:"description,omitempty"`
	DeliverPolicy   DeliverPolicy   `json:"deliver_policy"`
	OptStartSeq     uint64          `json:"opt_start_seq,omitempty"`
	OptStartTime    *time.Time      `json:"opt_start_time,omitempty"`
	AckPolicy       AckPolicy       `json:"ack_policy"`
	AckWait         time.Duration   `json:"ack_wait,omitempty"`
	MaxDeliver      int             `json:"max_deliver,omitempty"`
	BackOff         []time.Duration `json:"backoff,omitempty"`
	FilterSubject   string          `json:"filter_subject,omitempty"`
	ReplayPolicy    ReplayPolicy    `json:"replay_policy"`
	RateLimit       uint64          `json:"rate_limit_bps,omitempty"` // Bits per sec
	SampleFrequency string          `json:"sample_freq,omitempty"`
	MaxWaiting      int             `json:"max_waiting,omitempty"`
	MaxAckPending   int             `json:"max_ack_pending,omitempty"`
	FlowControl     bool            `json:"flow_control,omitempty"`
	Heartbeat       time.Duration   `json:"idle_heartbeat,omitempty"`
	HeadersOnly     bool            `json:"headers_only,omitempty"`

	// Pull based options.
	MaxRequestBatch   int           `json:"max_batch,omitempty"`
	MaxRequestExpires time.Duration `json:"max_expires,omitempty"`

	// Push based consumers.
	DeliverSubject string `json:"deliver_subject,omitempty"`
	DeliverGroup   string `json:"deliver_group,omitempty"`

	// Ephemeral inactivity threshold.
	InactiveThreshold time.Duration `json:"inactive_threshold,omitempty"`

	// Generally inherited by parent stream and other markers, now can be configured directly.
	Replicas int `json:"num_replicas"`
	// Force memory storage.
	MemoryStorage bool `json:"mem_storage,omitempty"`
}

// ConsumerInfo is the info from a JetStream consumer.
type ConsumerInfo struct {
	Stream         string         `json:"stream_name"`
	Name           string         `json:"name"`
	Created        time.Time      `json:"created"`
	Config         ConsumerConfig `json:"config"`
	Delivered      SequenceInfo   `json:"delivered"`
	AckFloor       SequenceInfo   `json:"ack_floor"`
	NumAckPending  int            `json:"num_ack_pending"`
	NumRedelivered int            `json:"num_redelivered"`
	NumWaiting     int            `json:"num_waiting"`
	NumPending     uint64         `json:"num_pending"`
	Cluster        *ClusterInfo   `json:"cluster,omitempty"`
	PushBound      bool           `json:"push_bound,omitempty"`
}

// SequenceInfo has both the consumer and the stream sequence and last activity.
type SequenceInfo struct {
	Consumer uint64     `json:"consumer_seq"`
	Stream   uint64     `json:"stream_seq"`
	Last     *time.Time `json:"last_active,omitempty"`
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

	// For pull subscribers, this is the next message subject to send requests to.
	nms string

	psubj    string // the subject that was passed by user to the subscribe calls
	consumer string
	stream   string
	deliver  string
	pull     bool
	dc       bool // Delete JS consumer

	// This is ConsumerInfo's Pending+Consumer.Delivered that we get from the
	// add consumer response. Note that some versions of the server gather the
	// consumer info *after* the creation of the consumer, which means that
	// some messages may have been already delivered. So the sum of the two
	// is a more accurate representation of the number of messages pending or
	// in the process of being delivered to the subscription when created.
	pending uint64

	// Ordered consumers
	ordered bool
	dseq    uint64
	sseq    uint64
	ccreq   *createConsumerRequest

	// Heartbeats and Flow Control handling from push consumers.
	hbc    *time.Timer
	hbi    time.Duration
	active bool
	cmeta  string
	fcr    string
	fcd    uint64
	fciseq uint64
	csfct  *time.Timer

	// Cancellation function to cancel context on drain/unsubscribe.
	cancel func()
}

// Deletes the JS Consumer.
// No connection nor subscription lock must be held on entry.
func (sub *Subscription) deleteConsumer() error {
	sub.mu.Lock()
	jsi := sub.jsi
	if jsi == nil {
		sub.mu.Unlock()
		return nil
	}
	stream, consumer := jsi.stream, jsi.consumer
	js := jsi.js
	sub.mu.Unlock()

	return js.DeleteConsumer(stream, consumer)
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

// Subscribe creates an async Subscription for JetStream.
// The stream and consumer names can be provided with the nats.Bind() option.
// For creating an ephemeral (where the consumer name is picked by the server),
// you can provide the stream name with nats.BindStream().
// If no stream name is specified, the library will attempt to figure out which
// stream the subscription is for. See important notes below for more details.
//
// IMPORTANT NOTES:
// * If none of the options Bind() nor Durable() are specified, the library will
// send a request to the server to create an ephemeral JetStream consumer,
// which will be deleted after an Unsubscribe() or Drain(), or automatically
// by the server after a short period of time after the NATS subscription is
// gone.
// * If Durable() option is specified, the library will attempt to lookup a JetStream
// consumer with this name, and if found, will bind to it and not attempt to
// delete it. However, if not found, the library will send a request to create
// such durable JetStream consumer. The library will delete the JetStream consumer
// after an Unsubscribe() or Drain().
// * If Bind() option is provided, the library will attempt to lookup the
// consumer with the given name, and if successful, bind to it. If the lookup fails,
// then the Subscribe() call will return an error.
func (js *js) Subscribe(subj string, cb MsgHandler, opts ...SubOpt) (*Subscription, error) {
	if cb == nil {
		return nil, ErrBadSubscription
	}
	return js.subscribe(subj, _EMPTY_, cb, nil, false, false, opts)
}

// SubscribeSync creates a Subscription that can be used to process messages synchronously.
// See important note in Subscribe()
func (js *js) SubscribeSync(subj string, opts ...SubOpt) (*Subscription, error) {
	mch := make(chan *Msg, js.nc.Opts.SubChanLen)
	return js.subscribe(subj, _EMPTY_, nil, mch, true, false, opts)
}

// QueueSubscribe creates a Subscription with a queue group.
// If no optional durable name nor binding options are specified, the queue name will be used as a durable name.
// See important note in Subscribe()
func (js *js) QueueSubscribe(subj, queue string, cb MsgHandler, opts ...SubOpt) (*Subscription, error) {
	if cb == nil {
		return nil, ErrBadSubscription
	}
	return js.subscribe(subj, queue, cb, nil, false, false, opts)
}

// QueueSubscribeSync creates a Subscription with a queue group that can be used to process messages synchronously.
// See important note in QueueSubscribe()
func (js *js) QueueSubscribeSync(subj, queue string, opts ...SubOpt) (*Subscription, error) {
	mch := make(chan *Msg, js.nc.Opts.SubChanLen)
	return js.subscribe(subj, queue, nil, mch, true, false, opts)
}

// ChanSubscribe creates channel based Subscription.
// See important note in Subscribe()
func (js *js) ChanSubscribe(subj string, ch chan *Msg, opts ...SubOpt) (*Subscription, error) {
	return js.subscribe(subj, _EMPTY_, nil, ch, false, false, opts)
}

// ChanQueueSubscribe creates channel based Subscription with a queue group.
// See important note in QueueSubscribe()
func (js *js) ChanQueueSubscribe(subj, queue string, ch chan *Msg, opts ...SubOpt) (*Subscription, error) {
	return js.subscribe(subj, queue, nil, ch, false, false, opts)
}

// PullSubscribe creates a Subscription that can fetch messages.
// See important note in Subscribe()
func (js *js) PullSubscribe(subj, durable string, opts ...SubOpt) (*Subscription, error) {
	mch := make(chan *Msg, js.nc.Opts.SubChanLen)
	return js.subscribe(subj, _EMPTY_, nil, mch, true, true, append(opts, Durable(durable)))
}

func processConsInfo(info *ConsumerInfo, userCfg *ConsumerConfig, isPullMode bool, subj, queue string) (string, error) {
	ccfg := &info.Config

	// Make sure this new subject matches or is a subset.
	if ccfg.FilterSubject != _EMPTY_ && subj != ccfg.FilterSubject {
		return _EMPTY_, ErrSubjectMismatch
	}

	// Prevent binding a subscription against incompatible consumer types.
	if isPullMode && ccfg.DeliverSubject != _EMPTY_ {
		return _EMPTY_, ErrPullSubscribeToPushConsumer
	} else if !isPullMode && ccfg.DeliverSubject == _EMPTY_ {
		return _EMPTY_, ErrPullSubscribeRequired
	}

	// If pull mode, nothing else to check here.
	if isPullMode {
		return _EMPTY_, checkConfig(ccfg, userCfg)
	}

	// At this point, we know the user wants push mode, and the JS consumer is
	// really push mode.

	dg := info.Config.DeliverGroup
	if dg == _EMPTY_ {
		// Prevent an user from attempting to create a queue subscription on
		// a JS consumer that was not created with a deliver group.
		if queue != _EMPTY_ {
			return _EMPTY_, fmt.Errorf("cannot create a queue subscription for a consumer without a deliver group")
		} else if info.PushBound {
			// Need to reject a non queue subscription to a non queue consumer
			// if the consumer is already bound.
			return _EMPTY_, fmt.Errorf("consumer is already bound to a subscription")
		}
	} else {
		// If the JS consumer has a deliver group, we need to fail a non queue
		// subscription attempt:
		if queue == _EMPTY_ {
			return _EMPTY_, fmt.Errorf("cannot create a subscription for a consumer with a deliver group %q", dg)
		} else if queue != dg {
			// Here the user's queue group name does not match the one associated
			// with the JS consumer.
			return _EMPTY_, fmt.Errorf("cannot create a queue subscription %q for a consumer with a deliver group %q",
				queue, dg)
		}
	}
	if err := checkConfig(ccfg, userCfg); err != nil {
		return _EMPTY_, err
	}
	return ccfg.DeliverSubject, nil
}

func checkConfig(s, u *ConsumerConfig) error {
	makeErr := func(fieldName string, usrVal, srvVal interface{}) error {
		return fmt.Errorf("configuration requests %s to be %v, but consumer's value is %v", fieldName, usrVal, srvVal)
	}

	if u.Durable != _EMPTY_ && u.Durable != s.Durable {
		return makeErr("durable", u.Durable, s.Durable)
	}
	if u.Description != _EMPTY_ && u.Description != s.Description {
		return makeErr("description", u.Description, s.Description)
	}
	if u.DeliverPolicy != deliverPolicyNotSet && u.DeliverPolicy != s.DeliverPolicy {
		return makeErr("deliver policy", u.DeliverPolicy, s.DeliverPolicy)
	}
	if u.OptStartSeq > 0 && u.OptStartSeq != s.OptStartSeq {
		return makeErr("optional start sequence", u.OptStartSeq, s.OptStartSeq)
	}
	if u.OptStartTime != nil && !u.OptStartTime.IsZero() && u.OptStartTime != s.OptStartTime {
		return makeErr("optional start time", u.OptStartTime, s.OptStartTime)
	}
	if u.AckPolicy != ackPolicyNotSet && u.AckPolicy != s.AckPolicy {
		return makeErr("ack policy", u.AckPolicy, s.AckPolicy)
	}
	if u.AckWait > 0 && u.AckWait != s.AckWait {
		return makeErr("ack wait", u.AckWait, s.AckWait)
	}
	if u.MaxDeliver > 0 && u.MaxDeliver != s.MaxDeliver {
		return makeErr("max deliver", u.MaxDeliver, s.MaxDeliver)
	}
	if u.ReplayPolicy != replayPolicyNotSet && u.ReplayPolicy != s.ReplayPolicy {
		return makeErr("replay policy", u.ReplayPolicy, s.ReplayPolicy)
	}
	if u.RateLimit > 0 && u.RateLimit != s.RateLimit {
		return makeErr("rate limit", u.RateLimit, s.RateLimit)
	}
	if u.SampleFrequency != _EMPTY_ && u.SampleFrequency != s.SampleFrequency {
		return makeErr("sample frequency", u.SampleFrequency, s.SampleFrequency)
	}
	if u.MaxWaiting > 0 && u.MaxWaiting != s.MaxWaiting {
		return makeErr("max waiting", u.MaxWaiting, s.MaxWaiting)
	}
	if u.MaxAckPending > 0 && u.MaxAckPending != s.MaxAckPending {
		return makeErr("max ack pending", u.MaxAckPending, s.MaxAckPending)
	}
	// For flow control, we want to fail if the user explicit wanted it, but
	// it is not set in the existing consumer. If it is not asked by the user,
	// the library still handles it and so no reason to fail.
	if u.FlowControl && !s.FlowControl {
		return makeErr("flow control", u.FlowControl, s.FlowControl)
	}
	if u.Heartbeat > 0 && u.Heartbeat != s.Heartbeat {
		return makeErr("heartbeat", u.Heartbeat, s.Heartbeat)
	}
	return nil
}

func (js *js) subscribe(subj, queue string, cb MsgHandler, ch chan *Msg, isSync, isPullMode bool, opts []SubOpt) (*Subscription, error) {
	cfg := ConsumerConfig{
		DeliverPolicy: deliverPolicyNotSet,
		AckPolicy:     ackPolicyNotSet,
		ReplayPolicy:  replayPolicyNotSet,
	}
	o := subOpts{cfg: &cfg}
	if len(opts) > 0 {
		for _, opt := range opts {
			if opt == nil {
				continue
			}
			if err := opt.configureSubscribe(&o); err != nil {
				return nil, err
			}
		}
	}

	// If no stream name is specified, the subject cannot be empty.
	if subj == _EMPTY_ && o.stream == _EMPTY_ {
		return nil, fmt.Errorf("nats: subject required")
	}

	// Note that these may change based on the consumer info response we may get.
	hasHeartbeats := o.cfg.Heartbeat > 0
	hasFC := o.cfg.FlowControl

	// Some checks for pull subscribers
	if isPullMode {
		// Check for bad ack policy
		if o.cfg.AckPolicy == AckNonePolicy || o.cfg.AckPolicy == AckAllPolicy {
			return nil, fmt.Errorf("nats: invalid ack mode for pull consumers: %s", o.cfg.AckPolicy)
		}
		// No deliver subject should be provided
		if o.cfg.DeliverSubject != _EMPTY_ {
			return nil, ErrPullSubscribeToPushConsumer
		}
	}

	// Some check/setting specific to queue subs
	if queue != _EMPTY_ {
		// Queue subscriber cannot have HB or FC (since messages will be randomly dispatched
		// to members). We may in the future have a separate NATS subscription that all members
		// would subscribe to and server would send on.
		if o.cfg.Heartbeat > 0 || o.cfg.FlowControl {
			// Not making this a public ErrXXX in case we allow in the future.
			return nil, fmt.Errorf("nats: queue subscription doesn't support idle heartbeat nor flow control")
		}

		// If this is a queue subscription and no consumer nor durable name was specified,
		// then we will use the queue name as a durable name.
		if o.consumer == _EMPTY_ && o.cfg.Durable == _EMPTY_ {
			if err := checkDurName(queue); err != nil {
				return nil, err
			}
			o.cfg.Durable = queue
		}
	}

	var (
		err           error
		shouldCreate  bool
		info          *ConsumerInfo
		deliver       string
		stream        = o.stream
		consumer      = o.consumer
		isDurable     = o.cfg.Durable != _EMPTY_
		consumerBound = o.bound
		ctx           = o.ctx
		notFoundErr   bool
		lookupErr     bool
		nc            = js.nc
		nms           string
		hbi           time.Duration
		ccreq         *createConsumerRequest // In case we need to hold onto it for ordered consumers.
		maxap         int
	)

	// Do some quick checks here for ordered consumers. We do these here instead of spread out
	// in the individual SubOpts.
	if o.ordered {
		// Make sure we are not durable.
		if isDurable {
			return nil, fmt.Errorf("nats: durable can not be set for an ordered consumer")
		}
		// Check ack policy.
		if o.cfg.AckPolicy != ackPolicyNotSet {
			return nil, fmt.Errorf("nats: ack policy can not be set for an ordered consumer")
		}
		// Check max deliver.
		if o.cfg.MaxDeliver != 1 && o.cfg.MaxDeliver != 0 {
			return nil, fmt.Errorf("nats: max deliver can not be set for an ordered consumer")
		}
		// No deliver subject, we pick our own.
		if o.cfg.DeliverSubject != _EMPTY_ {
			return nil, fmt.Errorf("nats: deliver subject can not be set for an ordered consumer")
		}
		// Queue groups not allowed.
		if queue != _EMPTY_ {
			return nil, fmt.Errorf("nats: queues not be set for an ordered consumer")
		}
		// Check for bound consumers.
		if consumer != _EMPTY_ {
			return nil, fmt.Errorf("nats: can not bind existing consumer for an ordered consumer")
		}
		// Check for pull mode.
		if isPullMode {
			return nil, fmt.Errorf("nats: can not use pull mode for an ordered consumer")
		}
		// Setup how we need it to be here.
		o.cfg.FlowControl = true
		o.cfg.AckPolicy = AckNonePolicy
		o.cfg.MaxDeliver = 1
		o.cfg.AckWait = 22 * time.Hour // Just set to something known, not utilized.
		if !hasHeartbeats {
			o.cfg.Heartbeat = orderedHeartbeatsInterval
		}
		hasFC, hasHeartbeats = true, true
		o.mack = true // To avoid auto-ack wrapping call below.
		hbi = o.cfg.Heartbeat
	}

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

	// With an explicit durable name, we can lookup the consumer first
	// to which it should be attaching to.
	// If bind to ordered consumer is true, skip the lookup.
	if consumer != _EMPTY_ {
		info, err = js.ConsumerInfo(stream, consumer)
		notFoundErr = errors.Is(err, ErrConsumerNotFound)
		lookupErr = err == ErrJetStreamNotEnabled || err == ErrTimeout || err == context.DeadlineExceeded
	}

	switch {
	case info != nil:
		deliver, err = processConsInfo(info, o.cfg, isPullMode, subj, queue)
		if err != nil {
			return nil, err
		}
		icfg := &info.Config
		hasFC, hbi = icfg.FlowControl, icfg.Heartbeat
		hasHeartbeats = hbi > 0
		maxap = icfg.MaxAckPending
	case (err != nil && !notFoundErr) || (notFoundErr && consumerBound):
		// If the consumer is being bound and we got an error on pull subscribe then allow the error.
		if !(isPullMode && lookupErr && consumerBound) {
			return nil, err
		}
	default:
		// Attempt to create consumer if not found nor using Bind.
		shouldCreate = true
		if o.cfg.DeliverSubject != _EMPTY_ {
			deliver = o.cfg.DeliverSubject
		} else if !isPullMode {
			deliver = nc.newInbox()
			cfg.DeliverSubject = deliver
		}

		// Do filtering always, server will clear as needed.
		cfg.FilterSubject = subj

		// Pass the queue to the consumer config
		if queue != _EMPTY_ {
			cfg.DeliverGroup = queue
		}

		// If not set, default to deliver all
		if cfg.DeliverPolicy == deliverPolicyNotSet {
			cfg.DeliverPolicy = DeliverAllPolicy
		}
		// If not set, default to ack explicit.
		if cfg.AckPolicy == ackPolicyNotSet {
			cfg.AckPolicy = AckExplicitPolicy
		}
		// If not set, default to instant
		if cfg.ReplayPolicy == replayPolicyNotSet {
			cfg.ReplayPolicy = ReplayInstantPolicy
		}

		// If we have acks at all and the MaxAckPending is not set go ahead
		// and set to the internal max for channel based consumers
		if cfg.MaxAckPending == 0 && ch != nil && cfg.AckPolicy != AckNonePolicy {
			cfg.MaxAckPending = cap(ch)
		}
		// Create request here.
		ccreq = &createConsumerRequest{
			Stream: stream,
			Config: &cfg,
		}
		hbi = cfg.Heartbeat
	}

	if isPullMode {
		nms = fmt.Sprintf(js.apiSubj(apiRequestNextT), stream, consumer)
		deliver = nc.newInbox()
	}

	// In case this has a context, then create a child context that
	// is possible to cancel via unsubscribe / drain.
	var cancel func()
	if ctx != nil {
		ctx, cancel = context.WithCancel(ctx)
	}

	jsi := &jsSub{
		js:       js,
		stream:   stream,
		consumer: consumer,
		deliver:  deliver,
		hbi:      hbi,
		ordered:  o.ordered,
		ccreq:    ccreq,
		dseq:     1,
		pull:     isPullMode,
		nms:      nms,
		psubj:    subj,
		cancel:   cancel,
	}

	// Check if we are manual ack.
	if cb != nil && !o.mack {
		ocb := cb
		cb = func(m *Msg) { ocb(m); m.Ack() }
	}
	sub, err := nc.subscribe(deliver, queue, cb, ch, isSync, jsi)
	if err != nil {
		return nil, err
	}

	// If we fail and we had the sub we need to cleanup, but can't just do a straight Unsubscribe or Drain.
	// We need to clear the jsi so we do not remove any durables etc.
	cleanUpSub := func() {
		if sub != nil {
			sub.mu.Lock()
			sub.jsi = nil
			sub.mu.Unlock()
			sub.Unsubscribe()
		}
	}

	// If we are creating or updating let's process that request.
	if shouldCreate {
		j, err := json.Marshal(ccreq)
		if err != nil {
			cleanUpSub()
			return nil, err
		}

		var ccSubj string
		if isDurable {
			ccSubj = js.apiSubj(fmt.Sprintf(apiDurableCreateT, stream, cfg.Durable))
		} else {
			ccSubj = js.apiSubj(fmt.Sprintf(apiConsumerCreateT, stream))
		}

		if js.opts.shouldTrace {
			ctrace := js.opts.ctrace
			if ctrace.RequestSent != nil {
				ctrace.RequestSent(ccSubj, j)
			}
		}
		resp, err := nc.Request(ccSubj, j, js.opts.wait)
		if err != nil {
			cleanUpSub()
			if err == ErrNoResponders {
				err = ErrJetStreamNotEnabled
			}
			return nil, err
		}
		if js.opts.shouldTrace {
			ctrace := js.opts.ctrace
			if ctrace.ResponseReceived != nil {
				ctrace.ResponseReceived(ccSubj, resp.Data, resp.Header)
			}
		}

		var cinfo consumerResponse
		err = json.Unmarshal(resp.Data, &cinfo)
		if err != nil {
			cleanUpSub()
			return nil, err
		}
		info = cinfo.ConsumerInfo

		if cinfo.Error != nil {
			// We will not be using this sub here if we were push based.
			if !isPullMode {
				cleanUpSub()
			}
			if consumer != _EMPTY_ &&
				(strings.Contains(cinfo.Error.Description, `consumer already exists`) ||
					strings.Contains(cinfo.Error.Description, `consumer name already in use`)) {

				info, err = js.ConsumerInfo(stream, consumer)
				if err != nil {
					return nil, err
				}
				deliver, err = processConsInfo(info, o.cfg, isPullMode, subj, queue)
				if err != nil {
					return nil, err
				}

				if !isPullMode {
					// We can't reuse the channel, so if one was passed, we need to create a new one.
					if isSync {
						ch = make(chan *Msg, cap(ch))
					} else if ch != nil {
						// User provided (ChanSubscription), simply try to drain it.
						for done := false; !done; {
							select {
							case <-ch:
							default:
								done = true
							}
						}
					}
					jsi.deliver = deliver
					jsi.hbi = info.Config.Heartbeat

					// Recreate the subscription here.
					sub, err = nc.subscribe(jsi.deliver, queue, cb, ch, isSync, jsi)
					if err != nil {
						return nil, err
					}
					hasFC = info.Config.FlowControl
					hasHeartbeats = info.Config.Heartbeat > 0
				}
			} else {
				if cinfo.Error.Code == 404 {
					return nil, ErrStreamNotFound
				}
				return nil, fmt.Errorf("nats: %s", cinfo.Error.Description)
			}
		} else {
			// Since the library created the JS consumer, it will delete it on Unsubscribe()/Drain()
			sub.mu.Lock()
			sub.jsi.dc = true
			sub.jsi.pending = info.NumPending + info.Delivered.Consumer
			// If this is an ephemeral, we did not have a consumer name, we get it from the info
			// after the AddConsumer returns.
			if consumer == _EMPTY_ {
				sub.jsi.consumer = info.Name
			}
			sub.mu.Unlock()
		}
		// Capture max ack pending from the info response here which covers both
		// success and failure followed by consumer lookup.
		maxap = info.Config.MaxAckPending
	}

	// If maxap is greater than the default sub's pending limit, use that.
	if maxap > DefaultSubPendingMsgsLimit {
		// For bytes limit, use the min of maxp*1MB or DefaultSubPendingBytesLimit
		bl := maxap * 1024 * 1024
		if bl < DefaultSubPendingBytesLimit {
			bl = DefaultSubPendingBytesLimit
		}
		sub.SetPendingLimits(maxap, bl)
	}

	// Do heartbeats last if needed.
	if hasHeartbeats {
		sub.scheduleHeartbeatCheck()
	}
	// For ChanSubscriptions, if we know that there is flow control, we will
	// start a go routine that evaluates the number of delivered messages
	// and process flow control.
	if sub.Type() == ChanSubscription && hasFC {
		sub.chanSubcheckForFlowControlResponse()
	}

	// Wait for context to get canceled if there is one.
	if ctx != nil {
		go func() {
			<-ctx.Done()
			sub.Unsubscribe()
		}()
	}

	return sub, nil
}

// This long-lived routine is used per ChanSubscription to check
// on the number of delivered messages and check for flow control response.
func (sub *Subscription) chanSubcheckForFlowControlResponse() {
	sub.mu.Lock()
	// We don't use defer since if we need to send an RC reply, we need
	// to do it outside the sub's lock. So doing explicit unlock...
	if sub.closed {
		sub.mu.Unlock()
		return
	}
	var fcReply string
	var nc *Conn

	jsi := sub.jsi
	if jsi.csfct == nil {
		jsi.csfct = time.AfterFunc(chanSubFCCheckInterval, sub.chanSubcheckForFlowControlResponse)
	} else {
		fcReply = sub.checkForFlowControlResponse()
		nc = sub.conn
		// Do the reset here under the lock, it's ok...
		jsi.csfct.Reset(chanSubFCCheckInterval)
	}
	sub.mu.Unlock()
	// This call will return an error (which we don't care here)
	// if nc is nil or fcReply is empty.
	nc.Publish(fcReply, nil)
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

// isJSControlMessage will return true if this is an empty control status message
// and indicate what type of control message it is, say jsCtrlHB or jsCtrlFC
func isJSControlMessage(msg *Msg) (bool, int) {
	if len(msg.Data) > 0 || msg.Header.Get(statusHdr) != controlMsg {
		return false, 0
	}
	val := msg.Header.Get(descrHdr)
	if strings.HasPrefix(val, "Idle") {
		return true, jsCtrlHB
	}
	if strings.HasPrefix(val, "Flow") {
		return true, jsCtrlFC
	}
	return true, 0
}

// Keeps track of the incoming message's reply subject so that the consumer's
// state (deliver sequence, etc..) can be checked against heartbeats.
// We will also bump the incoming data message sequence that is used in FC cases.
// Runs under the subscription lock
func (sub *Subscription) trackSequences(reply string) {
	// For flow control, keep track of incoming message sequence.
	sub.jsi.fciseq++
	sub.jsi.cmeta = reply
}

// Check to make sure messages are arriving in order.
// Returns true if the sub had to be replaced. Will cause upper layers to return.
// The caller has verified that sub.jsi != nil and that this is not a control message.
// Lock should be held.
func (sub *Subscription) checkOrderedMsgs(m *Msg) bool {
	// Ignore msgs with no reply like HBs and flow control, they are handled elsewhere.
	if m.Reply == _EMPTY_ {
		return false
	}

	// Normal message here.
	tokens, err := getMetadataFields(m.Reply)
	if err != nil {
		return false
	}
	sseq, dseq := uint64(parseNum(tokens[ackStreamSeqTokenPos])), uint64(parseNum(tokens[ackConsumerSeqTokenPos]))

	jsi := sub.jsi
	if dseq != jsi.dseq {
		sub.resetOrderedConsumer(jsi.sseq + 1)
		return true
	}
	// Update our tracking here.
	jsi.dseq, jsi.sseq = dseq+1, sseq
	return false
}

// Update and replace sid.
// Lock should be held on entry but will be unlocked to prevent lock inversion.
func (sub *Subscription) applyNewSID() (osid int64) {
	nc := sub.conn
	sub.mu.Unlock()

	nc.subsMu.Lock()
	osid = sub.sid
	delete(nc.subs, osid)
	// Place new one.
	nc.ssid++
	nsid := nc.ssid
	nc.subs[nsid] = sub
	nc.subsMu.Unlock()

	sub.mu.Lock()
	sub.sid = nsid
	return osid
}

// We are here if we have detected a gap with an ordered consumer.
// We will create a new consumer and rewire the low level subscription.
// Lock should be held.
func (sub *Subscription) resetOrderedConsumer(sseq uint64) {
	nc := sub.conn
	if sub.jsi == nil || nc == nil || sub.closed {
		return
	}

	var maxStr string
	// If there was an AUTO_UNSUB done, we need to adjust the new value
	// to send after the SUB for the new sid.
	if sub.max > 0 {
		if sub.jsi.fciseq < sub.max {
			adjustedMax := sub.max - sub.jsi.fciseq
			maxStr = strconv.Itoa(int(adjustedMax))
		} else {
			// We are already at the max, so we should just unsub the
			// existing sub and be done
			go func(sid int64) {
				nc.mu.Lock()
				nc.bw.appendString(fmt.Sprintf(unsubProto, sid, _EMPTY_))
				nc.kickFlusher()
				nc.mu.Unlock()
			}(sub.sid)
			return
		}
	}

	// Quick unsubscribe. Since we know this is a simple push subscriber we do in place.
	osid := sub.applyNewSID()

	// Grab new inbox.
	newDeliver := nc.newInbox()
	sub.Subject = newDeliver

	// Snapshot the new sid under sub lock.
	nsid := sub.sid

	// We are still in the low level readLoop for the connection so we need
	// to spin a go routine to try to create the new consumer.
	go func() {
		// Unsubscribe and subscribe with new inbox and sid.
		// Remap a new low level sub into this sub since its client accessible.
		// This is done here in this go routine to prevent lock inversion.
		nc.mu.Lock()
		nc.bw.appendString(fmt.Sprintf(unsubProto, osid, _EMPTY_))
		nc.bw.appendString(fmt.Sprintf(subProto, newDeliver, _EMPTY_, nsid))
		if maxStr != _EMPTY_ {
			nc.bw.appendString(fmt.Sprintf(unsubProto, nsid, maxStr))
		}
		nc.kickFlusher()
		nc.mu.Unlock()

		pushErr := func(err error) {
			nc.handleConsumerSequenceMismatch(sub, err)
			nc.unsubscribe(sub, 0, true)
		}

		sub.mu.Lock()
		jsi := sub.jsi
		// Reset some items in jsi.
		jsi.dseq = 1
		jsi.cmeta = _EMPTY_
		jsi.fcr, jsi.fcd = _EMPTY_, 0
		jsi.deliver = newDeliver
		// Reset consumer request for starting policy.
		cfg := jsi.ccreq.Config
		cfg.DeliverSubject = newDeliver
		cfg.DeliverPolicy = DeliverByStartSequencePolicy
		cfg.OptStartSeq = sseq

		ccSubj := fmt.Sprintf(apiConsumerCreateT, jsi.stream)
		j, err := json.Marshal(jsi.ccreq)
		js := jsi.js
		sub.mu.Unlock()

		if err != nil {
			pushErr(err)
			return
		}

		resp, err := nc.Request(js.apiSubj(ccSubj), j, js.opts.wait)
		if err != nil {
			if err == ErrNoResponders {
				err = ErrJetStreamNotEnabled
			}
			pushErr(err)
			return
		}

		var cinfo consumerResponse
		err = json.Unmarshal(resp.Data, &cinfo)
		if err != nil {
			pushErr(err)
			return
		}

		if cinfo.Error != nil {
			pushErr(fmt.Errorf("nats: %s", cinfo.Error.Description))
			return
		}

		sub.mu.Lock()
		jsi.consumer = cinfo.Name
		sub.mu.Unlock()
	}()
}

// For jetstream subscriptions, returns the number of delivered messages.
// For ChanSubscription, this value is computed based on the known number
// of messages added to the channel minus the current size of that channel.
// Lock held on entry
func (sub *Subscription) getJSDelivered() uint64 {
	if sub.typ == ChanSubscription {
		return sub.jsi.fciseq - uint64(len(sub.mch))
	}
	return sub.delivered
}

// checkForFlowControlResponse will check to see if we should send a flow control response
// based on the subscription current delivered index and the target.
// Runs under subscription lock
func (sub *Subscription) checkForFlowControlResponse() string {
	// Caller has verified that there is a sub.jsi and fc
	jsi := sub.jsi
	jsi.active = true
	if sub.getJSDelivered() >= jsi.fcd {
		fcr := jsi.fcr
		jsi.fcr, jsi.fcd = _EMPTY_, 0
		return fcr
	}
	return _EMPTY_
}

// Record an inbound flow control message.
// Runs under subscription lock
func (sub *Subscription) scheduleFlowControlResponse(reply string) {
	sub.jsi.fcr, sub.jsi.fcd = reply, sub.jsi.fciseq
}

// Checks for activity from our consumer.
// If we do not think we are active send an async error.
func (sub *Subscription) activityCheck() {
	sub.mu.Lock()
	jsi := sub.jsi
	if jsi == nil || sub.closed {
		sub.mu.Unlock()
		return
	}

	active := jsi.active
	jsi.hbc.Reset(jsi.hbi * hbcThresh)
	jsi.active = false
	nc := sub.conn
	sub.mu.Unlock()

	if !active {
		nc.mu.Lock()
		if errCB := nc.Opts.AsyncErrorCB; errCB != nil {
			nc.ach.push(func() { errCB(nc, sub, ErrConsumerNotActive) })
		}
		nc.mu.Unlock()
	}
}

// scheduleHeartbeatCheck sets up the timer check to make sure we are active
// or receiving idle heartbeats..
func (sub *Subscription) scheduleHeartbeatCheck() {
	sub.mu.Lock()
	defer sub.mu.Unlock()

	jsi := sub.jsi
	if jsi == nil {
		return
	}

	if jsi.hbc == nil {
		jsi.hbc = time.AfterFunc(jsi.hbi*hbcThresh, sub.activityCheck)
	} else {
		jsi.hbc.Reset(jsi.hbi * hbcThresh)
	}
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

// checkForSequenceMismatch will make sure we have not missed any messages since last seen.
func (nc *Conn) checkForSequenceMismatch(msg *Msg, s *Subscription, jsi *jsSub) {
	// Process heartbeat received, get latest control metadata if present.
	s.mu.Lock()
	ctrl, ordered := jsi.cmeta, jsi.ordered
	jsi.active = true
	s.mu.Unlock()

	if ctrl == _EMPTY_ {
		return
	}

	tokens, err := getMetadataFields(ctrl)
	if err != nil {
		return
	}

	// Consumer sequence.
	var ldseq string
	dseq := tokens[ackConsumerSeqTokenPos]
	hdr := msg.Header[lastConsumerSeqHdr]
	if len(hdr) == 1 {
		ldseq = hdr[0]
	}

	// Detect consumer sequence mismatch and whether
	// should restart the consumer.
	if ldseq != dseq {
		// Dispatch async error including details such as
		// from where the consumer could be restarted.
		sseq := parseNum(tokens[ackStreamSeqTokenPos])
		if ordered {
			s.mu.Lock()
			s.resetOrderedConsumer(jsi.sseq + 1)
			s.mu.Unlock()
		} else {
			ecs := &ErrConsumerSequenceMismatch{
				StreamResumeSequence: uint64(sseq),
				ConsumerSequence:     uint64(parseNum(dseq)),
				LastConsumerSequence: uint64(parseNum(ldseq)),
			}
			nc.handleConsumerSequenceMismatch(s, ecs)
		}
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
	// For creating or updating.
	cfg *ConsumerConfig
	// For binding a subscription to a consumer without creating it.
	bound bool
	// For manual ack
	mack bool
	// For an ordered consumer.
	ordered bool
	ctx     context.Context
}

// OrderedConsumer will create a FIFO direct/ephemeral consumer for in order delivery of messages.
// There are no redeliveries and no acks, and flow control and heartbeats will be added but
// will be taken care of without additional client code.
func OrderedConsumer() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.ordered = true
		return nil
	})
}

// ManualAck disables auto ack functionality for async subscriptions.
func ManualAck() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.mack = true
		return nil
	})
}

// Description will set the description for the created consumer.
func Description(description string) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.Description = description
		return nil
	})
}

// Check that the durable name is valid, that is, that it does not contain
// any ".", and if it does return ErrInvalidDurableName, otherwise nil.
func checkDurName(dur string) error {
	if strings.Contains(dur, ".") {
		return ErrInvalidDurableName
	}
	return nil
}

// Durable defines the consumer name for JetStream durable subscribers.
// This function will return ErrInvalidDurableName in the name contains
// any dot ".".
func Durable(consumer string) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		if opts.cfg.Durable != _EMPTY_ {
			return fmt.Errorf("nats: option Durable set more than once")
		}
		if opts.consumer != _EMPTY_ && opts.consumer != consumer {
			return fmt.Errorf("nats: duplicate consumer names (%s and %s)", opts.consumer, consumer)
		}
		if err := checkDurName(consumer); err != nil {
			return err
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

// DeliverLastPerSubject configures a Consumer to receive messages
// starting with the latest one for each filtered subject.
func DeliverLastPerSubject() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.DeliverPolicy = DeliverLastPerSubjectPolicy
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

// ReplayInstant replays the messages as fast as possible.
func ReplayInstant() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.ReplayPolicy = ReplayInstantPolicy
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

// BackOff is an array of time durations that represent the time to delay based on delivery count.
func BackOff(backOff []time.Duration) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.BackOff = backOff
		return nil
	})
}

// BindStream binds a consumer to a stream explicitly based on a name.
// When a stream name is not specified, the library uses the subscribe
// subject as a way to find the stream name. It is done by making a request
// to the server to get list of stream names that have a filter for this
// subject. If the returned list contains a single stream, then this
// stream name will be used, otherwise the `ErrNoMatchingStream` is returned.
// To avoid the stream lookup, provide the stream name with this function.
// See also `Bind()`.
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

// DeliverSubject specifies the JetStream consumer deliver subject.
//
// This option is used only in situations where the consumer does not exist
// and a creation request is sent to the server. If not provided, an inbox
// will be selected.
// If a consumer exists, then the NATS subscription will be created on
// the JetStream consumer's DeliverSubject, not necessarily this subject.
func DeliverSubject(subject string) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.DeliverSubject = subject
		return nil
	})
}

// HeadersOnly() will instruct the consumer to only deliver headers and no payloads.
func HeadersOnly() SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.HeadersOnly = true
		return nil
	})
}

// MaxRequestBatch sets the maximum pull consumer batch size that a Fetch()
// can request.
func MaxRequestBatch(max int) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.MaxRequestBatch = max
		return nil
	})
}

// MaxRequestExpires sets the maximum pull consumer request expiration that a
// Fetch() can request (using the Fetch's timeout value).
func MaxRequestExpires(max time.Duration) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		opts.cfg.MaxRequestExpires = max
		return nil
	})
}

// InactiveThreshold indicates how long the server should keep an ephemeral
// after detecting loss of interest.
func InactiveThreshold(threshold time.Duration) SubOpt {
	return subOptFn(func(opts *subOpts) error {
		if threshold < 0 {
			return fmt.Errorf("invalid InactiveThreshold value (%v), needs to be greater or equal to 0", threshold)
		}
		opts.cfg.InactiveThreshold = threshold
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

var (
	// errNoMessages is an error that a Fetch request using no_wait can receive to signal
	// that there are no more messages available.
	errNoMessages = errors.New("nats: no messages")

	// errRequestsPending is an error that represents a sub.Fetch requests that was using
	// no_wait and expires time got discarded by the server.
	errRequestsPending = errors.New("nats: requests pending")
)

// Returns if the given message is a user message or not, and if
// `checkSts` is true, returns appropriate error based on the
// content of the status (404, etc..)
func checkMsg(msg *Msg, checkSts, isNoWait bool) (usrMsg bool, err error) {
	// Assume user message
	usrMsg = true

	// If payload or no header, consider this a user message
	if len(msg.Data) > 0 || len(msg.Header) == 0 {
		return
	}
	// Look for status header
	val := msg.Header.Get(statusHdr)
	// If not present, then this is considered a user message
	if val == _EMPTY_ {
		return
	}
	// At this point, this is not a user message since there is
	// no payload and a "Status" header.
	usrMsg = false

	// If we don't care about status, we are done.
	if !checkSts {
		return
	}
	switch val {
	case noResponders:
		err = ErrNoResponders
	case noMessagesSts:
		// 404 indicates that there are no messages.
		err = errNoMessages
	case reqTimeoutSts:
		// In case of a fetch request with no wait request and expires time,
		// need to skip 408 errors and retry.
		if isNoWait {
			err = errRequestsPending
		} else {
			// Older servers may send a 408 when a request in the server was expired
			// and interest is still found, which will be the case for our
			// implementation. Regardless, ignore 408 errors until receiving at least
			// one message when making requests without no_wait.
			err = ErrTimeout
		}
	default:
		err = fmt.Errorf("nats: %s", msg.Header.Get(descrHdr))
	}
	return
}

// Fetch pulls a batch of messages from a stream for a pull consumer.
func (sub *Subscription) Fetch(batch int, opts ...PullOpt) ([]*Msg, error) {
	if sub == nil {
		return nil, ErrBadSubscription
	}
	if batch < 1 {
		return nil, ErrInvalidArg
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
	// Reject if this is not a pull subscription. Note that sub.typ is SyncSubscription,
	// so check for jsi.pull boolean instead.
	if jsi == nil || !jsi.pull {
		sub.mu.Unlock()
		return nil, ErrTypeSubscription
	}

	nc := sub.conn
	nms := sub.jsi.nms
	rply := sub.jsi.deliver
	js := sub.jsi.js
	pmc := len(sub.mch) > 0

	// All fetch requests have an expiration, in case of no explicit expiration
	// then the default timeout of the JetStream context is used.
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
	if ctx == nil {
		ctx, cancel = context.WithTimeout(context.Background(), ttl)
		defer cancel()
	} else if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		// Prevent from passing the background context which will just block
		// and cannot be canceled either.
		if octx, ok := ctx.(ContextOpt); ok && octx.Context == context.Background() {
			return nil, ErrNoDeadlineContext
		}

		// If the context did not have a deadline, then create a new child context
		// that will use the default timeout from the JS context.
		ctx, cancel = context.WithTimeout(ctx, ttl)
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

	// Use the deadline of the context to base the expire times.
	deadline, _ := ctx.Deadline()
	ttl = time.Until(deadline)
	checkCtxErr := func(err error) error {
		if o.ctx == nil && err == context.DeadlineExceeded {
			return ErrTimeout
		}
		return err
	}

	var (
		msgs = make([]*Msg, 0, batch)
		msg  *Msg
	)
	for pmc && len(msgs) < batch {
		// Check next msg with booleans that say that this is an internal call
		// for a pull subscribe (so don't reject it) and don't wait if there
		// are no messages.
		msg, err = sub.nextMsgWithContext(ctx, true, false)
		if err != nil {
			if err == errNoMessages {
				err = nil
			}
			break
		}
		// Check msg but just to determine if this is a user message
		// or status message, however, we don't care about values of status
		// messages at this point in the Fetch() call, so checkMsg can't
		// return an error.
		if usrMsg, _ := checkMsg(msg, false, false); usrMsg {
			msgs = append(msgs, msg)
		}
	}
	if err == nil && len(msgs) < batch {
		// For batch real size of 1, it does not make sense to set no_wait in
		// the request.
		noWait := batch-len(msgs) > 1
		var nr nextRequest

		sendReq := func() error {
			// The current deadline for the context will be used
			// to set the expires TTL for a fetch request.
			deadline, _ = ctx.Deadline()
			ttl = time.Until(deadline)

			// Check if context has already been canceled or expired.
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Make our request expiration a bit shorter than the current timeout.
			expires := ttl
			if ttl >= 20*time.Millisecond {
				expires = ttl - 10*time.Millisecond
			}

			nr.Batch = batch - len(msgs)
			nr.Expires = expires
			nr.NoWait = noWait
			req, _ := json.Marshal(nr)
			return nc.PublishRequest(nms, rply, req)
		}

		err = sendReq()
		for err == nil && len(msgs) < batch {
			// Ask for next message and wait if there are no messages
			msg, err = sub.nextMsgWithContext(ctx, true, true)
			if err == nil {
				var usrMsg bool

				usrMsg, err = checkMsg(msg, true, noWait)
				if err == nil && usrMsg {
					msgs = append(msgs, msg)
				} else if noWait && (err == errNoMessages || err == errRequestsPending) && len(msgs) == 0 {
					// If we have a 404/408 for our "no_wait" request and have
					// not collected any message, then resend request to
					// wait this time.
					noWait = false
					err = sendReq()
				} else if err == ErrTimeout && len(msgs) == 0 {
					// If we get a 408, we will bail if we already collected some
					// messages, otherwise ignore and go back calling NextMsg.
					err = nil
				}
			}
		}
	}
	// If there is at least a message added to msgs, then need to return OK and no error
	if err != nil && len(msgs) == 0 {
		return nil, checkCtxErr(err)
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
	resp, err := js.apiRequestWithContext(ctx, js.apiSubj(ccInfoSubj), nil)
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
		if info.Error.Code == 404 {
			return nil, ErrConsumerNotFound
		}
		return nil, fmt.Errorf("nats: %s", info.Error.Description)
	}
	return info.ConsumerInfo, nil
}

// a RequestWithContext with tracing via TraceCB
func (js *js) apiRequestWithContext(ctx context.Context, subj string, data []byte) (*Msg, error) {
	if js.opts.shouldTrace {
		ctrace := js.opts.ctrace
		if ctrace.RequestSent != nil {
			ctrace.RequestSent(subj, data)
		}
	}
	resp, err := js.nc.RequestWithContext(ctx, subj, data)
	if err != nil {
		return nil, err
	}
	if js.opts.shouldTrace {
		ctrace := js.opts.ctrace
		if ctrace.RequestSent != nil {
			ctrace.ResponseReceived(subj, resp.Data, resp.Header)
		}
	}

	return resp, nil
}

func (m *Msg) checkReply() (*js, *jsSub, error) {
	if m == nil || m.Sub == nil {
		return nil, nil, ErrMsgNotBound
	}
	if m.Reply == _EMPTY_ {
		return nil, nil, ErrMsgNoReply
	}
	sub := m.Sub
	if sub.jsi == nil {
		// Not using a JS context.
		return nil, nil, nil
	}
	sub.mu.Lock()
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
		return ErrMsgAlreadyAckd
	}

	m.Sub.mu.Lock()
	nc := m.Sub.conn
	m.Sub.mu.Unlock()

	usesCtx := o.ctx != nil
	usesWait := o.ttl > 0

	// Only allow either AckWait or Context option to set the timeout.
	if usesWait && usesCtx {
		return ErrContextAndTimeout
	}

	sync = sync || usesCtx || usesWait
	ctx := o.ctx
	wait := defaultRequestWait
	if usesWait {
		wait = o.ttl
	} else if js != nil {
		wait = js.opts.wait
	}

	var body []byte
	// This will be > 0 only when called from NakWithDelay()
	if o.nakDelay > 0 {
		body = []byte(fmt.Sprintf("%s {\"delay\": %d}", ackType, o.nakDelay.Nanoseconds()))
	} else {
		body = ackType
	}

	if sync {
		if usesCtx {
			_, err = nc.RequestWithContext(ctx, m.Reply, body)
		} else {
			_, err = nc.Request(m.Reply, body, wait)
		}
	} else {
		err = nc.Publish(m.Reply, body)
	}

	// Mark that the message has been acked unless it is ackProgress
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

// Nak negatively acknowledges a message. This tells the server to redeliver
// the message after the give `delay` duration. You can configure the number
// of redeliveries by passing nats.MaxDeliver when you Subscribe.
// The default is infinite redeliveries.
func (m *Msg) NakWithDelay(delay time.Duration, opts ...AckOpt) error {
	if delay > 0 {
		opts = append(opts, nakDelay(delay))
	}
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
	Domain       string
}

const (
	ackDomainTokenPos       = 2
	ackAccHashTokenPos      = 3
	ackStreamTokenPos       = 4
	ackConsumerTokenPos     = 5
	ackNumDeliveredTokenPos = 6
	ackStreamSeqTokenPos    = 7
	ackConsumerSeqTokenPos  = 8
	ackTimestampSeqTokenPos = 9
	ackNumPendingTokenPos   = 10
)

func getMetadataFields(subject string) ([]string, error) {
	const v1TokenCounts = 9
	const v2TokenCounts = 12
	const noDomainName = "_"

	const btsep = '.'
	tsa := [v2TokenCounts]string{}
	start, tokens := 0, tsa[:0]
	for i := 0; i < len(subject); i++ {
		if subject[i] == btsep {
			tokens = append(tokens, subject[start:i])
			start = i + 1
		}
	}
	tokens = append(tokens, subject[start:])
	//
	// Newer server will include the domain name and account hash in the subject,
	// and a token at the end.
	//
	// Old subject was:
	// $JS.ACK.<stream>.<consumer>.<delivered>.<sseq>.<cseq>.<tm>.<pending>
	//
	// New subject would be:
	// $JS.ACK.<domain>.<account hash>.<stream>.<consumer>.<delivered>.<sseq>.<cseq>.<tm>.<pending>.<a token with a random value>
	//
	// v1 has 9 tokens, v2 has 12, but we must not be strict on the 12th since
	// it may be removed in the future. Also, the library has no use for it.
	// The point is that a v2 ACK subject is valid if it has at least 11 tokens.
	//
	l := len(tokens)
	// If lower than 9 or more than 9 but less than 11, report an error
	if l < v1TokenCounts || (l > v1TokenCounts && l < v2TokenCounts-1) {
		return nil, ErrNotJSMessage
	}
	if tokens[0] != "$JS" || tokens[1] != "ACK" {
		return nil, ErrNotJSMessage
	}
	// For v1 style, we insert 2 empty tokens (domain and hash) so that the
	// rest of the library references known fields at a constant location.
	if l == 9 {
		// Extend the array (we know the backend is big enough)
		tokens = append(tokens, _EMPTY_, _EMPTY_)
		// Move to the right anything that is after "ACK" token.
		copy(tokens[ackDomainTokenPos+2:], tokens[ackDomainTokenPos:])
		// Clear the domain and hash tokens
		tokens[ackDomainTokenPos], tokens[ackAccHashTokenPos] = _EMPTY_, _EMPTY_

	} else if tokens[ackDomainTokenPos] == noDomainName {
		// If domain is "_", replace with empty value.
		tokens[ackDomainTokenPos] = _EMPTY_
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
		Domain:       tokens[ackDomainTokenPos],
		NumDelivered: uint64(parseNum(tokens[ackNumDeliveredTokenPos])),
		NumPending:   uint64(parseNum(tokens[ackNumPendingTokenPos])),
		Timestamp:    time.Unix(0, parseNum(tokens[ackTimestampSeqTokenPos])),
		Stream:       tokens[ackStreamTokenPos],
		Consumer:     tokens[ackConsumerTokenPos],
	}
	meta.Sequence.Stream = uint64(parseNum(tokens[ackStreamSeqTokenPos]))
	meta.Sequence.Consumer = uint64(parseNum(tokens[ackConsumerSeqTokenPos]))
	return meta, nil
}

// Quick parser for positive numbers in ack reply encoding.
func parseNum(d string) (n int64) {
	if len(d) == 0 {
		return -1
	}

	// ASCII numbers 0-9
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

	// For configuration mismatch check
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

	// For configuration mismatch check
	replayPolicyNotSet = 99
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

	// DeliverLastPerSubjectPolicy will start the consumer with the last message
	// for all subjects received.
	DeliverLastPerSubjectPolicy

	// For configuration mismatch check
	deliverPolicyNotSet = 99
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
	case jsonString("last_per_subject"):
		*p = DeliverLastPerSubjectPolicy
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
	case DeliverLastPerSubjectPolicy:
		return json.Marshal("last_per_subject")
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
