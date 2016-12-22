from golang:1.7.3

RUN go get -d github.com/docker/docker/api/types \
 && go get -d github.com/docker/docker/api/types/filters \
  && go get -d github.com/docker/docker/api/types/swarm \
   && go get -d github.com/docker/docker/client
RUN go get github.com/gorilla/mux \
    && go get github.com/prometheus/client_golang/prometheus

WORKDIR /go/src/github.com/alexellis/faas/gateway

COPY metrics metrics
COPY server.go .

RUN find /go/src/github.com/alexellis/faas/gateway/
RUN go build
EXPOSE 8080

CMD ["./gateway"]
