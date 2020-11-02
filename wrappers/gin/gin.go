package epsagongin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/epsagon/epsagon-go/wrappers/net/http"
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

type wrappedGinWriter struct {
	gin.ResponseWriter
	htrw http.ResponseWriter
}

func wrapGinHandler(handler gin.HandlerFunc, hostname string, relativePath string, config *epsagon.Config) gin.HandlerFunc {
	if config == nil {
		config = &epsagon.Config{}
	}
	return func(c *gin.Context) {
		wrapperTracer := tracer.CreateTracer(&config.Config)
		wrapperTracer.Start()
		defer wrapperTracer.Stop()

		c.Set(TracerKey, wrapperTracer)
		wrapper := epsagon.WrapGenericFunction(
			handler, config, wrapperTracer, false, relativePath,
		)
		triggerEvent := epsagonhttp.CreateHTTPTriggerEvent(
			wrapperTracer, c.Request, hostname)
		wrapperTracer.AddEvent(triggerEvent)
		wrappedResponseWriter := &wrappedGinWriter{
			ResponseWriter: c.Writer,
			htrw:           epsagonhttp.CreateWrappedResponseWriter(c.Writer, triggerEvent.Resource),
		}
		c.Writer = wrappedResponseWriter
		defer func() {
			wrappedResponseWriter.htrw.(*epsagonhttp.WrappedResponseWriter).UpdateResource()
			userError := recover()
			if userError != nil {
				triggerEvent.Resource.Metadata["status_code"] = "500"
				panic(userError)
			}
		}()

		defer func() {
			runner := wrapperTracer.GetRunnerEvent()
			if runner != nil {
				runner.Resource.Type = "gin"
			}
		}()
		wrapper.Call(c)
		triggerEvent.Resource.Metadata["status_code"] = fmt.Sprint(c.Writer.Status())
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

func (grw *wrappedGinWriter) Header() http.Header {
	return grw.htrw.Header()
}
func (grw *wrappedGinWriter) Write(data []byte) (int, error) {
	return grw.htrw.Write(data)
}
func (grw *wrappedGinWriter) WriteHeader(statusCode int) {
	grw.htrw.WriteHeader(statusCode)
}
