package epsagongin

import (
	"net/http/httptest"
	"testing"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGinWrapper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gin Wrapper")
}

type MockEngine struct {
	*gin.Engine
	TestHandler func(handler gin.HandlerFunc)
}

func (me *MockEngine) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	me.TestHandler(handlers[0])
	return nil
}

var _ = Describe("gin_wrapper", func() {
	Describe("GinRouterWrapper", func() {
		Context("Happy Flows", func() {
			var (
				events       []*protocol.Event
				exceptions   []*protocol.Exception
				wrapper      *GinRouterWrapper
				mockedEngine *MockEngine
				called       bool
			)
			BeforeEach(func() {
				called = false
				events = make([]*protocol.Event, 0)
				exceptions = make([]*protocol.Exception, 0)
				tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
					Events:     &events,
					Exceptions: &exceptions,
					Labels:     make(map[string]interface{}),
				}
				mockedEngine = &MockEngine{Engine: gin.New()}
				wrapper = &GinRouterWrapper{
					IRouter:  mockedEngine,
					Hostname: "test",
					Config: &epsagon.Config{Config: tracer.Config{
						Disable:  true,
						TestMode: true,
					}},
				}
			})
			It("calls the wrapped function", func() {
				testGinContext, _ := gin.CreateTestContext(httptest.NewRecorder())
				mockedEngine.TestHandler = func(handler gin.HandlerFunc) {
					handler(testGinContext)
					Expect(called).To(Equal(true))
				}
				wrapper.GET("/test", func(c *gin.Context) { called = true })
			})
			It("passes the tracer through gin context", func() {
				testGinContext, _ := gin.CreateTestContext(httptest.NewRecorder())
				mockedEngine.TestHandler = func(handler gin.HandlerFunc) {
					handler(testGinContext)
					Expect(called).To(Equal(true))
				}
				wrapper.GET("/test", func(c *gin.Context) {
					tracer := c.Keys[TracerKey].(tracer.Tracer)
					tracer.AddLabel("test", "ok")
					called = true
				})
				Expect(
					tracer.GlobalTracer.(*tracer.MockedEpsagonTracer).Labels["test"],
				).To(Equal("ok"))
			})
			It("Creates a runner event for the handler invocation", func() {
				testGinContext, _ := gin.CreateTestContext(httptest.NewRecorder())
				mockedEngine.TestHandler = func(handler gin.HandlerFunc) {
					handler(testGinContext)
				}
				wrapper.GET("/test", func(c *gin.Context) {
					called = true
				})
				Expect(len(events)).To(Equal(1))
			})
		})
	})
})
