package epsagonhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
)

// HandlerFunction is a generic http handler function
type HandlerFunction func(http.ResponseWriter, *http.Request)

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

func processRawQuery(urlObj *url.URL, wrapperTracer tracer.Tracer) string {
	if urlObj == nil {
		return ""
	}
	processed, err := json.Marshal(urlObj.Query())
	if err != nil {
		wrapperTracer.AddException(&protocol.Exception{
			Type:      "trigger-creation",
			Message:   fmt.Sprintf("Failed to serialize query params %s", urlObj.RawQuery),
			Traceback: string(debug.Stack()),
			Time:      tracer.GetTimestamp(),
		})
		return ""
	}
	return string(processed)
}

// CreateHTTPTriggerEvent creates an HTTP trigger event
func CreateHTTPTriggerEvent(wrapperTracer tracer.Tracer, request *http.Request, resourceName string) *protocol.Event {
	name := resourceName
	if len(name) == 0 {
		name = request.Host
	}
	event := &protocol.Event{
		Id:        "",
		Origin:    "trigger",
		StartTime: tracer.GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      name,
			Type:      "http",
			Operation: request.Method,
			Metadata: map[string]string{
				"query_string_parameters": processRawQuery(
					request.URL, wrapperTracer),
				"path": request.URL.Path,
			},
		},
	}
	if !wrapperTracer.GetConfig().MetadataOnly {
		headers, body := epsagon.ExtractRequestData(request)
		event.Resource.Metadata["request_headers"] = headers
		event.Resource.Metadata["request_body"] = body
	}
	return event
}

// WrapHandleFunc wraps a generic http.HandleFunc handler function with Epsagon
// Last two optional paramerts are the name of the handler (will be the resource name of the events) and an optional hardcoded hostname
func WrapHandleFunc(
	config *epsagon.Config, handler HandlerFunction, names ...string) HandlerFunction {
	var hostName, handlerName string
	if config == nil {
		config = &epsagon.Config{}
	}
	if len(names) >= 1 {
		handlerName = names[0]
	}
	if len(names) >= 2 {
		hostName = names[1]
	}
	return func(rw http.ResponseWriter, request *http.Request) {
		wrapperTracer := tracer.CreateTracer(&config.Config)
		wrapperTracer.Start()
		defer wrapperTracer.Stop()

		triggerEvent := CreateHTTPTriggerEvent(
			wrapperTracer, request, hostName)
		wrapperTracer.AddEvent(triggerEvent)
		triggerEvent.Resource.Metadata["status_code"] = "500"

		newRequest := request.WithContext(
			epsagon.ContextWithTracer(wrapperTracer, request.Context()))
		wrappedResponseWriter := WrappedResponseWriter{
			ResponseWriter: rw,
			resource:       triggerEvent.Resource,
		}
		defer func() {
			wrappedResponseWriter.UpdateResource()
		}()

		wrapper := epsagon.WrapGenericFunction(
			handler, config, wrapperTracer, false, handlerName,
		)
		wrapper.Call(wrappedResponseWriter, newRequest)
	}
}
