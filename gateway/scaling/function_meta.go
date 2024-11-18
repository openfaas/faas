// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

// Copyright (c) OpenFaaS Author(s). All rights reserved.

package scaling

import (
	"time"
)

// FunctionMeta holds the last refresh and any other
// meta-data needed for caching.
type FunctionMeta struct {
	LastRefresh          time.Time
	ServiceQueryResponse ServiceQueryResponse
}

// Expired find out whether the cache item has expired with
// the given expiry duration from when it was stored.
func (fm *FunctionMeta) Expired(expiry time.Duration) bool {
	return time.Now().After(fm.LastRefresh.Add(expiry))
}
