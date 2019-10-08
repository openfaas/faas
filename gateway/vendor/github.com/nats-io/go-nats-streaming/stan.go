// Copyright 2016-2018 The NATS Authors
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

// Package stan is a Go client for the NATS Streaming messaging system (https://nats.io).
package stan

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming/pb"
	"github.com/nats-io/nuid"
)

// Version is the NATS Streaming Go Client version
const Version = "0.4.0"

const (
	// DefaultNatsURL is the default URL the client connects to
	DefaultNatsURL = "nats://localhost:4222"
	// DefaultConnectWait is the default timeout used for the connect operation
	DefaultConnectWait = 2 * time.Second
	// DefaultDiscoverPrefix is the prefix subject used to connect to the NATS Streaming server
	DefaultDiscoverPrefix = "_STAN.discover"
	// DefaultACKPrefix is the prefix subject used to send ACKs to the NATS Streaming server
	DefaultACKPrefix = "_STAN.acks"
	// DefaultMaxPubAcksInflight is the default maximum number of published messages
	// without outstanding ACKs from the server
	DefaultMaxPubAcksInflight = 16384
	// DefaultPingInterval is the default interval (in seconds) at which a connection sends a PING to the server
	DefaultPingInterval = 5
	// DefaultPingMaxOut is the number of PINGs without a response before the connection is considered lost.
	DefaultPingMaxOut = 3
)

// Conn represents a connection to the NATS Streaming subsystem. It can Publish and
// Subscribe to messages within the NATS Streaming cluster.
type Conn interface {
	// Publish
	Publish(subject string, data []byte) error
	PublishAsync(subject string, data []byte, ah AckHandler) (string, error)

	// Subscribe
	Subscribe(subject string, cb MsgHandler, opts ...SubscriptionOption) (Subscription, error)

	// QueueSubscribe
	QueueSubscribe(subject, qgroup string, cb MsgHandler, opts ...SubscriptionOption) (Subscription, error)

	// Close
	Close() error

	// NatsConn returns the underlying NATS conn. Use this with care. For
	// example, closing the wrapped NATS conn will put the NATS Streaming Conn
	// in an invalid state.
	NatsConn() *nats.Conn
}

const (
	// Client send connID in ConnectRequest and PubMsg, and server
	// listens and responds to client PINGs. The validity of the
	// connection (based on connID) is checked on incoming PINGs.
	protocolOne = int32(1)
)

// Errors
var (
	ErrConnectReqTimeout = errors.New("stan: connect request timeout")
	ErrCloseReqTimeout   = errors.New("stan: close request timeout")
	ErrSubReqTimeout     = errors.New("stan: subscribe request timeout")
	ErrUnsubReqTimeout   = errors.New("stan: unsubscribe request timeout")
	ErrConnectionClosed  = errors.New("stan: connection closed")
	ErrTimeout           = errors.New("stan: publish ack timeout")
	ErrBadAck            = errors.New("stan: malformed ack")
	ErrBadSubscription   = errors.New("stan: invalid subscription")
	ErrBadConnection     = errors.New("stan: invalid connection")
	ErrManualAck         = errors.New("stan: cannot manually ack in auto-ack mode")
	ErrNilMsg            = errors.New("stan: nil message")
	ErrNoServerSupport   = errors.New("stan: not supported by server")
	ErrMaxPings          = errors.New("stan: connection lost due to PING failure")
)

var testAllowMillisecInPings = false

// AckHandler is used for Async Publishing to provide status of the ack.
// The func will be passed the GUID and any error state. No error means the
// message was successfully received by NATS Streaming.
type AckHandler func(string, error)

// ConnectionLostHandler is used to be notified if the Streaming connection
// is closed due to unexpected errors.
type ConnectionLostHandler func(Conn, error)

// Options can be used to a create a customized connection.
type Options struct {
	NatsURL            string
	NatsConn           *nats.Conn
	ConnectTimeout     time.Duration
	AckTimeout         time.Duration
	DiscoverPrefix     string
	MaxPubAcksInflight int
	PingIterval        int // In seconds
	PingMaxOut         int
	ConnectionLostCB   ConnectionLostHandler
}

