# Testing

## TL;DR

```shell
make test                            # full race-enabled suite, no docker
make test T=TestName PKG=./test/...  # iterate on a single test
```

## How the suite is set up

The repo uses two module files: `go.mod` for production (minimal dependencies)
and `go_test.mod` for testing. Always pass `-modfile=go_test.mod` to `go test`;
the make targets below do this for you.

Unit tests live at the repo root (package `nats`, white-box) and run with plain
`go test`. Integration tests live in `./test`, `./jetstream/test`, and
`./micro/test`; they bring up real NATS servers by talking to a **tester**
service, which spawns `nats-server` instances, clusters, and super-clusters on
demand. Tests drive it through the `github.com/synadia-io/orbit.go/ntf-client`
package.

There are two ways to run that tester, and tests reach both the same way:

- **In-process (the default)** — a `TestMain` starts
  `github.com/synadia-io/orbit.go/ntf` inside each test binary. No docker, no
  setup, nothing to tear down.
- **Out-of-process** — the `synadia/ntf-server` docker image, located via the
  `TESTER_NATS_URL` environment variable. This is what CI uses.

`TESTER_NATS_URL` always wins: set it and the tests talk to that tester instead
of starting their own. The `make test` targets clear it so an exported value
left over in your shell cannot silently redirect them, but a raw `go test`
inherits it — unset it if a plain `go test` tries to reach a container you did
not intend.

The two paths do **not** necessarily run the same server build: in-process links
whatever `nats-server` `orbit.go/ntf` depends on, while the container ships its
own (pinned in `.github/workflows/ci.yaml`). To stop CI drifting onto the wrong
one, `TestMain` refuses to start the in-process tester when `GITHUB_ACTIONS` is
set but `TESTER_NATS_URL` is not. `GITHUB_ACTIONS` rather than the more general
`CI` because devcontainers and other runners set that one too, and they have no
reason to be forced onto a container.

That guard only catches a missing URL, not a version difference, so the first
test to reach the tester prints the server it got:

```text
tester at nats://localhost:62354: nats-server 2.14.5
```

`go test` shows this whenever the package is run with `-v` (all CI steps are) or
when a test in it fails. If a failure reproduces locally but not in CI, or the
reverse, compare that line against the image tag in `ci.yaml` first.

## In-process mode (default, no docker)

```shell
make test                                  # everything
make test T=TestSubSubject PKG=./test/...  # one test, verbose
make test PKG=./jetstream/test/...         # one package
make test-norace                           # NoRace tests (race detector off)
```

Equivalent to:

```shell
go test -modfile=go_test.mod -tags=internal_testing -race -p=1 ./... --failfast -vet=off
```

Because the tester and the `nats-server` instances it spawns run inside the test
binary, `-race` instruments the server too: expect somewhat slower runs, and note
that a data race in `nats-server` surfaces here as a nats.go test failure.

This is also why `go_test.mod` declares `go 1.26.0` while `go.mod` stays at
`1.25.0` — `orbit.go/ntf` requires 1.26, and merely requiring a module raises the
floor for the whole test module. Production builds are unaffected; 1.26 is the
floor for the CI test matrix.

## Testing against a different nats-server

The in-process tester links `nats-server` as an ordinary Go dependency, so
pointing it at another branch, tag, commit, or local checkout is a module
operation — no docker, no image build:

```shell
make server-replace V=main                        # nats-server main
make server-replace V=v2.13.0                     # a specific tag
make server-replace V=/path/to/your/nats-server   # a local checkout
make test                                         # run against it
make server-replace-drop                          # undo
```

Verify what you are actually running:

```shell
go list -m -modfile=go_test.mod github.com/nats-io/nats-server/v2
```

**Do not use `go get` for this.** A branch pseudo-version is derived from the
last reachable tag, so `@main` can sort *below* the release `orbit.go/ntf`
requires. `go get` resolves that conflict by removing ntf from the module graph,
which silently leaves you with no in-process tester:

```text
go: downgraded github.com/nats-io/nats-server/v2 => v2.14.1-0.2026...
go: removed github.com/synadia-io/orbit.go/ntf
```

