package epsagonhttp

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
)

// WrappedResponseWriter is wrapping Resposne writer with Epsagon
// to enrich the trace with data from the response
type WrappedResponseWriter struct {
	http.ResponseWriter
	resource *protocol.Resource
	buf      bytes.Buffer
}

// CreateWrappedResponseWriter creates a newWrappedResponseWriter
func CreateWrappedResponseWriter(rw http.ResponseWriter, resource *protocol.Resource) *WrappedResponseWriter {
	return &WrappedResponseWriter{
		ResponseWriter: rw,
		resource:       resource,
		buf:            bytes.Buffer{},
	}
}

// Header wrapper
func (w *WrappedResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// WriteHeader wrapper, will set status_code immediately
func (w *WrappedResponseWriter) WriteHeader(statusCode int) {
	w.resource.Metadata["status_code"] = fmt.Sprint(statusCode)
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write wrapper
func (w *WrappedResponseWriter) Write(data []byte) (int, error) {
	w.buf.Write(data)
	return w.ResponseWriter.Write(data)
}

// UpdateResource updates the connected resource with the response headers and body
func (w *WrappedResponseWriter) UpdateResource() {
	w.resource.Metadata["response_headers"], _ = epsagon.FormatHeaders(w.ResponseWriter.Header())
	w.resource.Metadata["response_body"] = w.buf.String()
}
