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
	"math"
	"net"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/gnatsd/server"
	"github.com/nats-io/gnatsd/test"
	"github.com/nats-io/go-nats"
)

var testServers = []string{
	"nats://localhost:1222",
	"nats://localhost:1223",
	"nats://localhost:1224",
	"nats://localhost:1225",
	"nats://localhost:1226",
	"nats://localhost:1227",
	"nats://localhost:1228",
}

var servers = strings.Join(testServers, ",")

func serverVersionAtLeast(major, minor, update int) error {
	var (
		ma, mi, up int
	)
	fmt.Sscanf(server.VERSION, "%d.%d.%d", &ma, &mi, &up)
	if ma > major || (ma == major && mi > minor) || (ma == major && mi == minor && up >= update) {
		return nil
	}
	return fmt.Errorf("Server version is %v, requires %d.%d.%d+", server.VERSION, major, minor, update)
}

func TestServersOption(t *testing.T) {
	opts := nats.GetDefaultOptions()
	opts.NoRandomize = true

	_, err := opts.Connect()
	if err != nats.ErrNoServers {
		t.Fatalf("Wrong error: '%v'\n", err)
	}
	opts.Servers = testServers
	_, err = opts.Connect()
	if err == nil || err != nats.ErrNoServers {
		t.Fatalf("Did not receive proper error: %v\n", err)
	}

	// Make sure we can connect to first server if running
	s1 := RunServerOnPort(1222)
	// Do this in case some failure occurs before explicit shutdown
	defer s1.Shutdown()

	nc, err := opts.Connect()
	if err != nil {
		t.Fatalf("Could not connect: %v\n", err)
	}
	if nc.ConnectedUrl() != "nats://localhost:1222" {
		nc.Close()
		t.Fatalf("Does not report correct connection: %s\n",
			nc.ConnectedUrl())
	}
	nc.Close()
	s1.Shutdown()

	// Make sure we can connect to a non first server if running
	s2 := RunServerOnPort(1223)
	// Do this in case some failure occurs before explicit shutdown
	defer s2.Shutdown()

	nc, err = opts.Connect()
	if err != nil {
		t.Fatalf("Could not connect: %v\n", err)
	}
	defer nc.Close()
	if nc.ConnectedUrl() != "nats://localhost:1223" {
		t.Fatalf("Does not report correct connection: %s\n",
			nc.ConnectedUrl())
	}
}

func TestNewStyleServersOption(t *testing.T) {
	_, err := nats.Connect(nats.DefaultURL, nats.DontRandomize())
	if err != nats.ErrNoServers {
		t.Fatalf("Wrong error: '%v'\n", err)
	}
	servers := strings.Join(testServers, ",")

	_, err = nats.Connect(servers, nats.DontRandomize())
	if err == nil || err != nats.ErrNoServers {
		t.Fatalf("Did not receive proper error: %v\n", err)
	}

	// Make sure we can connect to first server if running
	s1 := RunServerOnPort(1222)
	// Do this in case some failure occurs before explicit shutdown
	defer s1.Shutdown()

	nc, err := nats.Connect(servers, nats.DontRandomize())
	if err != nil {
		t.Fatalf("Could not connect: %v\n", err)
	}
	if nc.ConnectedUrl() != "nats://localhost:1222" {
		nc.Close()
		t.Fatalf("Does not report correct connection: %s\n",
			nc.ConnectedUrl())
	}
	nc.Close()
	s1.Shutdown()

	// Make sure we can connect to a non-first server if running
	s2 := RunServerOnPort(1223)
	// Do this in case some failure occurs before explicit shutdown
	defer s2.Shutdown()

	nc, err = nats.Connect(servers, nats.DontRandomize())
	if err != nil {
		t.Fatalf("Could not connect: %v\n", err)
	}
	defer nc.Close()
	if nc.ConnectedUrl() != "nats://localhost:1223" {
		t.Fatalf("Does not report correct connection: %s\n",
			nc.ConnectedUrl())
	}
}

