FROM golang:1.7.3
RUN mkdir -p /go/src/app
COPY handler.go /go/src/app
WORKDIR /go/src/app
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

COPY fwatchdog  /usr/bin/

ENV fprocess="/go/src/app/app"
CMD ["fwatchdog"]
