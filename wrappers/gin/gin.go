package epsagongin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"

	"github.com/epsagon/epsagon-go/protocol"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/gin-gonic/gin"
)

// TracerKey is the key of the epsagon tracer in the gin.Context Keys map passed to the handlers
const TracerKey = "EpsagonTracer"

// EpsagonContext creates a context.Background() with epsagon's associated tracer for nexted instrumentations
func EpsagonContext(c *gin.Context) context.Context {
	return epsagon.ContextWithTracer(c.Keys[TracerKey].(tracer.Tracer))
}

// GinRouterWrapper is an epsagon instumentation wrapper for gin.RouterGroup
type GinRouterWrapper struct {
	gin.IRouter
	Hostname string
	Config   *epsagon.Config
}

func processRawQuery(urlObj *url.URL, wrapperTracer tracer.Tracer) string {
	processed, err := json.Marshal(urlObj.Query())
	if err != nil {
		wrapperTracer.AddException(&protocol.Exception{
			Type:      "trigger-creation",
			Message:   fmt.Sprintf("Failed to serialize query params %s", urlObj.RawQuery),
			Traceback: string(debug.Stack()),
			Time:      tracer.GetTimestamp(),
		})
	}
	return string(processed)
}

func addTriggerEvent(wrapperTracer tracer.Tracer, context *gin.Context, resourceName string) {
	name := resourceName
	if len(name) == 0 {
		name = context.Request.Host
	}
	event := &protocol.Event{
		Id:        "",
		Origin:    "trigger",
		StartTime: tracer.GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      name,
			Type:      "http",
			Operation: context.Request.Method,
			Metadata: map[string]string{
				"query_string_parameters": processRawQuery(
					context.Request.URL, wrapperTracer),
				"path": context.Request.URL.Path,
			},
		},
	}
	if !wrapperTracer.GetConfig().MetadataOnly {
		headers, body := epsagon.ExtractRequestData(context.Request)
		event.Resource.Metadata["request_headers"] = headers
		event.Resource.Metadata["request_body"] = body
	}
	wrapperTracer.AddEvent(event)
}

func wrapGinHandler(handler gin.HandlerFunc, hostname string, relativePath string, config *epsagon.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config == nil {
			config = &epsagon.Config{}
		}
		wrapperTracer := tracer.CreateTracer(&config.Config)
		wrapperTracer.Start()
		defer wrapperTracer.Stop()

		if c.Keys == nil {
			c.Keys = make(map[string]interface{})
		}
		c.Keys[TracerKey] = wrapperTracer
		wrapper := epsagon.WrapGenericFunction(
			handler, config, wrapperTracer, false, relativePath,
		)
		addTriggerEvent(wrapperTracer, c, hostname)
		wrapper.Call(c)

		runner := wrapperTracer.GetRunnerEvent()
		if runner != nil {
			runner.Resource.Type = "gin"
		}
	}
}

// Handle is a wrapper for gin.RouterGroup.Handle that adds epsagon instrumentaiton and event triggers
// to all invocations of that handler
func (router *GinRouterWrapper) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	if len(handlers) >= 1 {
		handlers[0] = wrapGinHandler(handlers[0], router.Hostname, relativePath, router.Config)
	}
	return router.IRouter.Handle(httpMethod, relativePath, handlers...)
}

// POST is a shortcut for router.Handle("POST", path, handle).
func (router *GinRouterWrapper) POST(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return router.Handle(http.MethodPost, relativePath, handlers...)
}

// GET is a shortcut for router.Handle("GET", path, handle).
func (router *GinRouterWrapper) GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return router.Handle(http.MethodGet, relativePath, handlers...)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle).
func (router *GinRouterWrapper) DELETE(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return router.Handle(http.MethodDelete, relativePath, handlers...)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle).
func (router *GinRouterWrapper) PATCH(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return router.Handle(http.MethodPatch, relativePath, handlers...)
}

// PUT is a shortcut for router.Handle("PUT", path, handle).
func (router *GinRouterWrapper) PUT(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return router.Handle(http.MethodPut, relativePath, handlers...)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle).
func (router *GinRouterWrapper) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return router.Handle(http.MethodOptions, relativePath, handlers...)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle).
func (router *GinRouterWrapper) HEAD(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return router.Handle(http.MethodHead, relativePath, handlers...)
}