func TestAuthServers(t *testing.T) {
	var plainServers = []string{
		"nats://localhost:1222",
		"nats://localhost:1224",
	}

	opts := test.DefaultTestOptions
	opts.Username = "derek"
	opts.Password = "foo"

	opts.Port = 1222
	as1 := RunServerWithOptions(opts)
	defer as1.Shutdown()
	opts.Port = 1224
	as2 := RunServerWithOptions(opts)
	defer as2.Shutdown()

	pservers := strings.Join(plainServers, ",")
	nc, err := nats.Connect(pservers, nats.DontRandomize(), nats.Timeout(5*time.Second))
	if err == nil {
		nc.Close()
		t.Fatalf("Expect Auth failure, got no error\n")
	}

	if matched, _ := regexp.Match(`authorization`, []byte(err.Error())); !matched {
		t.Fatalf("Wrong error, wanted Auth failure, got '%s'\n", err)
	}

	// Test that we can connect to a subsequent correct server.
	var authServers = []string{
		"nats://localhost:1222",
		"nats://derek:foo@localhost:1224",
	}
	aservers := strings.Join(authServers, ",")
	nc, err = nats.Connect(aservers, nats.DontRandomize(), nats.Timeout(5*time.Second))
	if err != nil {
		t.Fatalf("Expected to connect properly: %v\n", err)
	}
	defer nc.Close()
	if nc.ConnectedUrl() != authServers[1] {
		t.Fatalf("Does not report correct connection: %s\n",
			nc.ConnectedUrl())
	}
}

func TestBasicClusterReconnect(t *testing.T) {
	s1 := RunServerOnPort(1222)
	defer s1.Shutdown()
	s2 := RunServerOnPort(1224)
	defer s2.Shutdown()

	dch := make(chan bool)
	rch := make(chan bool)

	dcbCalled := false

	opts := []nats.Option{nats.DontRandomize(),
		nats.DisconnectHandler(func(nc *nats.Conn) {
			// Suppress any additional callbacks
			if dcbCalled {
				return
			}
			dcbCalled = true
			dch <- true
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) { rch <- true }),
	}

	nc, err := nats.Connect(servers, opts...)
	if err != nil {
		t.Fatalf("Expected to connect, got err: %v\n", err)
	}
	defer nc.Close()

	s1.Shutdown()

	// wait for disconnect
	if e := WaitTime(dch, 2*time.Second); e != nil {
		t.Fatal("Did not receive a disconnect callback message")
	}

	reconnectTimeStart := time.Now()

	// wait for reconnect
	if e := WaitTime(rch, 2*time.Second); e != nil {
		t.Fatal("Did not receive a reconnect callback message")
	}

	if nc.ConnectedUrl() != testServers[2] {
		t.Fatalf("Does not report correct connection: %s\n",
			nc.ConnectedUrl())
	}

	// Make sure we did not wait on reconnect for default time.
	// Reconnect should be fast since it will be a switch to the
	// second server and not be dependent on server restart time.

	// On Windows, a failed connect takes more than a second, so
	// account for that.
	maxDuration := 100 * time.Millisecond
	if runtime.GOOS == "windows" {
		maxDuration = 1100 * time.Millisecond
	}
	reconnectTime := time.Since(reconnectTimeStart)
	if reconnectTime > maxDuration {
		t.Fatalf("Took longer than expected to reconnect: %v\n", reconnectTime)
	}
}

func TestHotSpotReconnect(t *testing.T) {
	s1 := RunServerOnPort(1222)
	defer s1.Shutdown()

	var srvrs string
	if runtime.GOOS == "windows" {
		srvrs = strings.Join(testServers[:5], ",")
	} else {
		srvrs = servers
	}

	numClients := 32
	clients := []*nats.Conn{}

	wg := &sync.WaitGroup{}
	wg.Add(numClients)

	opts := []nats.Option{
		nats.ReconnectWait(50 * time.Millisecond),
		nats.ReconnectHandler(func(_ *nats.Conn) { wg.Done() }),
	}

	for i := 0; i < numClients; i++ {
		nc, err := nats.Connect(srvrs, opts...)
		if err != nil {
			t.Fatalf("Expected to connect, got err: %v\n", err)
		}
		defer nc.Close()
		if nc.ConnectedUrl() != testServers[0] {
			t.Fatalf("Connected to incorrect server: %v\n", nc.ConnectedUrl())
		}
		clients = append(clients, nc)
	}

	s2 := RunServerOnPort(1224)
	defer s2.Shutdown()
	s3 := RunServerOnPort(1226)
	defer s3.Shutdown()

	s1.Shutdown()

	numServers := 2

	// Wait on all reconnects
	wg.Wait()

	// Walk the clients and calculate how many of each..
	cs := make(map[string]int)
	for _, nc := range clients {
		cs[nc.ConnectedUrl()]++
		nc.Close()
	}
	if len(cs) != numServers {
		t.Fatalf("Wrong number of reported servers: %d vs %d\n", len(cs), numServers)
	}
	expected := numClients / numServers
	v := uint(float32(expected) * 0.40)

	// Check that each item is within acceptable range
	for s, total := range cs {
		delta := uint(math.Abs(float64(expected - total)))
		if delta > v {
			t.Fatalf("Connected clients to server: %s out of range: %d\n", s, total)
		}
	}
}