// DefaultOptions are the NATS Streaming client's default options
var DefaultOptions = Options{
	NatsURL:            DefaultNatsURL,
	ConnectTimeout:     DefaultConnectWait,
	AckTimeout:         DefaultAckWait,
	DiscoverPrefix:     DefaultDiscoverPrefix,
	MaxPubAcksInflight: DefaultMaxPubAcksInflight,
	PingIterval:        DefaultPingInterval,
	PingMaxOut:         DefaultPingMaxOut,
}

// Option is a function on the options for a connection.
type Option func(*Options) error

// NatsURL is an Option to set the URL the client should connect to.
func NatsURL(u string) Option {
	return func(o *Options) error {
		o.NatsURL = u
		return nil
	}
}

// ConnectWait is an Option to set the timeout for establishing a connection.
func ConnectWait(t time.Duration) Option {
	return func(o *Options) error {
		o.ConnectTimeout = t
		return nil
	}
}

// PubAckWait is an Option to set the timeout for waiting for an ACK for a
// published message.
func PubAckWait(t time.Duration) Option {
	return func(o *Options) error {
		o.AckTimeout = t
		return nil
	}
}

// MaxPubAcksInflight is an Option to set the maximum number of published
// messages without outstanding ACKs from the server.
func MaxPubAcksInflight(max int) Option {
	return func(o *Options) error {
		o.MaxPubAcksInflight = max
		return nil
	}
}

// NatsConn is an Option to set the underlying NATS connection to be used
// by a NATS Streaming Conn object.
func NatsConn(nc *nats.Conn) Option {
	return func(o *Options) error {
		o.NatsConn = nc
		return nil
	}
}

// Pings is an Option to set the ping interval and max out values.
// The interval needs to be at least 1 and represents the number
// of seconds.
// The maxOut needs to be at least 2, since the count of sent PINGs
// increase whenever a PING is sent and reset to 0 when a response
// is received. Setting to 1 would cause the library to close the
// connection right away.
func Pings(interval, maxOut int) Option {
	return func(o *Options) error {
		// For tests, we may pass negative value that will be interpreted
		// by the library as milliseconds. If this test boolean is set,
		// do not check values.
		if !testAllowMillisecInPings {
			if interval < 1 || maxOut <= 2 {
				return fmt.Errorf("Invalid ping values: interval=%v (min>0) maxOut=%v (min=2)", interval, maxOut)
			}
		}
		o.PingIterval = interval
		o.PingMaxOut = maxOut
		return nil
	}
}

// SetConnectionLostHandler is an Option to set the connection lost handler.
// This callback will be invoked should the client permanently lose
// contact with the server (or another client replaces it while being
// disconnected). The callback will not be invoked on normal Conn.Close().
func SetConnectionLostHandler(handler ConnectionLostHandler) Option {
	return func(o *Options) error {
		o.ConnectionLostCB = handler
		return nil
	}
}

// A conn represents a bare connection to a stan cluster.
type conn struct {
	sync.RWMutex
	clientID         string
	connID           []byte // This is a NUID that uniquely identify connections.
	pubPrefix        string // Publish prefix set by stan, append our subject.
	subRequests      string // Subject to send subscription requests.
	unsubRequests    string // Subject to send unsubscribe requests.
	subCloseRequests string // Subject to send subscription close requests.
	closeRequests    string // Subject to send close requests.
	ackSubject       string // publish acks
	ackSubscription  *nats.Subscription
	hbSubscription   *nats.Subscription
	subMap           map[string]*subscription
	pubAckMap        map[string]*ack
	pubAckChan       chan (struct{})
	opts             Options
	nc               *nats.Conn
	ncOwned          bool       // NATS Streaming created the connection, so needs to close it.
	pubNUID          *nuid.NUID // NUID generator for published messages.
	connLostCB       ConnectionLostHandler

	pingMu       sync.Mutex
	pingSub      *nats.Subscription
	pingTimer    *time.Timer
	pingBytes    []byte
	pingRequests string
	pingInbox    string
	pingInterval time.Duration
	pingMaxOut   int
	pingOut      int
}

// Closure for ack contexts.
type ack struct {
	t  *time.Timer
	ah AckHandler
	ch chan error
}

