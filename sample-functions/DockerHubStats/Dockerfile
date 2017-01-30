FROM aafrey/golang-alpine:arm
MAINTAINER alexellis2@gmail.com
ENTRYPOINT []

RUN apk --no-cache add make
WORKDIR /go/src/github.com/alexellis/faas/sample-functions/DockerHubStats
COPY . /go/src/github.com/alexellis/faas/sample-functions/DockerHubStats

RUN make
COPY fwatchdog /usr/bin/fwatchdog
ENV fprocess "/go/bin/DockerHubStats"

CMD [ "/usr/bin/fwatchdog"]
