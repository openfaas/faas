// Copyright 2012-2018 The NATS Authors
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

package test

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nats-io/gnatsd/server"
	"github.com/nats-io/gnatsd/test"
	"github.com/nats-io/go-nats"
)

func TestAuth(t *testing.T) {
	opts := test.DefaultTestOptions
	opts.Port = 8232
	opts.Username = "derek"
	opts.Password = "foo"
	s := RunServerWithOptions(opts)
	defer s.Shutdown()

	_, err := nats.Connect("nats://localhost:8232")
	if err == nil {
		t.Fatal("Should have received an error while trying to connect")
	}

	// This test may be a bit too strict for the future, but for now makes
	// sure that we correctly process the -ERR content on connect.
	if err.Error() != nats.ErrAuthorization.Error() {
		t.Fatalf("Expected error '%v', got '%v'", nats.ErrAuthorization, err)
	}

	nc, err := nats.Connect("nats://derek:foo@localhost:8232")
	if err != nil {
		t.Fatal("Should have connected successfully with a token")
	}
	nc.Close()

	// Use Options
	nc, err = nats.Connect("nats://localhost:8232", nats.UserInfo("derek", "foo"))
	if err != nil {
		t.Fatalf("Should have connected successfully with a token: %v", err)
	}
	nc.Close()
	// Verify that credentials in URL take precedence.
	nc, err = nats.Connect("nats://derek:foo@localhost:8232", nats.UserInfo("foo", "bar"))
	if err != nil {
		t.Fatalf("Should have connected successfully with a token: %v", err)
	}
	nc.Close()
}

func TestAuthFailNoDisconnectCB(t *testing.T) {
	opts := test.DefaultTestOptions
	opts.Port = 8232
	opts.Username = "derek"
	opts.Password = "foo"
	s := RunServerWithOptions(opts)
	defer s.Shutdown()

	copts := nats.GetDefaultOptions()
	copts.Url = "nats://localhost:8232"
	receivedDisconnectCB := int32(0)
	copts.DisconnectedCB = func(nc *nats.Conn) {
		atomic.AddInt32(&receivedDisconnectCB, 1)
	}

	_, err := copts.Connect()
	if err == nil {
		t.Fatal("Should have received an error while trying to connect")
	}
	if atomic.LoadInt32(&receivedDisconnectCB) > 0 {
		t.Fatal("Should not have received a disconnect callback on auth failure")
	}
}

func TestAuthFailAllowReconnect(t *testing.T) {
	ts := RunServerOnPort(23232)
	defer ts.Shutdown()

	var servers = []string{
		"nats://localhost:23232",
		"nats://localhost:23233",
		"nats://localhost:23234",
	}

	ots2 := test.DefaultTestOptions
	ots2.Port = 23233
	ots2.Username = "ivan"
	ots2.Password = "foo"
	ts2 := RunServerWithOptions(ots2)
	defer ts2.Shutdown()

	ts3 := RunServerOnPort(23234)
	defer ts3.Shutdown()

	reconnectch := make(chan bool)

	opts := nats.GetDefaultOptions()
	opts.Servers = servers
	opts.AllowReconnect = true
	opts.NoRandomize = true
	opts.MaxReconnect = 10
	opts.ReconnectWait = 100 * time.Millisecond

	opts.ReconnectedCB = func(_ *nats.Conn) {
		reconnectch <- true
	}

	// Connect
	nc, err := opts.Connect()
	if err != nil {
		t.Fatalf("Should have connected ok: %v", err)
	}
	defer nc.Close()

	// Stop the server
	ts.Shutdown()

	// The client will try to connect to the second server, and that
	// should fail. It should then try to connect to the third and succeed.

	// Wait for the reconnect CB.
	if e := Wait(reconnectch); e != nil {
		t.Fatal("Reconnect callback should have been triggered")
	}

	if nc.IsClosed() {
		t.Fatal("Should have reconnected")
	}

	if nc.ConnectedUrl() != servers[2] {
		t.Fatalf("Should have reconnected to %s, reconnected to %s instead", servers[2], nc.ConnectedUrl())
	}
}

func TestTokenAuth(t *testing.T) {
	opts := test.DefaultTestOptions
	opts.Port = 8232
	secret := "S3Cr3T0k3n!"
	opts.Authorization = secret
	s := RunServerWithOptions(opts)
	defer s.Shutdown()

	_, err := nats.Connect("nats://localhost:8232")
	if err == nil {
		t.Fatal("Should have received an error while trying to connect")
	}

	tokenURL := fmt.Sprintf("nats://%s@localhost:8232", secret)
	nc, err := nats.Connect(tokenURL)
	if err != nil {
		t.Fatal("Should have connected successfully")
	}
	nc.Close()

	// Use Options
	nc, err = nats.Connect("nats://localhost:8232", nats.Token(secret))
	if err != nil {
		t.Fatalf("Should have connected successfully: %v", err)
	}
	nc.Close()
	// Verify that token in the URL takes precedence.
	nc, err = nats.Connect(tokenURL, nats.Token("badtoken"))
	if err != nil {
		t.Fatalf("Should have connected successfully: %v", err)
	}
	nc.Close()
}

func TestPermViolation(t *testing.T) {
	opts := test.DefaultTestOptions
	opts.Port = 8232
	opts.Users = []*server.User{
		&server.User{
			Username: "ivan",
			Password: "pwd",
			Permissions: &server.Permissions{
				Publish:   []string{"foo"},
				Subscribe: []string{"bar"},
			},
		},
	}
	s := RunServerWithOptions(opts)
	defer s.Shutdown()

	errCh := make(chan error, 2)
	errCB := func(_ *nats.Conn, _ *nats.Subscription, err error) {
		errCh <- err
	}
	nc, err := nats.Connect(
		fmt.Sprintf("nats://ivan:pwd@localhost:%d", opts.Port),
		nats.ErrorHandler(errCB))
	if err != nil {
		t.Fatalf("Error on connect: %v", err)
	}
	defer nc.Close()

	// Cause a publish error
	nc.Publish("bar", []byte("fail"))
	// Cause a subscribe error
	nc.Subscribe("foo", func(_ *nats.Msg) {})

	expectedErrorTypes := []string{"publish", "subscription"}
	for _, expectedErr := range expectedErrorTypes {
		select {
		case e := <-errCh:
			if !strings.Contains(e.Error(), nats.PERMISSIONS_ERR) {
				t.Fatalf("Did not receive error about permissions")
			}
			if !strings.Contains(e.Error(), expectedErr) {
				t.Fatalf("Did not receive error about %q, got %v", expectedErr, e.Error())
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("Did not get the permission error")
		}
	}
	// Make sure connection has not been closed
	if nc.IsClosed() {
		t.Fatal("Connection should be not be closed")
	}
}
