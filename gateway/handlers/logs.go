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

		cn, ok := w.(http.CloseNotifier)
		if !ok {
			log.Println("LogHandler: response is not a CloseNotifier, required for streaming response")
			http.NotFound(w, r)
			return
		}

		wf, ok := w.(writerFlusher)
		if !ok {
			log.Println("LogHandler: response is not a Flusher, required for streaming response")
			http.NotFound(w, r)
			return
		}

		if writeRequestURI {
			log.Printf("LogProxy: proxying request to %s %s\n", logRequest.Host, logRequest.URL.String())
		}

		ctx, cancel := context.WithCancel(ctx)
		logRequest = logRequest.WithContext(ctx)
		defer cancel()

		logResp, err := http.DefaultTransport.RoundTrip(logRequest)
		if err != nil {
			log.Printf("LogProxy: forwarding request failed: %s\n", err.Error())
			http.Error(w, "log request failed", http.StatusInternalServerError)
			return
		}
		defer logResp.Body.Close()

		// watch for connection closures and stream data
		// connections and contexts should have cancel methods deferred already
		select {
		case err := <-copyNotify(&unbufferedWriter{wf}, logResp.Body):
			if err != nil {
				log.Printf("LogProxy: error while copy: %s", err.Error())
				return
			}
		case <-cn.CloseNotify():
			log.Printf("LogProxy: client connection closed")
			return
		}

		return
	}
}

// NewHijackLogHandlerFunc creates and http HandlerFunc from the supplied log Requestor.
// This remains in the package for reference purposes, providers should instead use NewLogHandlerFunc
func NewHijackLogHandlerFunc(logProvider url.URL) http.HandlerFunc {

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
		// A zero value for t means Write will not time out, allowing us to stream the logs while
		// following even if there is a large gap in logs
		clientConn.SetWriteDeadline(time.Time{})
		// flush the headers and the initial 200 status code to the client
		buf.Flush()

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
		// Body is always closeable if err is nil
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

type writerFlusher interface {
	io.Writer
	http.Flusher
}

// unbufferedWriter is an io Writer that immediately flushes the after every call to Write.
// This can wrap any http.ResponseWriter that also implements Flusher.  This ensures that log
// lines are immediately sent to the client
type unbufferedWriter struct {
	dst writerFlusher
}

// Write writes to the dst writer and then immediately flushes the writer
func (u *unbufferedWriter) Write(p []byte) (n int, err error) {
	n, err = u.dst.Write(p)
	u.dst.Flush()

	return n, err
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