// Connect will form a connection to the NATS Streaming subsystem.
// Note that clientID can contain only alphanumeric and `-` or `_` characters.
func Connect(stanClusterID, clientID string, options ...Option) (Conn, error) {
	// Process Options
	c := conn{clientID: clientID, opts: DefaultOptions, connID: []byte(nuid.Next()), pubNUID: nuid.New()}
	for _, opt := range options {
		if err := opt(&c.opts); err != nil {
			return nil, err
		}
	}
	// Check if the user has provided a connection as an option
	c.nc = c.opts.NatsConn
	// Create a NATS connection if it doesn't exist.
	if c.nc == nil {
		// We will set the max reconnect attempts to -1 (infinite)
		// and the reconnect buffer to -1 to prevent any buffering
		// (which may cause a published message to be flushed on
		// reconnect while the API may have returned an error due
		// to PubAck timeout.
		nc, err := nats.Connect(c.opts.NatsURL,
			nats.Name(clientID),
			nats.MaxReconnects(-1),
			nats.ReconnectBufSize(-1))
		if err != nil {
			return nil, err
		}
		c.nc = nc
		c.ncOwned = true
	} else if !c.nc.IsConnected() {
		// Bail if the custom NATS connection is disconnected
		return nil, ErrBadConnection
	}

	// Create a heartbeat inbox
	hbInbox := nats.NewInbox()
	var err error
	if c.hbSubscription, err = c.nc.Subscribe(hbInbox, c.processHeartBeat); err != nil {
		c.Close()
		return nil, err
	}

	// Prepare a subscription on ping responses, even if we are not
	// going to need it, so that if that fails, it fails before initiating
	// a connection.
	pingSub, err := c.nc.Subscribe(nats.NewInbox(), c.processPingResponse)
	if err != nil {
		c.Close()
		return nil, err
	}

	// Send Request to discover the cluster
	discoverSubject := c.opts.DiscoverPrefix + "." + stanClusterID
	req := &pb.ConnectRequest{
		ClientID:       clientID,
		HeartbeatInbox: hbInbox,
		ConnID:         c.connID,
		Protocol:       protocolOne,
		PingInterval:   int32(c.opts.PingIterval),
		PingMaxOut:     int32(c.opts.PingMaxOut),
	}
	b, _ := req.Marshal()
	reply, err := c.nc.Request(discoverSubject, b, c.opts.ConnectTimeout)
	if err != nil {
		c.Close()
		if err == nats.ErrTimeout {
			return nil, ErrConnectReqTimeout
		}
		return nil, err
	}
	// Process the response, grab server pubPrefix
	cr := &pb.ConnectResponse{}
	err = cr.Unmarshal(reply.Data)
	if err != nil {
		c.Close()
		return nil, err
	}
	if cr.Error != "" {
		c.Close()
		return nil, errors.New(cr.Error)
	}

	// Capture cluster configuration endpoints to publish and subscribe/unsubscribe.
	c.pubPrefix = cr.PubPrefix
	c.subRequests = cr.SubRequests
	c.unsubRequests = cr.UnsubRequests
	c.subCloseRequests = cr.SubCloseRequests
	c.closeRequests = cr.CloseRequests

	// Setup the ACK subscription
	c.ackSubject = DefaultACKPrefix + "." + nuid.Next()
	if c.ackSubscription, err = c.nc.Subscribe(c.ackSubject, c.processAck); err != nil {
		c.Close()
		return nil, err
	}
	c.ackSubscription.SetPendingLimits(1024*1024, 32*1024*1024)
	c.pubAckMap = make(map[string]*ack)

	// Create Subscription map
	c.subMap = make(map[string]*subscription)

	c.pubAckChan = make(chan struct{}, c.opts.MaxPubAcksInflight)

	// Capture the connection error cb
	c.connLostCB = c.opts.ConnectionLostCB

	unsubPingSub := true
	// Do this with servers which are at least at protcolOne.
	if cr.Protocol >= protocolOne {
		// Note that in the future server may override client ping
		// interval value sent in ConnectRequest, so use the
		// value in ConnectResponse to decide if we send PINGs
		// and at what interval.
		// In tests, the interval could be negative to indicate
		// milliseconds.
		if cr.PingInterval != 0 {
			unsubPingSub = false

			// These will be immutable.
			c.pingRequests = cr.PingRequests
			c.pingInbox = pingSub.Subject
			// In test, it is possible that we get a negative value
			// to represent milliseconds.
			if testAllowMillisecInPings && cr.PingInterval < 0 {
				c.pingInterval = time.Duration(cr.PingInterval*-1) * time.Millisecond
			} else {
				// PingInterval is otherwise assumed to be in seconds.
				c.pingInterval = time.Duration(cr.PingInterval) * time.Second
			}
			c.pingMaxOut = int(cr.PingMaxOut)
			c.pingBytes, _ = (&pb.Ping{ConnID: c.connID}).Marshal()
			c.pingSub = pingSub
			// Set the timer now that we are set. Use lock to create
			// synchronization point.
			c.pingMu.Lock()
			c.pingTimer = time.AfterFunc(c.pingInterval, c.pingServer)
			c.pingMu.Unlock()
		}
	}
	if unsubPingSub {
		pingSub.Unsubscribe()
	}

	// Attach a finalizer
	runtime.SetFinalizer(&c, func(sc *conn) { sc.Close() })

	return &c, nil
}

