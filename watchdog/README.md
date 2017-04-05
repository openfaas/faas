Watchdog
==========

The FaaS watchdog is designed to marshal a HTTP request between your public HTTP URL and a individual function.

Creating a function:

- [x] Use an existing or a new Docker image
- [x] Add the fwatchdog binary from the [Releases page](https://github.com/alexellis/faas/releases) via `curl` or `ADD https://`
- [x] Set an `fprocess` environmental variable with the function you want to run for each request
- [x] Expose port 8080
- [x] Set the `CMD` to `fwatchdog`

Example Dockerfile:

```
FROM alpine:3.5

ADD https://github.com/alexellis/faas/releases/download/v0.5-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

# Define your UNIX binary here
ENV fprocess="/bin/cat"

CMD ["fwatchdog"]
```

**Implementing the a healthcheck**

Docker swarm will keep your function out of the DNS-RR / IPVS pool if the task (container) is not healthy.

Here is an example of the `echo` function implementing a healthcheck with a 5-second checking interval.

```
FROM functions/alpine

ENV fprocess="cat /etc/hostname"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
```

The watchdog process creates a .lock file in `/tmp/` on starting its internal Golang HTTP server. `[ -e file_name ]` is shell to check if a file exists.

Swarm tutorial on Healthchecks:

 * [Test-drive Docker Healthcheck in 10 minutes](http://blog.alexellis.io/test-drive-healthcheck/)


**Environmental overrides:**

A number of environmental overrides can be added for additional flexibility and options:

| Option                 | Usage             |
|------------------------|--------------|
| `fprocess`             | The process to invoke for each function call. This must be a UNIX binary and accept input via STDIN and output via STDOUT.  |
| `marshal_requests`     | Instead of re-directing the raw HTTP body into your fprocess, it will first be marshalled into JSON. Use this if you need to work with HTTP headers |
| `write_timeout`        | HTTP timeout for writing a response body from your function  |
| `read_timeout`         | HTTP timeout for reading the payload from the client caller  |
| `suppress_lock`        | The watchdog will attempt to write a lockfile to /tmp/ for swarm healthchecks - set this to true to disable behaviour. |
