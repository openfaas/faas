// Package logs provides the standard interface and handler for OpenFaaS providers to expose function logs.
//
// The package defines the Requester interface that OpenFaaS providers should implement and then expose using
// the predefined NewLogHandlerFunc. See the example folder for a minimal log provider implementation.
//
// The Requester is where the actual specific logic for connecting to and querying the log system should be implemented.
//
package logs

import (
	"fmt"
	"time"
)

// Request is the query to return the function logs.
type Request struct {
	// Name is the function name and is required
	Name string `json:"name"`
	// Instance is the optional container name, that allows you to request logs from a specific function instance
	Instance string `json:"instance"`
	// Since is the optional datetime value to start the logs from
	Since *time.Time `json:"since"`
	// Tail sets the maximum number of log messages to return, <=0 means unlimited
	Tail int `json:"tail"`
	// Follow is allows the user to request a stream of logs until the timeout
	Follow bool `json:"follow"`
}

// String implements that Stringer interface and prints the log Request in a consistent way that
// allows you to safely compare if two requests have the same value.
func (r Request) String() string {
	return fmt.Sprintf("name:%s instance:%s since:%v tail:%d follow:%v", r.Name, r.Instance, r.Since, r.Tail, r.Follow)
}

// Message is a specific log message from a function container log stream
type Message struct {
	// Name is the function name
	Name string `json:"name"`
	// instance is the name/id of the specific function instance
	Instance string `json:"instance"`
	// Timestamp is the timestamp of when the log message was recorded
	Timestamp time.Time `json:"timestamp"`
	// Text is the raw log message content
	Text string `json:"text"`
}

// String implements the Stringer interface and allows for nice and simple string formatting of a log Message.
func (m Message) String() string {
	return fmt.Sprintf("%s %s (%s) %s", m.Timestamp.String(), m.Name, m.Instance, m.Text)
}