// Sends a PING (containing the connection's ID) to the server at intervals
// specified by PingInterval option when connection is created.
// Everytime a PING is sent, the number of outstanding PINGs is increased.
// If the total number is > than the PingMaxOut option, then the connection
// is closed, and connection error callback invoked if one was specified.
func (sc *conn) pingServer() {
	sc.pingMu.Lock()
	// In case the timer fired while we were stopping it.
	if sc.pingTimer == nil {
		sc.pingMu.Unlock()
		return
	}
	sc.pingOut++
	if sc.pingOut > sc.pingMaxOut {
		sc.pingMu.Unlock()
		sc.closeDueToPing(ErrMaxPings)
		return
	}
	sc.pingTimer.Reset(sc.pingInterval)
	nc := sc.nc
	sc.pingMu.Unlock()
	// Send the PING now. If the NATS connection is reported closed,
	// we are done.
	if err := nc.PublishRequest(sc.pingRequests, sc.pingInbox, sc.pingBytes); err == nats.ErrConnectionClosed {
		sc.closeDueToPing(err)
	}
}

// Receives PING responses from the server.
// If the response contains an error message, the connection is closed
// and the connection error callback is invoked (if one is specified).
// If no error, the number of ping out is reset to 0. There is no
// decrement by one since for a given PING, the client may received
// many responses when servers are running in channel partitioning mode.
// Regardless, any positive response from the server ought to signal
// that the connection is ok.
func (sc *conn) processPingResponse(m *nats.Msg) {
	// No data means OK (we don't have to call Unmarshal)
	if len(m.Data) > 0 {
		pingResp := &pb.PingResponse{}
		if err := pingResp.Unmarshal(m.Data); err != nil {
			return
		}
		if pingResp.Error != "" {
			sc.closeDueToPing(errors.New(pingResp.Error))
			return
		}
	}
	// Do not attempt to decrement, simply reset to 0.
	sc.pingMu.Lock()
	sc.pingOut = 0
	sc.pingMu.Unlock()
}

// Closes a connection and invoke the connection error callback if one
// was registered when the connection was created.
func (sc *conn) closeDueToPing(err error) {
	sc.Lock()
	if sc.nc == nil {
		sc.Unlock()
		return
	}
	// Stop timer, unsubscribe, fail the pubs, etc..
	sc.cleanupOnClose(err)
	// No need to send Close prototol, so simply close the underlying
	// NATS connection (if we own it, and if not already closed)
	if sc.ncOwned && !sc.nc.IsClosed() {
		sc.nc.Close()
	}
	// Mark this streaming connection as closed. Do this under pingMu lock.
	sc.pingMu.Lock()
	sc.nc = nil
	sc.pingMu.Unlock()
	// Capture callback (even though this is immutable).
	cb := sc.connLostCB
	sc.Unlock()
	if cb != nil {
		// Execute in separate go routine.
		go cb(sc, err)
	}
}

