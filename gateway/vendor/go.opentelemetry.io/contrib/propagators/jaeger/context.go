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

package jaeger

import "context"

type jaegerKeyType int

const (
	debugKey jaegerKeyType = iota
)

// withDebug returns a copy of parent with debug set as the debug flag value .
func withDebug(parent context.Context, debug bool) context.Context {
	return context.WithValue(parent, debugKey, debug)
}

// debugFromContext returns the debug value stored in ctx.
//
// If no debug value is stored in ctx false is returned.
func debugFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if debug, ok := ctx.Value(debugKey).(bool); ok {
		return debug
	}
	return false
}