func TestProperReconnectDelay(t *testing.T) {
	s1 := RunServerOnPort(1222)
	defer s1.Shutdown()

	var srvs string
	opts := nats.GetDefaultOptions()
	if runtime.GOOS == "windows" {
		srvs = strings.Join(testServers[:2], ",")
	} else {
		srvs = strings.Join(testServers, ",")
	}
	opts.NoRandomize = true

	dcbCalled := false
	closedCbCalled := false
	dch := make(chan bool)

	dcb := func(nc *nats.Conn) {
		// Suppress any additional calls
		if dcbCalled {
			return
		}
		dcbCalled = true
		dch <- true
	}

	ccb := func(_ *nats.Conn) {
		closedCbCalled = true
	}

	nc, err := nats.Connect(srvs, nats.DontRandomize(), nats.DisconnectHandler(dcb), nats.ClosedHandler(ccb))
	if err != nil {
		t.Fatalf("Expected to connect, got err: %v\n", err)
	}
	defer nc.Close()

	s1.Shutdown()

	// wait for disconnect
	if e := WaitTime(dch, 2*time.Second); e != nil {
		t.Fatal("Did not receive a disconnect callback message")
	}

	// Wait, want to make sure we don't spin on reconnect to non-existent servers.
	time.Sleep(1 * time.Second)

	// Make sure we are still reconnecting..
	if closedCbCalled {
		t.Fatal("Closed CB was triggered, should not have been.")
	}
	if status := nc.Status(); status != nats.RECONNECTING {
		t.Fatalf("Wrong status: %d\n", status)
	}
}

func TestProperFalloutAfterMaxAttempts(t *testing.T) {
	s1 := RunServerOnPort(1222)
	defer s1.Shutdown()

	opts := nats.GetDefaultOptions()
	// Reduce the list of servers for Windows tests
	if runtime.GOOS == "windows" {
		opts.Servers = testServers[:2]
		opts.MaxReconnect = 2
	} else {
		opts.Servers = testServers
		opts.MaxReconnect = 5
	}
	opts.NoRandomize = true
	opts.ReconnectWait = (25 * time.Millisecond)

	dch := make(chan bool)
	opts.DisconnectedCB = func(_ *nats.Conn) {
		dch <- true
	}

	closedCbCalled := false
	cch := make(chan bool)

	opts.ClosedCB = func(_ *nats.Conn) {
		closedCbCalled = true
		cch <- true
	}

	nc, err := opts.Connect()
	if err != nil {
		t.Fatalf("Expected to connect, got err: %v\n", err)
	}
	defer nc.Close()

	s1.Shutdown()

	// On Windows, creating a TCP connection to a server not running takes more than
	// a second. So be generous with the WaitTime.

	// wait for disconnect
	if e := WaitTime(dch, 5*time.Second); e != nil {
		t.Fatal("Did not receive a disconnect callback message")
	}

	// Wait for ClosedCB
	if e := WaitTime(cch, 5*time.Second); e != nil {
		t.Fatal("Did not receive a closed callback message")
	}

	// Make sure we are not still reconnecting..
	if !closedCbCalled {
		t.Logf("%+v\n", nc)
		t.Fatal("Closed CB was not triggered, should have been.")
	}

	// Expect connection to be closed...
	if !nc.IsClosed() {
		t.Fatalf("Wrong status: %d\n", nc.Status())
	}
}