// Do some cleanup when connection is lost or closed.
// Connection lock is held on entry, and sc.nc is guaranteed not to be nil.
func (sc *conn) cleanupOnClose(err error) {
	sc.pingMu.Lock()
	if sc.pingTimer != nil {
		sc.pingTimer.Stop()
		sc.pingTimer = nil
	}
	sc.pingMu.Unlock()

	// Unsubscribe only if the NATS connection is not already closed...
	if !sc.nc.IsClosed() {
		if sc.ackSubscription != nil {
			sc.ackSubscription.Unsubscribe()
		}
		if sc.pingSub != nil {
			sc.pingSub.Unsubscribe()
		}
	}
	// Fail all pending pubs
	for guid, pubAck := range sc.pubAckMap {
		if pubAck.t != nil {
			pubAck.t.Stop()
		}
		if pubAck.ah != nil {
			pubAck.ah(guid, err)
		} else if pubAck.ch != nil {
			pubAck.ch <- err
		}
		delete(sc.pubAckMap, guid)
		if len(sc.pubAckChan) > 0 {
			<-sc.pubAckChan
		}
	}
}

// Close a connection to the stan system.
func (sc *conn) Close() error {
	sc.Lock()
	defer sc.Unlock()

	if sc.nc == nil {
		// We are already closed.
		return nil
	}

	// Capture for NATS calls below.
	nc := sc.nc
	if sc.ncOwned {
		defer nc.Close()
	}

	// Now close ourselves.
	sc.cleanupOnClose(ErrConnectionClosed)

	// Signals we are closed.
	// Do this also under pingMu lock so that we don't need
	// to grab sc's lock in pingServer.
	sc.pingMu.Lock()
	sc.nc = nil
	sc.pingMu.Unlock()

	req := &pb.CloseRequest{ClientID: sc.clientID}
	b, _ := req.Marshal()
	reply, err := nc.Request(sc.closeRequests, b, sc.opts.ConnectTimeout)
	if err != nil {
		if err == nats.ErrTimeout {
			return ErrCloseReqTimeout
		}
		return err
	}
	cr := &pb.CloseResponse{}
	err = cr.Unmarshal(reply.Data)
	if err != nil {
		return err
	}
	if cr.Error != "" {
		return errors.New(cr.Error)
	}
	return nil
}

// NatsConn returns the underlying NATS conn. Use this with care. For example,
// closing the wrapped NATS conn will put the NATS Streaming Conn in an invalid
// state.
func (sc *conn) NatsConn() *nats.Conn {
	sc.RLock()
	nc := sc.nc
	sc.RUnlock()
	return nc
}

// Process a heartbeat from the NATS Streaming cluster
func (sc *conn) processHeartBeat(m *nats.Msg) {
	// No payload assumed, just reply.
	sc.RLock()
	nc := sc.nc
	sc.RUnlock()
	if nc != nil {
		nc.Publish(m.Reply, nil)
	}
}

// Process an ack from the NATS Streaming cluster
func (sc *conn) processAck(m *nats.Msg) {
	pa := &pb.PubAck{}
	err := pa.Unmarshal(m.Data)
	if err != nil {
		panic(fmt.Errorf("Error during ack unmarshal: %v", err))
	}

	// Remove
	a := sc.removeAck(pa.Guid)
	if a != nil {
		// Capture error if it exists.
		if pa.Error != "" {
			err = errors.New(pa.Error)
		}
		if a.ah != nil {
			// Perform the ackHandler callback
			a.ah(pa.Guid, err)
		} else if a.ch != nil {
			// Send to channel directly
			a.ch <- err
		}
	}
}

// Publish will publish to the cluster and wait for an ACK.
func (sc *conn) Publish(subject string, data []byte) error {
	// Need to make this a buffered channel of 1 in case
	// a publish call is blocked in pubAckChan but cleanupOnClose()
	// is trying to push the error to this channel.
	ch := make(chan error, 1)
	_, err := sc.publishAsync(subject, data, nil, ch)
	if err == nil {
		err = <-ch
	}
	return err
}

// PublishAsync will publish to the cluster on pubPrefix+subject and asynchronously
// process the ACK or error state. It will return the GUID for the message being sent.
func (sc *conn) PublishAsync(subject string, data []byte, ah AckHandler) (string, error) {
	return sc.publishAsync(subject, data, ah, nil)
}

