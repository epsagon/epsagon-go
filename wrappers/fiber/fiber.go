package epsagonfiber

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

// FiberEpsagonMiddleware is an epsagon instumentation middleware for fiber apps
type FiberEpsagonMiddleware struct {
	Config *epsagon.Config
}

func parseQueryArgs(args *fasthttp.Args) map[string]string {
	result := make(map[string]string)
	args.VisitAll(func(key, val []byte) {
		result[string(key)] = string(val)
	})
	return result
}

// convert map values to string. On error, add exception to tracer with the
// given error message
func convertMapValuesToString(
	values map[string]string,
	wrapperTracer tracer.Tracer,
	errorMessage string) string {
	processed, err := json.Marshal(values)
	if err != nil {
		wrapperTracer.AddException(&protocol.Exception{
			Type:      "trigger-creation",
			Message:   errorMessage,
			Traceback: string(debug.Stack()),
			Time:      tracer.GetTimestamp(),
		})
		return ""
	}
	return string(processed)
}

func processQueryFromURI(uriObj *fasthttp.URI, wrapperTracer tracer.Tracer) string {
	if uriObj == nil {
		return ""
	}
	args := parseQueryArgs(uriObj.QueryArgs())
	return convertMapValuesToString(
		args,
		wrapperTracer,
		fmt.Sprintf("Failed to serialize query params %s", uriObj.QueryString()))

}

func processHeaders(rawHeaders interface{}, wrapperTracer tracer.Tracer) string {
	if rawHeaders == nil {
		return ""
	}
	headers := make(map[string]string)
	headersType := ""
	switch rawHeaders.(type) {
	case *fasthttp.RequestHeader:
		headersType = "request"
		rawHeaders.(*fasthttp.RequestHeader).VisitAll(func(key, val []byte) {
			headers[string(key)] = string(val)
		})
	case *fasthttp.ResponseHeader:
		headersType = "response"
		rawHeaders.(*fasthttp.ResponseHeader).VisitAll(func(key, val []byte) {
			headers[string(key)] = string(val)
		})
	default:
		if wrapperTracer.GetConfig().Debug {
			log.Printf("EPSAGON DEBUG: received an invalid type of headers")
		}
		return ""
	}
	return convertMapValuesToString(
		headers,
		wrapperTracer,
		fmt.Sprintf("Failed to serialize %s headers", headersType))
}

// Gets the request content type header, empty string if not found
func getRequestContentType(fiberCtx *fiber.Ctx) string {
	return string(fiberCtx.Request().Header.ContentType())
}

// CreateHTTPTriggerEvent creates an HTTP trigger event
func CreateHTTPTriggerEvent(wrapperTracer tracer.Tracer, fiberCtx *fiber.Ctx, resourceName string) *protocol.Event {
	request := fiberCtx.Request()

	name := resourceName
	if len(name) == 0 {
		name = fiberCtx.Hostname()
	}

	event := &protocol.Event{
		Id:        "",
		Origin:    "trigger",
		StartTime: tracer.GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      name,
			Type:      "http",
			Operation: fiberCtx.Method(),
			Metadata: map[string]string{
				"path": fiberCtx.Path(),
			},
		},
	}
	if !wrapperTracer.GetConfig().MetadataOnly {
		event.Resource.Metadata["query_string_params"] = processQueryFromURI(request.URI(), wrapperTracer)
		event.Resource.Metadata["request_headers"] = processHeaders(&request.Header, wrapperTracer)
		event.Resource.Metadata["request_body"] = string(fiberCtx.Body())
	}
	return event
}

func (middleware *FiberEpsagonMiddleware) HandlerFunc() fiber.Handler {
	config := middleware.Config
	if config == nil {
		config = &epsagon.Config{}
	}
	return func(c *fiber.Ctx) (err error) {
		if epsagon.ShouldIgnoreRequest(getRequestContentType(c), c.Path()) {
			return c.Next()
		}

		callingOriginalHandler := false
		called := false
		var triggerEvent *protocol.Event = nil
		defer func() {
			userError := recover()
			if userError == nil {
				return
			}
			if !callingOriginalHandler {
				err = c.Next()
				return
			}
			if !called { // panic only if error happened in original handler
				triggerEvent.Resource.Metadata["status_code"] = "500"
				panic(userError)
			}
		}()

		wrapperTracer := tracer.CreateTracer(&config.Config)
		wrapperTracer.Start()
		defer wrapperTracer.SendStopSignal()
		userContext := c.UserContext()
		c.SetUserContext(epsagon.ContextWithTracer(wrapperTracer, userContext))
		triggerEvent = CreateHTTPTriggerEvent(wrapperTracer, c, c.Hostname())
		wrapperTracer.AddEvent(triggerEvent)
		wrapper := epsagon.WrapGenericFunction(
			func(c *fiber.Ctx) error {
				err = c.Next()
				return err
			}, config, wrapperTracer, false, c.Path())
		defer postExecutionUpdates(wrapperTracer, triggerEvent, c, wrapper)
		callingOriginalHandler = true
		wrapper.Call(c)
		called = true
		return err
	}
}

func postExecutionUpdates(
	wrapperTracer tracer.Tracer, triggerEvent *protocol.Event,
	c *fiber.Ctx, handlerWrapper *epsagon.GenericWrapper) {
	runner := handlerWrapper.GetRunnerEvent()
	if runner != nil {
		runner.Resource.Type = "fiber"
	}
	response := c.Response()
	triggerEvent.Resource.Metadata["status_code"] = strconv.Itoa(response.StatusCode())
	if !wrapperTracer.GetConfig().MetadataOnly {
		triggerEvent.Resource.Metadata["response_headers"] = processHeaders(&response.Header, wrapperTracer)
		triggerEvent.Resource.Metadata["response_body"] = string(response.Body())
	}
}
