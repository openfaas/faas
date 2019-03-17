# goleak [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov]

Goroutine leak detector to help avoid Goroutine leaks.

## Development Status: Alpha

goleak is still in development, and APIs are still in flux.

## Installation

You can use `go get` to get the latest version:

`go get -u go.uber.org/goleak`

`goleak` also supports semver releases. It is compatible with Go 1.5+.

## Quick Start

To verify that there are no unexpected goroutines running at the end of a test:

```go
func TestA(t *testing.T) {
	defer goleak.Verify(t)

	// test logic here.
}
```

Instead of checking for leaks at the end of every test, `goleak` can also be run
at the end of every test package by creating a `TestMain` function for your 
package:

```go
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
```


[doc-img]: https://godoc.org/go.uber.org/goleak?status.svg
[doc]: https://godoc.org/go.uber.org/goleak
[ci-img]: https://travis-ci.org/uber-go/goleak.svg?branch=master
[ci]: https://travis-ci.org/uber-go/goleak
[cov-img]: https://codecov.io/gh/uber-go/goleak/branch/master/graph/badge.svg
[cov]: https://codecov.io/gh/uber-go/goleak
[benchmarking suite]: https://github.com/uber-go/goleak/tree/master/benchmarks
[glide.lock]: https://github.com/uber-go/goleak/blob/master/glide.lock