func (sc *conn) publishAsync(subject string, data []byte, ah AckHandler, ch chan error) (string, error) {
	a := &ack{ah: ah, ch: ch}
	sc.Lock()
	if sc.nc == nil {
		sc.Unlock()
		return "", ErrConnectionClosed
	}

	subj := sc.pubPrefix + "." + subject
	// This is only what we need from PubMsg in the timer below,
	// so do this so that pe doesn't escape.
	peGUID := sc.pubNUID.Next()
	// We send connID regardless of server we connect to. Older server
	// will simply not decode it.
	pe := &pb.PubMsg{ClientID: sc.clientID, Guid: peGUID, Subject: subject, Data: data, ConnID: sc.connID}
	b, _ := pe.Marshal()

	// Map ack to guid.
	sc.pubAckMap[peGUID] = a
	// snapshot
	ackSubject := sc.ackSubject
	ackTimeout := sc.opts.AckTimeout
	pac := sc.pubAckChan
	nc := sc.nc
	sc.Unlock()

	// Use the buffered channel to control the number of outstanding acks.
	pac <- struct{}{}

	err := nc.PublishRequest(subj, ackSubject, b)

	// Setup the timer for expiration.
	sc.Lock()
	if err != nil || sc.nc == nil {
		sc.Unlock()
		// If we got and error on publish or the connection has been closed,
		// we need to return an error only if:
		// - we can remove the pubAck from the map
		// - we can't, but this is an async pub with no provided AckHandler
		removed := sc.removeAck(peGUID) != nil
		if removed || (ch == nil && ah == nil) {
			if err == nil {
				err = ErrConnectionClosed
			}
			return "", err
		}
		// pubAck was removed from cleanupOnClose() and error will be sent
		// to appropriate go channel (ah or ch).
		return peGUID, nil
	}
	a.t = time.AfterFunc(ackTimeout, func() {
		pubAck := sc.removeAck(peGUID)
		// processAck could get here before and handle the ack.
		// If that's the case, we would get nil here and simply return.
		if pubAck == nil {
			return
		}
		if pubAck.ah != nil {
			pubAck.ah(peGUID, ErrTimeout)
		} else if a.ch != nil {
			pubAck.ch <- ErrTimeout
		}
	})
	sc.Unlock()

	return peGUID, nil
}

// removeAck removes the ack from the pubAckMap and cancels any state, e.g. timers
func (sc *conn) removeAck(guid string) *ack {
	var t *time.Timer
	sc.Lock()
	a := sc.pubAckMap[guid]
	if a != nil {
		t = a.t
		delete(sc.pubAckMap, guid)
	}
	pac := sc.pubAckChan
	sc.Unlock()

	// Cancel timer if needed.
	if t != nil {
		t.Stop()
	}

	// Remove from channel to unblock PublishAsync
	if a != nil && len(pac) > 0 {
		<-pac
	}
	return a
}

// Process an msg from the NATS Streaming cluster
func (sc *conn) processMsg(raw *nats.Msg) {
	msg := &Msg{}
	err := msg.Unmarshal(raw.Data)
	if err != nil {
		panic(fmt.Errorf("Error processing unmarshal for msg: %v", err))
	}
	// Lookup the subscription
	sc.RLock()
	nc := sc.nc
	isClosed := nc == nil
	sub := sc.subMap[raw.Subject]
	sc.RUnlock()

	// Check if sub is no longer valid or connection has been closed.
	if sub == nil || isClosed {
		return
	}

	// Store in msg for backlink
	msg.Sub = sub

	sub.RLock()
	cb := sub.cb
	ackSubject := sub.ackInbox
	isManualAck := sub.opts.ManualAcks
	subsc := sub.sc // Can be nil if sub has been unsubscribed.
	sub.RUnlock()

	// Perform the callback
	if cb != nil && subsc != nil {
		cb(msg)
	}

	// Process auto-ack
	if !isManualAck && nc != nil {
		ack := &pb.Ack{Subject: msg.Subject, Sequence: msg.Sequence}
		b, _ := ack.Marshal()
		// FIXME(dlc) - Async error handler? Retry?
		nc.Publish(ackSubject, b)
	}
}