func TestProperFalloutAfterMaxAttemptsWithAuthMismatch(t *testing.T) {
	var myServers = []string{
		"nats://localhost:1222",
		"nats://localhost:4443",
	}
	s1 := RunServerOnPort(1222)
	defer s1.Shutdown()

	s2, _ := RunServerWithConfig("./configs/tlsverify.conf")
	defer s2.Shutdown()

	opts := nats.GetDefaultOptions()
	opts.Servers = myServers
	opts.NoRandomize = true
	if runtime.GOOS == "windows" {
		opts.MaxReconnect = 2
	} else {
		opts.MaxReconnect = 5
	}
	opts.ReconnectWait = (25 * time.Millisecond)

	dch := make(chan bool)
	opts.DisconnectedCB = func(_ *nats.Conn) {
		dch <- true
	}

	closedCbCalled := false
	cch := make(chan bool)

	opts.ClosedCB = func(_ *nats.Conn) {
		closedCbCalled = true
		cch <- true
	}

	nc, err := opts.Connect()
	if err != nil {
		t.Fatalf("Expected to connect, got err: %v\n", err)
	}
	defer nc.Close()

	s1.Shutdown()

	// On Windows, creating a TCP connection to a server not running takes more than
	// a second. So be generous with the WaitTime.

	// wait for disconnect
	if e := WaitTime(dch, 5*time.Second); e != nil {
		t.Fatal("Did not receive a disconnect callback message")
	}

	// Wait for ClosedCB
	if e := WaitTime(cch, 5*time.Second); e != nil {
		reconnects := nc.Stats().Reconnects
		t.Fatalf("Did not receive a closed callback message, #reconnects: %v", reconnects)
	}

	// Make sure we have not exceeded MaxReconnect
	reconnects := nc.Stats().Reconnects
	if reconnects != uint64(opts.MaxReconnect) {
		t.Fatalf("Num reconnects was %v, expected %v", reconnects, opts.MaxReconnect)
	}

	// Make sure we are not still reconnecting..
	if !closedCbCalled {
		t.Logf("%+v\n", nc)
		t.Fatal("Closed CB was not triggered, should have been.")
	}

	// Expect connection to be closed...
	if !nc.IsClosed() {
		t.Fatalf("Wrong status: %d\n", nc.Status())
	}
}

func TestTimeoutOnNoServers(t *testing.T) {
	s1 := RunServerOnPort(1222)
	defer s1.Shutdown()

	opts := nats.GetDefaultOptions()
	if runtime.GOOS == "windows" {
		opts.Servers = testServers[:2]
		opts.MaxReconnect = 2
		opts.ReconnectWait = (100 * time.Millisecond)
	} else {
		opts.Servers = testServers
		// 1 second total time wait
		opts.MaxReconnect = 10
		opts.ReconnectWait = (100 * time.Millisecond)
	}
	opts.NoRandomize = true

	dch := make(chan bool)
	opts.DisconnectedCB = func(nc *nats.Conn) {
		// Suppress any additional calls
		nc.SetDisconnectHandler(nil)
		dch <- true
	}

	cch := make(chan bool)
	opts.ClosedCB = func(_ *nats.Conn) {
		cch <- true
	}

	nc, err := opts.Connect()
	if err != nil {
		t.Fatalf("Expected to connect, got err: %v\n", err)
	}
	defer nc.Close()

	s1.Shutdown()

	// On Windows, creating a connection to a non-running server takes
	// more than a second. So be generous with WaitTime

	// wait for disconnect
	if e := WaitTime(dch, 5*time.Second); e != nil {
		t.Fatal("Did not receive a disconnect callback message")
	}

	startWait := time.Now()

	// Wait for ClosedCB
	if e := WaitTime(cch, 5*time.Second); e != nil {
		t.Fatal("Did not receive a closed callback message")
	}

	if runtime.GOOS != "windows" {
		timeWait := time.Since(startWait)

		// Use 500ms as variable time delta
		variable := (500 * time.Millisecond)
		expected := (time.Duration(opts.MaxReconnect) * opts.ReconnectWait)

		if timeWait > (expected + variable) {
			t.Fatalf("Waited too long for Closed state: %d\n", timeWait/time.Millisecond)
		}
	}
}

