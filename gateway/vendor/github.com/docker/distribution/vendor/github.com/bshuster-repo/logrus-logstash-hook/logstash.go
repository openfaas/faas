package logrus_logstash

import (
	"net"
	"strings"

	"github.com/Sirupsen/logrus"
)

// Hook represents a connection to a Logstash instance
type Hook struct {
	conn             net.Conn
	appName          string
	alwaysSentFields logrus.Fields
	hookOnlyPrefix   string
}

// NewHook creates a new hook to a Logstash instance, which listens on
// `protocol`://`address`.
func NewHook(protocol, address, appName string) (*Hook, error) {
	return NewHookWithFields(protocol, address, appName, make(logrus.Fields))
}

// NewHookWithConn creates a new hook to a Logstash instance, using the supplied connection
func NewHookWithConn(conn net.Conn, appName string) (*Hook, error) {
	return NewHookWithFieldsAndConn(conn, appName, make(logrus.Fields))
}

// NewHookWithFields creates a new hook to a Logstash instance, which listens on
// `protocol`://`address`. alwaysSentFields will be sent with every log entry.
func NewHookWithFields(protocol, address, appName string, alwaysSentFields logrus.Fields) (*Hook, error) {
	return NewHookWithFieldsAndPrefix(protocol, address, appName, alwaysSentFields, "")
}

// NewHookWithFieldsAndPrefix creates a new hook to a Logstash instance, which listens on
// `protocol`://`address`. alwaysSentFields will be sent with every log entry. prefix is used to select fields to filter
func NewHookWithFieldsAndPrefix(protocol, address, appName string, alwaysSentFields logrus.Fields, prefix string) (*Hook, error) {
	conn, err := net.Dial(protocol, address)
	if err != nil {
		return nil, err
	}
	return NewHookWithFieldsAndConnAndPrefix(conn, appName, alwaysSentFields, prefix)
}

// NewHookWithFieldsAndConn creates a new hook to a Logstash instance using the supplied connection
func NewHookWithFieldsAndConn(conn net.Conn, appName string, alwaysSentFields logrus.Fields) (*Hook, error) {
	return NewHookWithFieldsAndConnAndPrefix(conn, appName, alwaysSentFields, "")
}

//NewHookWithFieldsAndConnAndPrefix creates a new hook to a Logstash instance using the suppolied connection and prefix
func NewHookWithFieldsAndConnAndPrefix(conn net.Conn, appName string, alwaysSentFields logrus.Fields, prefix string) (*Hook, error) {
	return &Hook{conn: conn, appName: appName, alwaysSentFields: alwaysSentFields, hookOnlyPrefix: prefix}, nil
}

//NewFilterHook makes a new hook which does not forward to logstash, but simply enforces the prefix rules
func NewFilterHook() *Hook {
	return NewFilterHookWithPrefix("")
}

//NewFilterHookWithPrefix make a new hook which does not forward to logstash, but simply enforces the specified prefix
func NewFilterHookWithPrefix(prefix string) *Hook {
	return &Hook{conn: nil, appName: "", alwaysSentFields: make(logrus.Fields), hookOnlyPrefix: prefix}
}

func (h *Hook) filterHookOnly(entry *logrus.Entry) {
	if h.hookOnlyPrefix != "" {
		for key := range entry.Data {
			if strings.HasPrefix(key, h.hookOnlyPrefix) {
				delete(entry.Data, key)
			}
		}
	}

}

//WithPrefix sets a prefix filter to use in all subsequent logging
func (h *Hook) WithPrefix(prefix string) {
	h.hookOnlyPrefix = prefix
}

func (h *Hook) WithField(key string, value interface{}) {
	h.alwaysSentFields[key] = value
}

func (h *Hook) WithFields(fields logrus.Fields) {
	//Add all the new fields to the 'alwaysSentFields', possibly overwriting exising fields
	for key, value := range fields {
		h.alwaysSentFields[key] = value
	}
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	//make sure we always clear the hookonly fields from the entry
	defer h.filterHookOnly(entry)

	// Add in the alwaysSentFields. We don't override fields that are already set.
	for k, v := range h.alwaysSentFields {
		if _, inMap := entry.Data[k]; !inMap {
			entry.Data[k] = v
		}
	}

	//For a filteringHook, stop here
	if h.conn == nil {
		return nil
	}

	formatter := LogstashFormatter{Type: h.appName}

	dataBytes, err := formatter.FormatWithPrefix(entry, h.hookOnlyPrefix)
	if err != nil {
		return err
	}
	if _, err = h.conn.Write(dataBytes); err != nil {
		return err
	}
	return nil
}

func (h *Hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}
