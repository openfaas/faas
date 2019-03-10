package handlers

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const crlf = "\r\n"
const upstreamLogsEndpoint = "/system/logs"

// NewLogHandlerFunc creates and http HandlerFunc from the supplied log Requestor.
func NewLogHandlerFunc(logProvider url.URL) http.HandlerFunc {

	writeRequestURI := false
	if _, exists := os.LookupEnv("write_request_uri"); exists {
		writeRequestURI = exists
	}

	upstreamLogProviderBase := strings.TrimSuffix(logProvider.String(), "/")

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if r.Body != nil {
			defer r.Body.Close()
		}

		logRequest := buildUpstreamRequest(r, upstreamLogProviderBase, upstreamLogsEndpoint)
		if logRequest.Body != nil {
			defer logRequest.Body.Close()
		}

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			log.Println("LogProxy: response is not a Hijacker, required for streaming response")
			http.NotFound(w, r)
			return
		}

		clientConn, buf, err := hijacker.Hijack()
		if err != nil {
			log.Println("LogProxy: failed to hijack connection for streaming response")
			return
		}

		defer clientConn.Close()
		clientConn.SetWriteDeadline(time.Time{}) // allow arbitrary time between log writes
		buf.Flush()                              // will write the headers and the initial 200 response

		if writeRequestURI {
			log.Printf("LogProxy: proxying request to %s %s\n", logRequest.Host, logRequest.URL.String())
		}

		ctx, cancel := context.WithCancel(ctx)
		logRequest = logRequest.WithContext(ctx)
		defer cancel()

		logResp, err := http.DefaultTransport.RoundTrip(logRequest)
		if err != nil {
			log.Printf("LogProxy: forwarding request failed: %s\n", err.Error())
			buf.WriteString("HTTP/1.1 500 Server Error" + crlf)
			buf.Flush()
			return
		}
		defer logResp.Body.Close()

		// write response headers directly to the buffer per RFC 2616
		// https://www.w3.org/Protocols/rfc2616/rfc2616-sec6.html
		buf.WriteString("HTTP/1.1 " + logResp.Status + crlf)
		logResp.Header.Write(buf)
		buf.WriteString(crlf)
		buf.Flush()

		// watch for connection closures and stream data
		// connections and contexts should have cancel methods deferred already
		select {
		case err := <-copyNotify(buf, logResp.Body):
			if err != nil {
				log.Printf("LogProxy: error while copy: %s", err.Error())
				return
			}
			logResp.Trailer.Write(buf)
		case err := <-closeNotify(ctx, clientConn):
			if err != nil {
				log.Printf("LogProxy: client connection closed: %s", err.Error())
			}
			return
		}

		return
	}
}

func copyNotify(destination io.Writer, source io.Reader) <-chan error {
	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(destination, source)
		done <- err
	}()
	return done
}

// closeNotify will watch the connection and notify when then connection is closed
func closeNotify(ctx context.Context, c net.Conn) <-chan error {
	notify := make(chan error, 1)

	go func() {
		buf := make([]byte, 1)
		// blocks until non-zero read or error.  From the fd.Read docs:
		// If the caller wanted a zero byte read, return immediately
		// without trying (but after acquiring the readLock).
		// Otherwise syscall.Read returns 0, nil which looks like
		// io.EOF.
		// It is important that `buf` is allocated a non-zero size
		n, err := c.Read(buf)
		if err != nil {
			log.Printf("LogProxy: test connection: %s\n", err)
			notify <- err
			return
		}
		if n > 0 {
			log.Printf("LogProxy: unexpected data: %s\n", buf[:n])
			return
		}
	}()
	return notify
}