func TestPingReconnect(t *testing.T) {
	RECONNECTS := 4
	s1 := RunServerOnPort(1222)
	defer s1.Shutdown()

	opts := nats.GetDefaultOptions()
	opts.Servers = testServers
	opts.NoRandomize = true
	opts.ReconnectWait = 200 * time.Millisecond
	opts.PingInterval = 50 * time.Millisecond
	opts.MaxPingsOut = -1

	var wg sync.WaitGroup
	wg.Add(1)
	rch := make(chan time.Time, RECONNECTS)
	dch := make(chan time.Time, RECONNECTS)

	opts.DisconnectedCB = func(_ *nats.Conn) {
		d := dch
		select {
		case d <- time.Now():
		default:
			d = nil
		}
	}

	opts.ReconnectedCB = func(c *nats.Conn) {
		r := rch
		select {
		case r <- time.Now():
		default:
			r = nil
			wg.Done()
		}
	}

	nc, err := opts.Connect()
	if err != nil {
		t.Fatalf("Expected to connect, got err: %v\n", err)
	}
	defer nc.Close()

	wg.Wait()
	s1.Shutdown()

	<-dch
	for i := 0; i < RECONNECTS-1; i++ {
		disconnectedAt := <-dch
		reconnectAt := <-rch
		pingCycle := disconnectedAt.Sub(reconnectAt)
		if pingCycle > 2*opts.PingInterval {
			t.Fatalf("Reconnect due to ping took %s", pingCycle.String())
		}
	}
}

type checkPoolUpdatedDialer struct {
	conn         net.Conn
	first, final bool
	ra           int
}

func (d *checkPoolUpdatedDialer) Dial(network, address string) (net.Conn, error) {
	doReal := false
	if d.first {
		d.first = false
		doReal = true
	} else if d.final {
		d.ra++
		return nil, fmt.Errorf("On purpose")
	} else {
		d.ra++
		if d.ra == 15 {
			d.ra = 0
			doReal = true
		}
	}
	if doReal {
		c, err := net.Dial(network, address)
		if err != nil {
			return nil, err
		}
		d.conn = c
		return c, nil
	}
	return nil, fmt.Errorf("On purpose")
}

