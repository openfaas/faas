FROM golang:1.7.5
RUN mkdir -p /go/src/app
WORKDIR /go/src/app
COPY main.go        .
COPY readconfig.go  .
COPY config_test.go .

WORKDIR /go/src/app
RUN go get -d -v

RUN go test
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

