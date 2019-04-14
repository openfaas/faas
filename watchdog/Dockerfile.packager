FROM openfaas/watchdog:build as build
FROM scratch

ARG PLATFORM

COPY --from=build /go/src/github.com/openfaas/faas/watchdog/watchdog$PLATFORM ./fwatchdog