FROM golang:1.8.3
RUN mkdir -p /go/src/github.com/alexellis/faas/watchdog
WORKDIR /go/src/github.com/alexellis/faas/watchdog

COPY main.go        .
COPY readconfig.go  .
COPY config_test.go .
COPY requesthandler_test.go .
#COPY fastForkRequestHandler.go  .
#COPY requestHandler.go   .
COPY types types

# Run a gofmt and exclude all vendored code.
RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*")

RUN go test

# Stripping via -ldflags "-s -w" 
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w" -installsuffix cgo -o watchdog . \
    && GOARM=6 GOARCH=arm CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w" -installsuffix cgo -o watchdog-armhf . \
    && GOARCH=arm64 CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w" -installsuffix cgo -o watchdog-arm64 . \
    && GOOS=windows CGO_ENABLED=0 go build -a -ldflags "-s -w" -installsuffix cgo -o watchdog.exe .
