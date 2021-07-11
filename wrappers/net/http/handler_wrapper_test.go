package epsagonhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("http_wrapper", func() {
	Describe("WrapHandleFunc", func() {
		var (
			events         []*protocol.Event
			exceptions     []*protocol.Exception
			request        *http.Request
			responseWriter *httptest.ResponseRecorder
			called         bool
			config         *epsagon.Config
		)
		BeforeEach(func() {
			called = false
			config = &epsagon.Config{Config: tracer.Config{
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
			body := []byte("hello")
			request = httptest.NewRequest("POST", "https://www.help.com", ioutil.NopCloser(bytes.NewReader(body)))
			responseWriter = httptest.NewRecorder()
		})
		Context("Happy Flows", func() {
			It("calls the wrapped function", func() {
				wrapper := WrapHandleFunc(
					config,
					func(rw http.ResponseWriter, req *http.Request) {
						called = true
					},
				)
				wrapper(responseWriter, request)
				Expect(called).To(Equal(true))
			})
			It("passes the tracer through request context", func() {
				wrapper := WrapHandleFunc(
					config,
					func(rw http.ResponseWriter, req *http.Request) {
						called = true
						tracer := epsagon.ExtractTracer([]context.Context{req.Context()})
						tracer.AddLabel("test", "ok")
					},
				)
				wrapper(responseWriter, request)
				Expect(called).To(Equal(true))
				Expect(
					tracer.GlobalTracer.(*tracer.MockedEpsagonTracer).Labels["test"],
				).To(Equal("ok"))
			})
			It("Creates a runner and trigger events for handler invocation", func() {
				wrapper := WrapHandleFunc(
					config,
					func(rw http.ResponseWriter, req *http.Request) {
						called = true
					},
					"test-handler",
				)
				wrapper(responseWriter, request)
				Expect(len(events)).To(Equal(2))
				var runnerEvent *protocol.Event
				for _, event := range events {
					if event.Origin == "runner" {
						runnerEvent = event
					}
				}
				Expect(runnerEvent).NotTo(Equal(nil))
				Expect(runnerEvent.Resource.Type).To(Equal("go-function"))
				Expect(runnerEvent.Resource.Name).To(Equal("test-handler"))
			})
			It("Adds correct trigger event", func() {
				body := []byte("hello world")
				request = httptest.NewRequest(
					"POST",
					"https://www.help.com/test?hello=world&good=bye",
					ioutil.NopCloser(bytes.NewReader(body)))
				wrapper := WrapHandleFunc(
					config,
					func(rw http.ResponseWriter, req *http.Request) {
						called = true
						internalHandlerBody, err := ioutil.ReadAll(req.Body)
						if err != nil {
							Expect(true).To(Equal(false))
						}
						Expect(internalHandlerBody).To(Equal(body))
						resp, err := json.Marshal(map[string]string{"hello": "world"})
						if err != nil {
							Expect(true).To(Equal(false))
						}
						rw.Header().Add("Content-Type", "application/json; charset=utf-8")
						_, err = rw.Write(resp)
						if err != nil {
							Expect(true).To(Equal(false))
						}

					},
					"test-handler",
				)
				wrapper(responseWriter, request)
				Expect(len(events)).To(Equal(2))
				var triggerEvent *protocol.Event
				for _, event := range events {
					if event.Origin == "trigger" {
						triggerEvent = event
					}
				}
				Expect(triggerEvent).NotTo(Equal(nil))
				Expect(triggerEvent.Resource.Name).To(Equal("www.help.com"))
				Expect(triggerEvent.Resource.Type).To(Equal("http"))
				Expect(triggerEvent.Resource.Operation).To(Equal("POST"))
				expectedQuery, _ := json.Marshal(map[string][]string{
					"hello": {"world"}, "good": {"bye"}})
				Expect(triggerEvent.Resource.Metadata["query_string_parameters"]).To(
					Equal(string(expectedQuery)))
				Expect(triggerEvent.Resource.Metadata["path"]).To(
					Equal("/test"))
				Expect(triggerEvent.Resource.Metadata["request_body"]).To(
					Equal(string(body)))
				Expect(triggerEvent.Resource.Metadata["response_body"]).To(
					Equal("{\"hello\":\"world\"}"))
				Expect(triggerEvent.Resource.Metadata["response_headers"]).To(
					Equal("{\"Content-Type\":\"application/json; charset=utf-8\"}"))
				Expect(triggerEvent.Resource.Metadata["status_code"]).To(
					Equal("200"))
			})
			It("Doesn't collect body and headers if MetadataOnly", func() {
				config.MetadataOnly = true
				body := []byte("hello world")
				request = httptest.NewRequest(
					"POST",
					"https://www.help.com/test?hello=world&good=bye",
					ioutil.NopCloser(bytes.NewReader(body)))
				wrapper := WrapHandleFunc(
					config,
					func(rw http.ResponseWriter, req *http.Request) {
						called = true
						internalHandlerBody, err := ioutil.ReadAll(req.Body)
						if err != nil {
							Expect(true).To(Equal(false))
						}
						Expect(internalHandlerBody).To(Equal(body))
						resp, err := json.Marshal(map[string]string{"hello": "world"})
						if err != nil {
							Expect(true).To(Equal(false))
						}
						rw.Header().Add("Content-Type", "application/json; charset=utf-8")
						_, err = rw.Write(resp)
						if err != nil {
							Expect(true).To(Equal(false))
						}

					},
					"test-handler",
				)
				wrapper(responseWriter, request)
				Expect(len(events)).To(Equal(2))
				var triggerEvent *protocol.Event
				for _, event := range events {
					if event.Origin == "trigger" {
						triggerEvent = event
					}
				}
				Expect(triggerEvent).NotTo(Equal(nil))
				Expect(triggerEvent.Resource.Metadata["request_body"]).To(
					Equal(""))
				Expect(triggerEvent.Resource.Metadata["response_body"]).To(
					Equal(""))
				Expect(triggerEvent.Resource.Metadata["response_headers"]).To(
					Equal(""))
				Expect(triggerEvent.Resource.Metadata["query_string_parameters"]).To(
					Equal(""))
			})
		})
		Context("Error Flows", func() {
			It("Adds Exception if handler explodes", func() {
				errorMessage := "boom"
				Expect(func() {
					wrapper := WrapHandleFunc(
						config,
						func(rw http.ResponseWriter, req *http.Request) {
							called = true
							panic(errorMessage)
						},
						"test-handler",
					)
					wrapper(responseWriter, request)
				}).To(
					PanicWith(epsagon.MatchUserError(errorMessage)))
				Expect(called).To(Equal(true))
				Expect(len(events)).To(Equal(2))
				var runnerEvent, triggerEvent *protocol.Event
				for _, event := range events {
					if event.Origin == "runner" {
						runnerEvent = event
					}
					if event.Origin == "trigger" {
						triggerEvent = event
					}
				}
				Expect(runnerEvent.Exception).NotTo(Equal(nil))
				Expect(triggerEvent.Exception).NotTo(Equal(nil))
				Expect(triggerEvent.Resource.Metadata["status_code"]).To(
					Equal("500"))
			})
		})
	})
})
