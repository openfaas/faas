package logrus_logstash

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
)

// Formatter generates json in logstash format.
// Logstash site: http://logstash.net/
type LogstashFormatter struct {
	Type string // if not empty use for logstash type field.

	// TimestampFormat sets the format used for timestamps.
	TimestampFormat string
}

func (f *LogstashFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return f.FormatWithPrefix(entry, "")
}

func (f *LogstashFormatter) FormatWithPrefix(entry *logrus.Entry, prefix string) ([]byte, error) {
	fields := make(logrus.Fields)
	for k, v := range entry.Data {
		//remvove the prefix when sending the fields to logstash
		if prefix != "" && strings.HasPrefix(k, prefix) {
			k = strings.TrimPrefix(k, prefix)
		}

		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/Sirupsen/logrus/issues/377
			fields[k] = v.Error()
		default:
			fields[k] = v
		}
	}

	fields["@version"] = "1"

	timeStampFormat := f.TimestampFormat

	if timeStampFormat == "" {
		//timeStampFormat = logrus.DefaultTimestampFormat
		timeStampFormat = "2006-01-02 15:04:05.000"
	}

	fields["@timestamp"] = entry.Time.Format(timeStampFormat)

	// set message field
	v, ok := entry.Data["message"]
	if ok {
		fields["fields.message"] = v
	}
	fields["message"] = entry.Message

	// set level field
	v, ok = entry.Data["level"]
	if ok {
		fields["fields.level"] = v
	}
	fields["level"] = entry.Level.String()

	// set type field
	if f.Type != "" {
		v, ok = entry.Data["type"]
		if ok {
			fields["fields.type"] = v
		}
		fields["type"] = f.Type
	}

	serialized, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}
