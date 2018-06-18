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

// +build go1.7

// A Go client for the NATS messaging system (https://nats.io).
package nats

import (
	"context"
	"fmt"
	"reflect"
)

// RequestWithContext takes a context, a subject and payload
// in bytes and request expecting a single response.
func (nc *Conn) RequestWithContext(ctx context.Context, subj string, data []byte) (*Msg, error) {
	if ctx == nil {
		return nil, ErrInvalidContext
	}
	if nc == nil {
		return nil, ErrInvalidConnection
	}

	nc.mu.Lock()
	// If user wants the old style.
	if nc.Opts.UseOldRequestStyle {
		nc.mu.Unlock()
		return nc.oldRequestWithContext(ctx, subj, data)
	}

	// Do setup for the new style.
	if nc.respMap == nil {
		// _INBOX wildcard
		nc.respSub = fmt.Sprintf("%s.*", NewInbox())
		nc.respMap = make(map[string]chan *Msg)
	}
	// Create literal Inbox and map to a chan msg.
	mch := make(chan *Msg, RequestChanLen)
	respInbox := nc.newRespInbox()
	token := respToken(respInbox)
	nc.respMap[token] = mch
	createSub := nc.respMux == nil
	ginbox := nc.respSub
	nc.mu.Unlock()

	if createSub {
		// Make sure scoped subscription is setup only once.
		var err error
		nc.respSetup.Do(func() { err = nc.createRespMux(ginbox) })
		if err != nil {
			return nil, err
		}
	}

	err := nc.PublishRequest(subj, respInbox, data)
	if err != nil {
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
		nc.mu.Lock()
		delete(nc.respMap, token)
		nc.mu.Unlock()
		return nil, ctx.Err()
	}

	return msg, nil
}

// oldRequestWithContext utilizes inbox and subscription per request.
func (nc *Conn) oldRequestWithContext(ctx context.Context, subj string, data []byte) (*Msg, error) {
	inbox := NewInbox()
	ch := make(chan *Msg, RequestChanLen)

	s, err := nc.subscribe(inbox, _EMPTY_, nil, ch)
	if err != nil {
		return nil, err
	}
	s.AutoUnsubscribe(1)
	defer s.Unsubscribe()

	err = nc.PublishRequest(subj, inbox, data)
	if err != nil {
		return nil, err
	}

	return s.NextMsgWithContext(ctx)
}

// NextMsgWithContext takes a context and returns the next message
// available to a synchronous subscriber, blocking until it is delivered
// or context gets canceled.
func (s *Subscription) NextMsgWithContext(ctx context.Context) (*Msg, error) {
	if ctx == nil {
		return nil, ErrInvalidContext
	}
	if s == nil {
		return nil, ErrBadSubscription
	}

	s.mu.Lock()
	err := s.validateNextMsgState()
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}

	// snapshot
	mch := s.mch
	s.mu.Unlock()

	var ok bool
	var msg *Msg

	select {
	case msg, ok = <-mch:
		if !ok {
			return nil, ErrConnectionClosed
		}
		err := s.processNextMsgDelivered(msg)
		if err != nil {
			return nil, err
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return msg, nil
}

// RequestWithContext will create an Inbox and perform a Request
// using the provided cancellation context with the Inbox reply
// for the data v. A response will be decoded into the vPtrResponse.
func (c *EncodedConn) RequestWithContext(ctx context.Context, subject string, v interface{}, vPtr interface{}) error {
	if ctx == nil {
		return ErrInvalidContext
	}

	b, err := c.Enc.Encode(subject, v)
	if err != nil {
		return err
	}
	m, err := c.Conn.RequestWithContext(ctx, subject, b)
	if err != nil {
		return err
	}
	if reflect.TypeOf(vPtr) == emptyMsgType {
		mPtr := vPtr.(*Msg)
		*mPtr = *m
	} else {
		err := c.Enc.Decode(m.Subject, m.Data, vPtr)
		if err != nil {
			return err
		}
	}

	return nil
}
