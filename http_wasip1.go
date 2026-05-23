//go:build wasip1

package client

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"unsafe"
)

// emptyBuf provides a valid non-nil pointer for optional host function parameters
// when the caller supplies no data.
var emptyBuf [1]byte

func init() {
	t := &hostTransport{}
	http.DefaultTransport = t
	http.DefaultClient = &http.Client{Transport: t}
}

//go:wasmimport looking-glass http_stream_open
func hostHTTPStreamOpen(
	methodPtr unsafe.Pointer, methodLen uint32,
	urlPtr unsafe.Pointer, urlLen uint32,
	hdrPtr unsafe.Pointer, hdrLen uint32,
	bodyPtr unsafe.Pointer, bodyLen uint32,
) int32

//go:wasmimport looking-glass http_stream_status
func hostHTTPStreamStatus(handle int32) uint32

//go:wasmimport looking-glass http_stream_read
func hostHTTPStreamRead(handle int32, bufPtr unsafe.Pointer, bufLen uint32) int32

//go:wasmimport looking-glass http_stream_close
func hostHTTPStreamClose(handle int32)

// hostTransport implements http.RoundTripper by routing every request through
// the looking-glass host. It supports both short request/response exchanges and
// indefinitely long response streams such as Server-Sent Events.
type hostTransport struct{}

// RoundTrip implements http.RoundTripper.
func (t *hostTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	method := req.Method
	if method == "" {
		method = http.MethodGet
	}
	methodBytes := []byte(method)

	urlBytes := []byte(req.URL.String())

	// Headers are serialised as "Key: Value\n" lines for the host to parse.
	var hdrSB strings.Builder
	for k, vs := range req.Header {
		for _, v := range vs {
			hdrSB.WriteString(k)
			hdrSB.WriteString(": ")
			hdrSB.WriteString(v)
			hdrSB.WriteByte('\n')
		}
	}
	hdrBytes := []byte(hdrSB.String())

	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading request body: %w", err)
		}
	}

	// Always pass a valid Go pointer; the host ignores zero-length slices.
	methodPtr := unsafe.Pointer(&methodBytes[0])
	urlPtr := unsafe.Pointer(&urlBytes[0])

	hdrPtr := unsafe.Pointer(&emptyBuf[0])
	hdrLen := uint32(0)
	if len(hdrBytes) > 0 {
		hdrPtr = unsafe.Pointer(&hdrBytes[0])
		hdrLen = uint32(len(hdrBytes))
	}

	bodyPtr := unsafe.Pointer(&emptyBuf[0])
	bodyLen := uint32(0)
	if len(bodyBytes) > 0 {
		bodyPtr = unsafe.Pointer(&bodyBytes[0])
		bodyLen = uint32(len(bodyBytes))
	}

	handle := hostHTTPStreamOpen(
		methodPtr, uint32(len(methodBytes)),
		urlPtr, uint32(len(urlBytes)),
		hdrPtr, hdrLen,
		bodyPtr, bodyLen,
	)
	if handle < 0 {
		return nil, fmt.Errorf("http: open stream for %s %s failed (errno %d)", method, req.URL, -handle)
	}

	status := int(hostHTTPStreamStatus(handle))
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       &streamBody{handle: handle},
		Request:    req,
	}, nil
}

// streamBody is an io.ReadCloser backed by a host stream handle.
type streamBody struct {
	handle int32
	closed bool
}

// Read implements io.Reader. Returns io.EOF when the response body is exhausted.
func (b *streamBody) Read(p []byte) (int, error) {
	if b.closed {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	n := hostHTTPStreamRead(b.handle, unsafe.Pointer(&p[0]), uint32(len(p)))
	if n < 0 {
		return 0, fmt.Errorf("stream read error (errno %d)", -n)
	}
	if n == 0 {
		return 0, io.EOF
	}
	return int(n), nil
}

// Close implements io.Closer.
func (b *streamBody) Close() error {
	if !b.closed {
		b.closed = true
		hostHTTPStreamClose(b.handle)
	}
	return nil
}
