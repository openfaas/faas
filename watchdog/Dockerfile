FROM golang:1.10 as build

ARG VERSION
ARG GIT_COMMIT

RUN mkdir -p /go/src/github.com/openfaas/faas/watchdog
WORKDIR /go/src/github.com/openfaas/faas/watchdog

COPY vendor                     vendor
COPY metrics                    metrics
COPY types                      types
COPY main.go                    .
COPY handler.go			        .
COPY readconfig.go      	    .
COPY readconfig_test.go 	    .
COPY requesthandler_test.go 	.
COPY version.go                 .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))"

RUN go test -v ./...
# Stripping via -ldflags "-s -w" 
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w \
        -X main.GitCommit=$GIT_COMMIT \
        -X main.Version=$VERSION" \
        -installsuffix cgo -o watchdog . \
    && GOARM=7 GOARCH=arm CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w \ 
        -X main.GitCommit=$GIT_COMMIT \
        -X main.Version=$VERSION" \
        -installsuffix cgo -o watchdog-armhf . \
    && GOARCH=arm64 CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w \ 
        -X main.GitCommit=$GIT_COMMIT \
        -X main.Version=$VERSION" \ 
        -installsuffix cgo -o watchdog-arm64 . \
    && GOOS=windows CGO_ENABLED=0 go build -a -ldflags "-s -w \
        -X main.GitCommit=$GIT_COMMIT \
        -X main.Version=$VERSION" \ 
        -installsuffix cgo -o watchdog.exe .
