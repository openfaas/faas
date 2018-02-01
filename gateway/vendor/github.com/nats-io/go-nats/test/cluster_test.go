package test

import (
	"math"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

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