func TestServerPoolUpdatedWhenRouteGoesAway(t *testing.T) {
	if err := serverVersionAtLeast(1, 0, 7); err != nil {
		t.Skipf(err.Error())
	}
	s1Opts := test.DefaultTestOptions
	s1Opts.Host = "127.0.0.1"
	s1Opts.Port = 4222
	s1Opts.Cluster.Host = "127.0.0.1"
	s1Opts.Cluster.Port = 6222
	s1Opts.Routes = server.RoutesFromStr("nats://127.0.0.1:6223,nats://127.0.0.1:6224")
	s1 := test.RunServer(&s1Opts)
	defer s1.Shutdown()

	s1Url := "nats://127.0.0.1:4222"
	s2Url := "nats://127.0.0.1:4223"
	s3Url := "nats://127.0.0.1:4224"

	ch := make(chan bool, 1)
	chch := make(chan bool, 1)
	connHandler := func(_ *nats.Conn) {
		chch <- true
	}
	nc, err := nats.Connect(s1Url,
		nats.ReconnectHandler(connHandler),
		nats.DiscoveredServersHandler(func(_ *nats.Conn) {
			ch <- true
		}))
	if err != nil {
		t.Fatalf("Error on connect")
	}

	s2Opts := test.DefaultTestOptions
	s2Opts.Host = "127.0.0.1"
	s2Opts.Port = s1Opts.Port + 1
	s2Opts.Cluster.Host = "127.0.0.1"
	s2Opts.Cluster.Port = 6223
	s2Opts.Routes = server.RoutesFromStr("nats://127.0.0.1:6222,nats://127.0.0.1:6224")
	s2 := test.RunServer(&s2Opts)
	defer s2.Shutdown()

	// Wait to be notified
	if err := Wait(ch); err != nil {
		t.Fatal("New server callback was not invoked")
	}

	checkPool := func(expected []string) {
		// Don't use discovered here, but Servers to have the full list.
		// Also, there may be cases where the mesh is not formed yet,
		// so try again on failure.
		var (
			ds      []string
			timeout = time.Now().Add(5 * time.Second)
		)
		for time.Now().Before(timeout) {
			ds = nc.Servers()
			if len(ds) == len(expected) {
				m := make(map[string]struct{}, len(ds))
				for _, url := range ds {
					m[url] = struct{}{}
				}
				ok := true
				for _, url := range expected {
					if _, present := m[url]; !present {
						ok = false
						break
					}
				}
				if ok {
					return
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
		stackFatalf(t, "Expected %v, got %v", expected, ds)
	}
	// Verify that we now know about s2
	checkPool([]string{s1Url, s2Url})

	s3Opts := test.DefaultTestOptions
	s3Opts.Host = "127.0.0.1"
	s3Opts.Port = s2Opts.Port + 1
	s3Opts.Cluster.Host = "127.0.0.1"
	s3Opts.Cluster.Port = 6224
	s3Opts.Routes = server.RoutesFromStr("nats://127.0.0.1:6222,nats://127.0.0.1:6223")
	s3 := test.RunServer(&s3Opts)
	defer s3.Shutdown()

	// Wait to be notified
	if err := Wait(ch); err != nil {
		t.Fatal("New server callback was not invoked")
	}
	// Verify that we now know about s3
	checkPool([]string{s1Url, s2Url, s3Url})

	// Stop s1. Since this was passed to the Connect() call, this one should
	// still be present.
	s1.Shutdown()
	// Wait for reconnect
	if err := Wait(chch); err != nil {
		t.Fatal("Reconnect handler not invoked")
	}
	checkPool([]string{s1Url, s2Url, s3Url})

	// Check the server we reconnected to.
	reConnectedTo := nc.ConnectedUrl()
	expected := []string{s1Url}
	restartS2 := false
	if reConnectedTo == s2Url {
		restartS2 = true
		s2.Shutdown()
		expected = append(expected, s3Url)
	} else if reConnectedTo == s3Url {
		s3.Shutdown()
		expected = append(expected, s2Url)
	} else {
		t.Fatalf("Unexpected server client has reconnected to: %v", reConnectedTo)
	}
	// Wait for reconnect
	if err := Wait(chch); err != nil {
		t.Fatal("Reconnect handler not invoked")
	}
	// The implicit server that we just shutdown should have been removed from the pool
	checkPool(expected)

	// Restart the one that was shutdown and check that it is now back in the pool
	if restartS2 {
		s2 = test.RunServer(&s2Opts)
		defer s2.Shutdown()
		expected = append(expected, s2Url)
	} else {
		s3 = test.RunServer(&s3Opts)
		defer s3.Shutdown()
		expected = append(expected, s3Url)
	}
	// Since this is not a "new" server, the DiscoveredServersCB won't be invoked.
	checkPool(expected)

	nc.Close()

	// Restart s1
	s1 = test.RunServer(&s1Opts)
	defer s1.Shutdown()

	// We should have all 3 servers running now...

	// Create a client connection with special dialer.
	d := &checkPoolUpdatedDialer{first: true}
	nc, err = nats.Connect(s1Url,
		nats.MaxReconnects(10),
		nats.ReconnectWait(15*time.Millisecond),
		nats.SetCustomDialer(d),
		nats.ReconnectHandler(connHandler),
		nats.ClosedHandler(connHandler))
	if err != nil {
		t.Fatalf("Error on connect")
	}
	defer nc.Close()

	// Make sure that we have all 3 servers in the pool (this will wait if required)
	checkPool(expected)

	// Cause disconnection between client and server. We are going to reconnect
	// and we want to check that when we get the INFO again with the list of
	// servers, we don't lose the knowledge of how many times we tried to
	// reconnect.
	d.conn.Close()

	// Wait for client to reconnect to a server
	if err := Wait(chch); err != nil {
		t.Fatal("Reconnect handler not invoked")
	}
	// At this point, we should have tried to reconnect 5 times to each server.
	// For the one we reconnected to, its max reconnect attempts should have been
	// cleared, not for the other ones.

	// Cause a disconnect again and ensure we won't reconnect.
	d.final = true
	d.conn.Close()

	// Wait for Close callback to be invoked.
	if err := Wait(chch); err != nil {
		t.Fatal("Close handler not invoked")
	}

	// Since MaxReconnect is 10, after trying 5 more times on 2 of the servers,
	// these should have been removed. We have still 5 more tries for the server
	// we did previously reconnect to.
	// So total of reconnect attempt should be: 2*5+1*10=20
	if d.ra != 20 {
		t.Fatalf("Should have tried to reconnect 20 more times, got %v", d.ra)
	}

	nc.Close()
}
