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

**Environmental overrides:**

A number of environmental overrides can be added for additional flexibility and options:

| Option                 | Usage             |
|------------------------|--------------|
| `fprocess`     | The process to invoke for each function call. This must be a UNIX binary and accept input via STDIN and output via STDOUT.  |
| `marshal_requests`     | Instead of re-directing the raw HTTP body into your fprocess, it will first be marshalled into JSON. Use this if you need to work with HTTP headers |
| `write_timeout`     | HTTP timeout for writing a response body from your function  |
| `read_timeout`     | HTTP timeout for reading the payload from the client caller  |
