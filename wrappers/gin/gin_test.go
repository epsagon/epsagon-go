package epsagongin

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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
		var (
			events         []*protocol.Event
			exceptions     []*protocol.Exception
			wrapper        *GinRouterWrapper
			mockedEngine   *MockEngine
			called         bool
			testGinContext *gin.Context
		)
		BeforeEach(func() {
			called = false
			config := &epsagon.Config{Config: tracer.Config{
				Disable:  true,
				TestMode: true,
			}}
			events = make([]*protocol.Event, 0)
			exceptions = make([]*protocol.Exception, 0)
			tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
				Events:     &events,
				Exceptions: &exceptions,
				Labels:     make(map[string]interface{}),
				Config:     &config.Config,
			}
			mockedEngine = &MockEngine{
				Engine: gin.New(),
				TestHandler: func(handler gin.HandlerFunc) {
					handler(testGinContext)
				},
			}
			wrapper = &GinRouterWrapper{
				IRouter:  mockedEngine,
				Hostname: "test",
				Config:   config,
			}
			body := []byte("hello")
			testGinContext, _ = gin.CreateTestContext(httptest.NewRecorder())
			testGinContext.Request = httptest.NewRequest("POST", "https://www.help.com", ioutil.NopCloser(bytes.NewReader(body)))
			Expect(testGinContext.Request).NotTo(Equal(nil))
		})
		Context("Happy Flows", func() {
			It("calls the wrapped function", func() {
				mockedEngine.TestHandler = func(handler gin.HandlerFunc) {
					handler(testGinContext)
					Expect(called).To(Equal(true))
				}
				wrapper.GET("/test", func(c *gin.Context) { called = true })
			})
			It("passes the tracer through gin context", func() {
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
			It("Creates a runner and trigger events for handler invocation", func() {
				mockedEngine.TestHandler = func(handler gin.HandlerFunc) {
					handler(testGinContext)
				}
				wrapper.GET("/test", func(c *gin.Context) {
					called = true
				})
				Expect(len(events)).To(Equal(2))
			})
			It("Adds correct trigger event", func() {
				body := []byte("hello world")
				testGinContext.Request = httptest.NewRequest(
					"POST",
					"https://www.help.com/test?hello=world&good=bye",
					ioutil.NopCloser(bytes.NewReader(body)))
				wrapper.Hostname = ""
				wrapper.GET("/test", func(c *gin.Context) {
					internalHandlerBody, err := ioutil.ReadAll(c.Request.Body)
					if err != nil {
						Expect(true).To(Equal(false))
					}
					Expect(internalHandlerBody).To(Equal(body))
					called = true
				})
				Expect(len(events)).To(Equal(2))
				var triggerEvent *protocol.Event
				for _, event := range events {
					if event.Origin == "trigger" {
						triggerEvent = event
					}
				}
				Expect(triggerEvent).NotTo(Equal(nil))
				Expect(triggerEvent.Resource.Name).To(Equal("www.help.com"))
				expectedQuery, _ := json.Marshal(map[string][]string{
					"hello": {"world"}, "good": {"bye"}})
				Expect(triggerEvent.Resource.Metadata["query_string_parameters"]).To(
					Equal(string(expectedQuery)))
				Expect(triggerEvent.Resource.Metadata["path"]).To(
					Equal("/test"))
				Expect(triggerEvent.Resource.Metadata["request_body"]).To(
					Equal(string(body)))
			})
		})
		Context("Error Flows", func() {
			It("Adds Exception if handler explodes", func() {
				errorMessage := "boom"
				Expect(func() {
					wrapper.GET("/test", func(c *gin.Context) {
						panic(errorMessage)
					})
				}).To(
					PanicWith(epsagon.MatchUserError(errorMessage)))
				Expect(len(events)).To(Equal(2))
				var runnerEvent *protocol.Event
				for _, event := range events {
					if event.Origin == "runner" {
						runnerEvent = event
					}
				}
				Expect(runnerEvent.Exception).NotTo(Equal(nil))
			})
		})
	})
})
