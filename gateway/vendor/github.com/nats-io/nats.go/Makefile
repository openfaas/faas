# Local convenience targets. The default path needs no docker: a TestMain runs
# the ntf tester inside each test binary.
#
#   make test                            # full suite
#   make test T=TestName PKG=./test/...  # one test
#
# The test-docker targets run against an external tester instead. Docker modes
# and the server-replace workflow: see TESTING.md.

# Tester image is pinned in ci.yaml (single source of truth); parsed from there.
TESTER_IMAGE   ?= $(shell sed -n 's|^ *image: *\(synadia/ntf-server:[^ ]*\).*|\1|p' .github/workflows/ci.yaml)
ifeq ($(strip $(TESTER_IMAGE)),)
$(error could not parse the tester image from .github/workflows/ci.yaml)
endif
TESTER_NAME    ?= nats-tester
TESTER_NETWORK ?= nats-tester-net
GO_IMAGE       ?= golang:1.26-alpine

# T limits the run to a single test (-run, verbose); PKG limits the packages.
TESTER_NATS_URL ?= nats://localhost:4222
PKG ?= ./...

.PHONY: tester-net tester-up tester-up-host tester-down test-tester test test-norace test-docker server-replace server-replace-drop

# Points the in-process tester at another nats-server via a replace directive in
# go_test.mod. V is a version/branch/commit, or a path when it starts with / or .
# Never commit the directive. Prefer this over `go get`, which can drop ntf from
# the module graph entirely — see TESTING.md.
#   make server-replace V=main | v2.13.0 | /path/to/nats-server
server-replace:
	@test -n "$(V)" || { echo "usage: make server-replace V=main|v2.13.0|/path/to/nats-server"; exit 1; }
	go mod edit -modfile=go_test.mod -replace=github.com/nats-io/nats-server/v2=$(if $(filter /% .%,$(V)),$(V),github.com/nats-io/nats-server/v2@$(V))
	GOFLAGS=-mod=mod go mod tidy -modfile=go_test.mod
	@echo "now building against:"
	@go list -m -modfile=go_test.mod github.com/nats-io/nats-server/v2

# Removes the directive. Does not fully restore go_test.mod; see the warning it
# prints below.
server-replace-drop:
	go mod edit -modfile=go_test.mod -dropreplace=github.com/nats-io/nats-server/v2
	GOFLAGS=-mod=mod go mod tidy -modfile=go_test.mod
	@echo "now building against:"
	@go list -m -modfile=go_test.mod github.com/nats-io/nats-server/v2
	@git diff --quiet -- go_test.mod go_test.sum || { \
		echo ""; \
		echo "WARNING: go_test.mod/go_test.sum still differ from HEAD — the replaced"; \
		echo "module's dependencies stayed upgraded. Do not commit these. For an exact"; \
		echo "restore (discards other local edits to both files):"; \
		echo "    git checkout -- go_test.mod go_test.sum"; \
	}

# -count=1 everywhere: results depend on external state (which tester answers)
# that Go's cache key misses, so a hit can replay an in-process result for a
# test-docker run. -p=1: the shared container tester does not tolerate
# concurrent CreateServer calls from independent test binaries.

test:
	TESTER_NATS_URL= go test -modfile=go_test.mod -tags=internal_testing -race -p=1 -count=1 $(if $(T),-v -run '$(T)') $(PKG) --failfast -vet=off

# test-docker is test against the container started by `make tester-up-host`.
test-docker:
	TESTER_NATS_URL=$(TESTER_NATS_URL) go test -modfile=go_test.mod -tags=internal_testing -race -p=1 -count=1 $(if $(T),-v -run '$(T)') $(PKG) --failfast -vet=off

# TestNoRace* tests must run with the race detector off.
test-norace:
	TESTER_NATS_URL= go test -modfile=go_test.mod -p=1 -count=1 $(if $(T),-v) -run '$(or $(T),TestNoRace)' $(PKG) --failfast -vet=off

tester-net:
	@docker network inspect $(TESTER_NETWORK) >/dev/null 2>&1 || \
		docker network create $(TESTER_NETWORK)

# No host port publishing, for sibling-container mode (make test-tester).
# Publishing here would break server bring-up: docker-proxy holds 0.0.0.0:<port>
# inside the container's netns, racing the tester's port-reservation handover
# ("bind: address already in use"). The ephemeral range is left at the kernel
# default so TIME_WAIT churn cannot crowd it; narrowing it only matters when
# ports must match a published range (see tester-up-host).
tester-up: tester-net
	docker run -d \
		--name $(TESTER_NAME) \
		--network $(TESTER_NETWORK) \
		--restart unless-stopped \
		-e NATS_ADVERTISE=$(TESTER_NAME) \
		$(TESTER_IMAGE) serve
	@echo "Tester up as '$(TESTER_NAME)'. Run the suite: make test-tester"

# Publishes ports so `go test` on the host reaches the spawned servers;
# ip_local_port_range is narrowed to match the published range. On macOS
# docker-proxy races the port handover, failing ~5-10% of server creations in
# heavy suites — rerun, or use sibling-container mode, which is immune.
tester-up-host: tester-net
	docker run -d \
		--name $(TESTER_NAME) \
		--network $(TESTER_NETWORK) \
		--restart unless-stopped \
		--sysctl net.ipv4.ip_local_port_range="30000 31000" \
		-p 4222:4222 \
		-p 30000-31000:30000-31000 \
		-e NATS_ADVERTISE=localhost \
		$(TESTER_IMAGE) serve
	@echo "Tester up as '$(TESTER_NAME)'. Run: make test-docker [T=...] [PKG=...]"
	@echo "(plain 'make test' ignores this container and uses the in-process tester)"

# Removes the container, discarding its logs. To keep it for debugging instead:
# docker logs -f $(TESTER_NAME) / docker restart $(TESTER_NAME).
tester-down:
	-docker rm -f $(TESTER_NAME)
	-docker network rm $(TESTER_NETWORK)

# Same two passes as the CI `test` job: NoRace first, then the full race suite.
test-tester: tester-net
	docker run --rm \
		--network $(TESTER_NETWORK) \
		-v $(CURDIR):/src \
		-w /src \
		-e TESTER_NATS_URL=nats://$(TESTER_NAME):4222 \
		-e CGO_ENABLED=1 \
		$(GO_IMAGE) sh -c '\
			apk add --no-cache gcc libc-dev git make >/dev/null && \
			go test -modfile=go_test.mod -v -run=TestNoRace -p=1 -count=1 ./... --failfast -vet=off && \
			go test -modfile=go_test.mod -tags=internal_testing -race -v -p=1 -count=1 ./... --failfast -vet=off'
