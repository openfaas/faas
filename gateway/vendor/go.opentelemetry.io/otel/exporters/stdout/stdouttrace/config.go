// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stdouttrace // import "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

import (
	"io"
	"os"
)

var (
	defaultWriter      = os.Stdout
	defaultPrettyPrint = false
	defaultTimestamps  = true
)

// config contains options for the STDOUT exporter.
type config struct {
	// Writer is the destination.  If not set, os.Stdout is used.
	Writer io.Writer

	// PrettyPrint will encode the output into readable JSON. Default is
	// false.
	PrettyPrint bool

	// Timestamps specifies if timestamps should be printed. Default is
	// true.
	Timestamps bool
}

// newConfig creates a validated Config configured with options.
func newConfig(options ...Option) (config, error) {
	cfg := config{
		Writer:      defaultWriter,
		PrettyPrint: defaultPrettyPrint,
		Timestamps:  defaultTimestamps,
	}
	for _, opt := range options {
		opt.apply(&cfg)

	}
	return cfg, nil
}

// Option sets the value of an option for a Config.
type Option interface {
	apply(*config)
}

// WithWriter sets the export stream destination.
func WithWriter(w io.Writer) Option {
	return writerOption{w}
}

type writerOption struct {
	W io.Writer
}

func (o writerOption) apply(cfg *config) {
	cfg.Writer = o.W
}

// WithPrettyPrint sets the export stream format to use JSON.
func WithPrettyPrint() Option {
	return prettyPrintOption(true)
}

type prettyPrintOption bool

func (o prettyPrintOption) apply(cfg *config) {
	cfg.PrettyPrint = bool(o)
}

// WithoutTimestamps sets the export stream to not include timestamps.
func WithoutTimestamps() Option {
	return timestampsOption(false)
}

type timestampsOption bool

func (o timestampsOption) apply(cfg *config) {
	cfg.Timestamps = bool(o)
}
