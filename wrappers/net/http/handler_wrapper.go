package epsagonhttp

import (
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
				"path": request.URL.Path,
			},
		},
	}
	if !wrapperTracer.GetConfig().MetadataOnly {
		event.Resource.Metadata["query_string_parameters"] = processRawQuery(
			request.URL, wrapperTracer)
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
		triggerEvent.Resource.Metadata["status_code"] = "200"
		defer func() {
			if userError := recover(); userError != nil {
				triggerEvent.Resource.Metadata["status_code"] = "500"
				panic(userError)
			}
		}()

		newRequest := request.WithContext(
			epsagon.ContextWithTracer(wrapperTracer, request.Context()))

		if !config.MetadataOnly {
			rw = &WrappedResponseWriter{
				ResponseWriter: rw,
				resource:       triggerEvent.Resource,
			}
		}
		defer func() {
			if wrappedResponseWriter, ok := rw.(*WrappedResponseWriter); ok {
				wrappedResponseWriter.UpdateResource()
			}
		}()

		wrapper := epsagon.WrapGenericFunction(
			handler, config, wrapperTracer, false, handlerName,
		)
		wrapper.Call(rw, newRequest)
	}
}
