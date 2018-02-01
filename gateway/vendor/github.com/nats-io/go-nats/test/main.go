// Copyright 2017 Apcera Inc. All rights reserved.

package test

import (
	"os"
	"testing"
)

// TestMain runs all tests. Added since tests were moved to a separate package.
func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
