FROM golang:1.7.5
RUN mkdir -p /go/src/github.com/alexellis/faas/watchdog
WORKDIR /go/src/github.com/alexellis/faas/watchdog

COPY main.go        .
COPY readconfig.go  .
COPY config_test.go .
COPY requesthandler_test.go .
#COPY fastForkRequestHandler.go  .
#COPY requestHandler.go   .
COPY types types

RUN go test

# Stripping via -ldflags "-s -w" 
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w" -installsuffix cgo -o watchdog . \
    && GOARM=6 GOARCH=arm CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w" -installsuffix cgo -o watchdog-armhf . \
    && GOOS=windows CGO_ENABLED=0 go build -a -ldflags "-s -w" -installsuffix cgo -o watchdog.exe .
