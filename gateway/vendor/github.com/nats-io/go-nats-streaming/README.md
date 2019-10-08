# NATS Streaming

NATS Streaming is an extremely performant, lightweight reliable streaming platform powered by [NATS](https://nats.io).

[![License Apache 2](https://img.shields.io/badge/License-Apache2-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)
[![Build Status](https://travis-ci.org/nats-io/go-nats-streaming.svg?branch=master)](http://travis-ci.org/nats-io/go-nats-streaming)
[![Coverage Status](https://coveralls.io/repos/nats-io/go-nats-streaming/badge.svg?branch=master)](https://coveralls.io/r/nats-io/go-nats-streaming?branch=master)

NATS Streaming provides the following high-level feature set:
- Log based persistence
- At-Least-Once Delivery model, giving reliable message delivery
- Rate matched on a per subscription basis
- Replay/Restart
- Last Value Semantics

## Notes

- Please raise questions/issues via the [Issue Tracker](https://github.com/nats-io/go-nats-streaming/issues).

## Installation

```bash
# Go client
go get github.com/nats-io/go-nats-streaming
```

## Basic Usage

```go

sc, _ := stan.Connect(clusterID, clientID)

// Simple Synchronous Publisher
sc.Publish("foo", []byte("Hello World")) // does not return until an ack has been received from NATS Streaming

// Simple Async Subscriber
sub, _ := sc.Subscribe("foo", func(m *stan.Msg) {
    fmt.Printf("Received a message: %s\n", string(m.Data))
})

// Unsubscribe
sub.Unsubscribe()

// Close connection
sc.Close()
```

### Subscription Start (i.e. Replay) Options

NATS Streaming subscriptions are similar to NATS subscriptions, but clients may start their subscription at an earlier point in the message stream, allowing them to receive messages that were published before this client registered interest.

The options are described with examples below:

```go

// Subscribe starting with most recently published value
sub, err := sc.Subscribe("foo", func(m *stan.Msg) {
    fmt.Printf("Received a message: %s\n", string(m.Data))
}, stan.StartWithLastReceived())

// Receive all stored values in order
sub, err := sc.Subscribe("foo", func(m *stan.Msg) {
    fmt.Printf("Received a message: %s\n", string(m.Data))
}, stan.DeliverAllAvailable())

// Receive messages starting at a specific sequence number
sub, err := sc.Subscribe("foo", func(m *stan.Msg) {
    fmt.Printf("Received a message: %s\n", string(m.Data))
}, stan.StartAtSequence(22))

// Subscribe starting at a specific time
var startTime time.Time
...
sub, err := sc.Subscribe("foo", func(m *stan.Msg) {
    fmt.Printf("Received a message: %s\n", string(m.Data))
}, stan.StartAtTime(startTime))

// Subscribe starting a specific amount of time in the past (e.g. 30 seconds ago)
sub, err := sc.Subscribe("foo", func(m *stan.Msg) {
    fmt.Printf("Received a message: %s\n", string(m.Data))
}, stan.StartAtTimeDelta(time.ParseDuration("30s")))
```

### Durable Subscriptions

Replay of messages offers great flexibility for clients wishing to begin processing at some earlier point in the data stream.
However, some clients just need to pick up where they left off from an earlier session, without having to manually track their position in the stream of messages.
Durable subscriptions allow clients to assign a durable name to a subscription when it is created.
Doing this causes the NATS Streaming server to track the last acknowledged message for that clientID + durable name, so that only messages since the last acknowledged message will be delivered to the client.

```go
sc, _ := stan.Connect("test-cluster", "client-123")

// Subscribe with durable name
sc.Subscribe("foo", func(m *stan.Msg) {
    fmt.Printf("Received a message: %s\n", string(m.Data))
}, stan.DurableName("my-durable"))
...
// client receives message sequence 1-40
...
// client disconnects for an hour
...
// client reconnects with same clientID "client-123"
sc, _ := stan.Connect("test-cluster", "client-123")

// client re-subscribes to "foo" with same durable name "my-durable"
sc.Subscribe("foo", func(m *stan.Msg) {
    fmt.Printf("Received a message: %s\n", string(m.Data))
}, stan.DurableName("my-durable"))
...
// client receives messages 41-current
```

### Queue Groups

All subscriptions with the same queue name (regardless of the connection
they originate from) will form a queue group.
Each message will be delivered to only one subscriber per queue group,
using queuing semantics. You can have as many queue groups as you wish.

Normal subscribers will continue to work as expected.

#### Creating a Queue Group

A queue group is automatically created when the first queue subscriber is
created. If the group already exists, the member is added to the group.

```go
sc, _ := stan.Connect("test-cluster", "clientid")

// Create a queue subscriber on "foo" for group "bar"
qsub1, _ := sc.QueueSubscribe("foo", "bar", qcb)

// Add a second member
qsub2, _ := sc.QueueSubscribe("foo", "bar", qcb)

// Notice that you can have a regular subscriber on that subject
sub, _ := sc.Subscribe("foo", cb)

// A message on "foo" will be received by sub and qsub1 or qsub2.
```

#### Start Position

Note that once a queue group is formed, a member's start position is ignored
when added to the group. It will start receive messages from the last
position in the group.

Suppose the channel `foo` exists and there are `500` messages stored, the group
`bar` is already created, there are two members and the last
message sequence sent is `100`. A new member is added. Note its start position:

```go
sc.QueueSubscribe("foo", "bar", qcb, stan.StartAtSequence(200))
```

This will not produce an error, but the start position will be ignored. Assuming
this member would be the one receiving the next message, it would receive message
sequence `101`.

#### Leaving the Group

There are two ways of leaving the group: closing the subscriber's connection or
calling `Unsubscribe`:

```go
// Have qsub leave the queue group
qsub.Unsubscribe()
```

If the leaving member had un-acknowledged messages, those messages are reassigned
to the remaining members.

#### Closing a Queue Group

There is no special API for that. Once all members have left (either calling `Unsubscribe`,
or their connections are closed), the group is removed from the server.

The next call to `QueueSubscribe` with the same group name will create a brand new group,
that is, the start position will take effect and delivery will start from there.

### Durable Queue Groups

As described above, for non durable queue subscribers, when the last member leaves the group,
that group is removed. A durable queue group allows you to have all members leave but still
maintain state. When a member re-joins, it starts at the last position in that group.

#### Creating a Durable Queue Group

A durable queue group is created in a similar manner as that of a standard queue group,
except the `DurableName` option must be used to specify durability.

```go
sc.QueueSubscribe("foo", "bar", qcb, stan.DurableName("dur"))
```
A group called `dur:bar` (the concatenation of durable name and group name) is created in
the server. This means two things:

- The character `:` is not allowed for a queue subscriber's durable name.
- Durable and non-durable queue groups with the same name can coexist.

```go
// Non durable queue subscriber on group "bar"
qsub, _ := sc.QueueSubscribe("foo", "bar", qcb)

// Durable queue subscriber on group "bar"
durQsub, _ := sc.QueueSubscribe("foo", "bar", qcb, stan.DurableName("mydurablegroup"))

// The same message produced on "foo" would be received by both queue subscribers.
```

#### Start Position

The rules for non-durable queue subscribers apply to durable subscribers.

#### Leaving the Group

As for non-durable queue subscribers, if a member's connection is closed, or if
`Unsubscribe` its called, the member leaves the group. Any unacknowledged message
is transferred to remaining members. See *Closing the Group* for important difference
with non-durable queue subscribers.

#### Closing the Group

The *last* member calling `Unsubscribe` will close (that is destroy) the
group. So if you want to maintain durability of the group, you should not be
calling `Unsubscribe`.

So unlike for non-durable queue subscribers, it is possible to maintain a queue group
with no member in the server. When a new member re-joins the durable queue group,
it will resume from where the group left of, actually first receiving all unacknowledged
messages that may have been left when the last member previously left.


### Wildcard Subscriptions

NATS Streaming subscriptions **do not** support wildcards.


## Advanced Usage

### Connection Status

The fact that the NATS Streaming server and clients are not directly connected poses a challenge when it comes to know if a client is still valid.
When a client disconnects, the streaming server is not notified, hence the importance of calling `Close()`. The server sends heartbeats
to the client's private inbox and if it misses a certain number of responses, it will consider the client's connection lost and remove it
from its state.

Before version `0.4.0`, the client library was not sending PINGs to the streaming server to detect connection failure. This was problematic
especially if an application was never sending data (had only subscriptions for instance). Picture the case where a client connects to a
NATS Server which has a route to a NATS Streaming server (either connecting to a standalone NATS Server or the server it embeds). If the
connection between the streaming server and the client's NATS Server is broken, the client's NATS connection would still be ok, yet, no
communication with the streaming server is possible. This is why relying on `Conn.NatsConn()` to check the status is not helpful.

Starting version `0.4.0` of this library and server `0.10.0`, the client library will now send PINGs at regular intervals (default is 5 seconds)
and will close the streaming connection after a certain number of PINGs have been sent without any response (default is 3). When that
happens, a callback - if one is registered - will be invoked to notify the user that the connection is permanently lost, and the reason
for the failure.

Here is how you would specify your own PING values and the callback:

```go

    // Send PINGs every 10 seconds, and fail after 5 PINGs without any response.
    sc, err := stan.Connect(clusterName, clientName,
        stan.Pings(10, 5),    
        stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
            log.Fatalf("Connection lost, reason: %v", reason)
        }))      
```

Note that the only way to be notified is to set the callback. If the callback is not set, PINGs are still sent and the connection
will be closed if needed, but the application won't know if it has only subscriptions.

When the connection is lost, your application would have to re-create it and all subscriptions if any.

When no NATS connection is provided to the `Connect()` call, the library creates its own NATS connection and will now
set the reconnect attempts to "infinite", which was not the case before. It should therefore be possible for the library to
always reconnect, but this does not mean that the streaming connection will not be closed, even if you set a very high
threshold for the PINGs max out value. Keep in mind that while the client is disconnected, the server is sending heartbeats to
the clients too, and when not getting any response, it will remove that client from its state. When the communication is restored,
the PINGs sent to the server will allow to detect this condition and report to the client that the connection is now closed.

Also, while a client is "disconnected" from the server, another application with connectivity to the streaming server may
connect and uses the same client ID. The server, when detecting the duplicate client ID, will try to contact the first client
to know if it should reject the connect request of the second client. Since the communication between the server and the
first client is broken, the server will not get a response and therefore will replace the first client with the second one.

Prior to client `0.4.0` and server `0.10.0`, if the communication between the first client and server were to be restored,
and the application would send messages, the server would accept those because the published messages client ID would be
valid, although the client is not. With client at `0.4.0+` and server `0.10.0+`, additional information is sent with each
message to allow the server to reject messages from a client that has been replaced by another client.

### Asynchronous Publishing

The basic publish API (`Publish(subject, payload)`) is synchronous; it does not return control to the caller until the NATS Streaming server has acknowledged receipt of the message. To accomplish this, a [NUID](https://github.com/nats-io/nuid) is generated for the message on creation, and the client library waits for a publish acknowledgement from the server with a matching NUID before it returns control to the caller, possibly with an error indicating that the operation was not successful due to some server problem or authorization error.

Advanced users may wish to process these publish acknowledgements manually to achieve higher publish throughput by not waiting on individual acknowledgements during the publish operation. An asynchronous publish API is provided for this purpose:

```go
    ackHandler := func(ackedNuid string, err error) {
        if err != nil {
            log.Printf("Warning: error publishing msg id %s: %v\n", ackedNuid, err.Error())
        } else {
            log.Printf("Received ack for msg id %s\n", ackedNuid)
        }
    }

    // can also use PublishAsyncWithReply(subj, replysubj, payload, ah)
    nuid, err := sc.PublishAsync("foo", []byte("Hello World"), ackHandler) // returns immediately
    if err != nil {
        log.Printf("Error publishing msg %s: %v\n", nuid, err.Error())
    }
```

### Message Acknowledgements and Redelivery

NATS Streaming offers At-Least-Once delivery semantics, meaning that once a message has been delivered to an eligible subscriber, if an acknowledgement is not received within the configured timeout interval, NATS Streaming will attempt redelivery of the message.
This timeout interval is specified by the subscription option `AckWait`, which defaults to 30 seconds.

By default, messages are automatically acknowledged by the NATS Streaming client library after the subscriber's message handler is invoked. However, there may be cases in which the subscribing client wishes to accelerate or defer acknowledgement of the message.
To do this, the client must set manual acknowledgement mode on the subscription, and invoke `Ack()` on the `Msg`. ex:

```go
// Subscribe with manual ack mode, and set AckWait to 60 seconds
aw, _ := time.ParseDuration("60s")
sub, err := sc.Subscribe("foo", func(m *stan.Msg) {
  m.Ack() // ack message before performing I/O intensive operation
  ///...
  fmt.Printf("Received a message: %s\n", string(m.Data))
}, stan.SetManualAckMode(), stan.AckWait(aw))
```

## Rate limiting/matching

A classic problem of publish-subscribe messaging is matching the rate of message producers with the rate of message consumers.
Message producers can often outpace the speed of the subscribers that are consuming their messages.
This mismatch is commonly called a "fast producer/slow consumer" problem, and may result in dramatic resource utilization spikes in the underlying messaging system as it tries to buffer messages until the slow consumer(s) can catch up.

### Publisher rate limiting

NATS Streaming provides a connection option called `MaxPubAcksInflight` that effectively limits the number of unacknowledged messages that a publisher may have in-flight at any given time. When this maximum is reached, further `PublishAsync()` calls will block until the number of unacknowledged messages falls below the specified limit. ex:

```go
sc, _ := stan.Connect(clusterID, clientID, MaxPubAcksInflight(25))

ah := func(nuid string, err error) {
    // process the ack
    ...
}

for i := 1; i < 1000; i++ {
    // If the server is unable to keep up with the publisher, the number of outstanding acks will eventually
    // reach the max and this call will block
    guid, _ := sc.PublishAsync("foo", []byte("Hello World"), ah)
}
```

### Subscriber rate limiting

Rate limiting may also be accomplished on the subscriber side, on a per-subscription basis, using a subscription option called `MaxInflight`.
This option specifies the maximum number of outstanding acknowledgements (messages that have been delivered but not acknowledged) that NATS Streaming will allow for a given subscription.
When this limit is reached, NATS Streaming will suspend delivery of messages to this subscription until the number of unacknowledged messages falls below the specified limit. ex:

```go
// Subscribe with manual ack mode and a max in-flight limit of 25
sc.Subscribe("foo", func(m *stan.Msg) {
  fmt.Printf("Received message #: %s\n", string(m.Data))
  ...
  // Does not ack, or takes a very long time to ack
  ...
  // Message delivery will suspend when the number of unacknowledged messages reaches 25
}, stan.SetManualAckMode(), stan.MaxInflight(25))

```

## License

Unless otherwise noted, the NATS source files are distributed
under the Apache Version 2.0 license found in the LICENSE file.