A `replace` directive overrides the version outright and has no such failure
mode, which is what `make server-replace` uses.

`go_test.mod` is a tracked file, so the replace must not be committed. Note that
`make server-replace-drop` removes the directive but does **not** fully restore
the file: while the replacement was active, `go mod tidy` recorded the replaced
module's own newer dependencies, and MVS never downgrades them again. The target
warns when this happens; for an exact restore use
`git checkout -- go_test.mod go_test.sum`.

The docker path can also test a different server, but only versions Synadia
publishes as images — this is what the nightly `latest-server.yaml` workflow
uses:

```shell
make tester-up-host TESTER_IMAGE=synadia/ntf-server:nightly-latest-nats-main
make test-docker
```

## Host-side mode (docker, iterating on individual tests)

`make tester-up-host` starts the tester with its ports published, so `go test`
run from your terminal can reach the spawned servers via localhost.

```shell
make tester-up-host
make test-docker T=TestSubSubject PKG=./test/...  # one test, verbose
make test-docker PKG=./jetstream/test/...         # one package
make test-docker                                  # everything
```

Use the `test-docker` targets, not `test` — plain `make test` ignores the
container and uses the in-process tester.

`make test-docker` wraps the full invocation, which is equivalent to:

```shell
TESTER_NATS_URL=nats://localhost:4222 \
  go test -modfile=go_test.mod -tags=internal_testing -race -p=1 ./... --failfast -vet=off
```

`-p=1` is required here: a shared container tester does not tolerate concurrent
CreateServer calls from independent test binaries. The in-process tester is
per-binary and could run packages in parallel — measurably faster — but `-p=1`
is kept deliberately, since more concurrent server churn risks intermittent
connection failures and the wall-clock saving is not worth chasing flakes.

Known caveat: on macOS, docker-proxy races the tester's port handover, so in
heavy suites roughly 5-10% of server creations can fail with
`bind: address already in use`. Rerun the failing test, or use
sibling-container mode for full-suite runs.

## Sibling-container mode (full suite, matches CI)

```shell
make tester-up
make test-tester
make tester-down
```

`make test-tester` runs the whole suite (NoRace pass, then the race-enabled
pass) inside a Go container on the same docker network as the tester — no
published ports, so the docker-proxy race above does not apply. This is the
same shape CI uses, with the tester attached as a service container
(`.github/workflows/ci.yaml`).

## NoRace tests

Tests prefixed `TestNoRace` are guarded by `//go:build !race &&
!skip_no_race_tests` and must run with the race detector off:
`make test-norace`.

## Build tags

- `internal_testing` — exposes internal test hooks from `testing_internal.go`;
  required by some tests in `./test`. The make targets set it.
- `skip_no_race_tests` — excludes the NoRace tests from a non-race run (used
  by `scripts/cov.sh`).
- `compat` — compatibility tests in `test/compat_test.go`, which connect to an
  external NATS server via `NATS_URL`.
- `go1.23` — iterator-based tests in `test/nats_iter_test.go`.

## Coverage

```shell
TESTER_NATS_URL=nats://localhost:4222 ./scripts/cov.sh
```

Merges unit and integration coverage into `acc.out` and opens the HTML report
(CI passes an argument to skip the browser).

## Troubleshooting

- `cannot reach the tester at ...` — `TESTER_NATS_URL` is set but nothing is
  answering there. Either start the container (`make tester-up-host`) or unset
  the variable to use the in-process tester.
- `TESTER_NATS_URL must be set in CI` — running under GitHub Actions without
  `TESTER_NATS_URL`. This guard exists so CI cannot silently swap the pinned
  container server for the in-process one, which is a different version. Point
  `TESTER_NATS_URL` at a tester, or unset `GITHUB_ACTIONS`.
- Tester container misbehaving — `docker logs -f nats-tester` shows server spawn
  and config errors; `docker restart nats-tester` restarts it while keeping
  those logs. `make tester-down` removes the container and discards them.
- Updating test-only dependencies: `go mod tidy -modfile=go_test.mod` (never
  change the main `go.mod` for test dependencies).
