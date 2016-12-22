from golang:1.7.3

RUN go get -d github.com/docker/docker/api/types \
 && go get -d github.com/docker/docker/api/types/filters \
  && go get -d github.com/docker/docker/api/types/swarm \
   && go get -d github.com/docker/docker/client

WORKDIR /go/src/app/
RUN go get github.com/gorilla/mux
COPY server.go .
RUN go build
EXPOSE 8080

CMD ["./app"]

