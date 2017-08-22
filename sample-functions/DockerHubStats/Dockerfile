FROM golang:1.7.5-alpine
MAINTAINER alexellis2@gmail.com
ENTRYPOINT []

RUN apk --no-cache add make
WORKDIR /go/src/github.com/alexellis/faas/sample-functions/DockerHubStats
COPY . /go/src/github.com/alexellis/faas/sample-functions/DockerHubStats

RUN make

ADD https://github.com/alexellis/faas/releases/download/0.6.0/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess "/go/bin/DockerHubStats"

CMD [ "/usr/bin/fwatchdog"]
