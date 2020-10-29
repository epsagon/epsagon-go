package epsagonhttp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const TEST_RESPONSE_STRING = "response_test_string"

func TestEpsagonHTTPWrappers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "epsagon http wrapper suite")
}

func verifyTraceIDExists(event *protocol.Event) {
	traceID, ok := event.Resource.Metadata[tracer.EpsagonHTTPTraceIDKey]
	Expect(ok).To(BeTrue())
	Expect(traceID).To(Not(BeZero()))
}

func verifyTraceIDNotExists(event *protocol.Event) {
	Expect(event.Resource.Metadata).NotTo(
		HaveKey(tracer.EpsagonHTTPTraceIDKey))
}

func verifyResponseSuccess(response *http.Response, err error) {
	Expect(err).To(BeNil())
	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	Expect(err).To(BeNil())
	responseString := string(responseData)
	Expect(responseString).To(Equal(TEST_RESPONSE_STRING))
}

type mockTransport struct {
	called bool
}

func (m *mockTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	m.called = true
	return http.DefaultTransport.RoundTrip(req)
}

var _ = Describe("ClientWrapper", func() {
	var (
		events        []*protocol.Event
		exceptions    []*protocol.Exception
		requests      []*http.Request
		testServer    *httptest.Server
		response_data []byte
	)
	BeforeEach(func() {
		requests = make([]*http.Request, 0)
		events = make([]*protocol.Event, 0)
		exceptions = make([]*protocol.Exception, 0)
		response_data = []byte(TEST_RESPONSE_STRING)
		tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
			Events:     &events,
			Exceptions: &exceptions,
			Config:     &tracer.Config{},
		}
		testServer = httptest.NewServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				requests = append(requests, req)
				res.Write(response_data)
			}))
	})
	AfterEach(func() {
		tracer.GlobalTracer = nil
		testServer.Close()
	})

	Describe(".Do", func() {
		BeforeEach(func() {
			events = make([]*protocol.Event, 0)
			exceptions = make([]*protocol.Exception, 0)
			requests = make([]*http.Request, 0)
		})
		Context("sending a request to existing server", func() {
			It("adds an event with no error", func() {
				client := Wrap(http.Client{})
				req, err := http.NewRequest(http.MethodGet, testServer.URL, nil)
				if err != nil {
					Fail("couldn't create request")
				}
				client.Do(req)
				Expect(requests).To(HaveLen(1))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				verifyTraceIDExists(events[0])
			})
		})
		Context("sending a request to existing server, no tracer", func() {
			It("adds an event with no error", func() {
				tracer.GlobalTracer = nil
				client := Wrap(http.Client{})
				req, err := http.NewRequest(http.MethodGet, testServer.URL, nil)
				if err != nil {
					Fail("couldn't create request")
				}
				response, err := client.Do(req)
				verifyResponseSuccess(response, err)
			})
		})
		Context("request to whitelisted url", func() {
			It("Adds event with trace ID", func() {
				client := Wrap(http.Client{})
				req, err := http.NewRequest(
					http.MethodGet,
					fmt.Sprintf("https://test.%s.com", APPSYNC_API_SUBDOMAIN),
					nil,
				)
				if err != nil {
					Fail("couldn't create request")
				}
				client.Do(req)
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
				verifyTraceIDExists(events[0])
			})
		})
		Context("request to blacklisted url", func() {
			It("Adds event with trace ID", func() {
				client := Wrap(http.Client{})
				req, err := http.NewRequest(
					http.MethodGet,
					fmt.Sprintf("https://%s", EPSAGON_DOMAIN),
					nil,
				)
				if err != nil {
					Fail("couldn't create request")
				}
				client.Do(req)
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				verifyTraceIDNotExists(events[0])
			})
		})
	})
	Describe(".Get", func() {
		BeforeEach(func() {
			events = make([]*protocol.Event, 0)
			exceptions = make([]*protocol.Exception, 0)
		})
		Context("request created succesfully", func() {
			It("Adds event", func() {
				client := Wrap(http.Client{})
				client.Get(testServer.URL)
				Expect(requests).To(HaveLen(1))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				Expect(events[0].Resource.Metadata["response_body"]).To(
					Equal(string(response_data)))
				verifyTraceIDExists(events[0])
			})
		})
		Context("sending a request to existing server, no tracer", func() {
			It("adds an event with no error", func() {
				tracer.GlobalTracer = nil
				client := Wrap(http.Client{})
				response, err := client.Get(testServer.URL)
				verifyResponseSuccess(response, err)
			})
		})
		Context("request to whitelisted url", func() {
			It("Adds event with trace ID", func() {
				client := Wrap(http.Client{})
				client.Get(fmt.Sprintf("https://test.%s.com", APPSYNC_API_SUBDOMAIN))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
				verifyTraceIDExists(events[0])
			})
		})
		Context("request to blacklisted url", func() {
			It("Adds event with trace ID", func() {
				client := Wrap(http.Client{})
				client.Get(fmt.Sprintf("https://%s", EPSAGON_DOMAIN))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				verifyTraceIDNotExists(events[0])
			})
		})
		Context("bad input failing to create request", func() {
			It("Adds event with error code error", func() {
				client := Wrap(http.Client{})
				client.Get(testServer.URL + "balbla")
				Expect(requests).To(HaveLen(0))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
				verifyTraceIDNotExists(events[0])
			})
		})
	})
	Describe(".Post", func() {
		BeforeEach(func() {
			events = make([]*protocol.Event, 0)
			exceptions = make([]*protocol.Exception, 0)
		})
		Context("request created succesfully", func() {
			It("Adds event", func() {
				client := Wrap(http.Client{})
				data := "{\"hello\":\"world\"}"
				client.Post(
					testServer.URL,
					"application/json",
					strings.NewReader(data))
				Expect(requests).To(HaveLen(1))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				Expect(events[0].Resource.Metadata["response_body"]).To(
					Equal(string(response_data)))
				Expect(events[0].Resource.Metadata["request_body"]).To(
					Equal(data))
				verifyTraceIDExists(events[0])
			})
		})
		Context("sending a request to existing server, no tracer", func() {
			It("adds an event with no error", func() {
				tracer.GlobalTracer = nil
				client := Wrap(http.Client{})
				data := "{\"hello\":\"world\"}"
				response, err := client.Post(
					testServer.URL,
					"application/json",
					strings.NewReader(data))
				verifyResponseSuccess(response, err)
			})
		})
		Context("client with metadataOnly", func() {
			It("Adds event", func() {
				client := Wrap(http.Client{})
				client.MetadataOnly = true
				data := "{\"hello\":\"world\"}"
				client.Post(
					testServer.URL,
					"application/json",
					strings.NewReader(data))
				Expect(requests).To(HaveLen(1))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				Expect(events[0].Resource.Metadata).NotTo(
					HaveKey("response_body"))
				Expect(events[0].Resource.Metadata).NotTo(
					HaveKey("request_body"))
				verifyTraceIDExists(events[0])
			})
		})
		Context("request to whitelisted url", func() {
			It("Adds event with trace ID", func() {
				client := Wrap(http.Client{})
				data := "{\"hello\":\"world\"}"
				client.Post(
					fmt.Sprintf("https://test.%s.com", APPSYNC_API_SUBDOMAIN),
					"application/json",
					strings.NewReader(data))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
				verifyTraceIDExists(events[0])
			})
		})
		Context("request to blacklisted url", func() {
			It("Adds event with trace ID", func() {
				client := Wrap(http.Client{})
				data := "{\"hello\":\"world\"}"
				client.Post(
					fmt.Sprintf("https://%s", EPSAGON_DOMAIN),
					"application/json",
					strings.NewReader(data))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				verifyTraceIDNotExists(events[0])
			})
		})
		Context("bad input failing to create request", func() {
			It("Adds event", func() {
				client := Wrap(http.Client{})
				client.Post(
					testServer.URL+"blabla",
					"application/json",
					strings.NewReader("{\"hello\":\"world\"}"))
				Expect(requests).To(HaveLen(0))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
				verifyTraceIDNotExists(events[0])
			})
		})
	})
	Describe(".PostForm", func() {
		BeforeEach(func() {
			events = make([]*protocol.Event, 0)
			exceptions = make([]*protocol.Exception, 0)
		})
		Context("request created succesfully", func() {
			It("Adds event", func() {
				client := Wrap(http.Client{})
				client.PostForm(
					testServer.URL,
					map[string][]string{
						"hello": []string{"world", "of", "serverless"},
					},
				)
				Expect(requests).To(HaveLen(1))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				verifyTraceIDExists(events[0])
			})
		})
		Context("sending a request to existing server, no tracer", func() {
			It("adds an event with no error", func() {
				tracer.GlobalTracer = nil
				client := Wrap(http.Client{})
				response, err := client.PostForm(
					testServer.URL,
					map[string][]string{
						"hello": []string{"world", "of", "serverless"},
					},
				)
				verifyResponseSuccess(response, err)
			})
		})
		Context("request to whitelisted url", func() {
			It("Adds event with trace ID", func() {
				client := Wrap(http.Client{})
				client.PostForm(
					fmt.Sprintf("https://test.%s.com", APPSYNC_API_SUBDOMAIN),
					map[string][]string{
						"hello": []string{"world", "of", "serverless"},
					},
				)
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
				verifyTraceIDExists(events[0])
			})
		})
		Context("request to blacklisted url", func() {
			It("Adds event with trace ID", func() {
				client := Wrap(http.Client{})
				client.PostForm(
					fmt.Sprintf("https://%s", EPSAGON_DOMAIN),
					map[string][]string{
						"hello": []string{"world", "of", "serverless"},
					},
				)
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				verifyTraceIDNotExists(events[0])
			})
		})
		Context("bad input failing to create request", func() {
			It("Adds event with error code error", func() {
				client := Wrap(http.Client{})
				client.PostForm(
					testServer.URL+"blabla",
					map[string][]string{
						"hello": []string{"world", "of", "serverless"},
					},
				)
				Expect(requests).To(HaveLen(0))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
				verifyTraceIDNotExists(events[0])
			})
		})
	})
	Describe(".Head", func() {
		BeforeEach(func() {
			events = make([]*protocol.Event, 0)
			exceptions = make([]*protocol.Exception, 0)
		})
		Context("request created succesfully", func() {
			It("Adds event", func() {
				client := Wrap(http.Client{})
				client.Head(testServer.URL)
				Expect(requests).To(HaveLen(1))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				verifyTraceIDExists(events[0])
			})
		})
		Context("sending a request to existing server, no tracer", func() {
			It("adds an event with no error", func() {
				tracer.GlobalTracer = nil
				client := Wrap(http.Client{})
				_, err := client.Head(testServer.URL)
				Expect(err).To(BeNil())
			})
		})
		Context("request to whitelisted url", func() {
			It("Adds event with trace ID", func() {
				client := Wrap(http.Client{})
				client.Head(fmt.Sprintf("https://test.%s.com", APPSYNC_API_SUBDOMAIN))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
				verifyTraceIDExists(events[0])
			})
		})
		Context("request to blacklisted url", func() {
			It("Adds event with trace ID", func() {
				client := Wrap(http.Client{})
				client.Head(fmt.Sprintf("https://%s", EPSAGON_DOMAIN))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				verifyTraceIDNotExists(events[0])
			})
		})
		Context("bad input failing to create request", func() {
			It("Adds event with error code error", func() {
				client := Wrap(http.Client{})
				client.Head(testServer.URL + "blabla")
				Expect(requests).To(HaveLen(0))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
				verifyTraceIDNotExists(events[0])
			})
		})
	})
	Describe("http.RoundTripper", func() {
		BeforeEach(func() {
			events = make([]*protocol.Event, 0)
			exceptions = make([]*protocol.Exception, 0)
			requests = make([]*http.Request, 0)
		})
		Context("sending a request to existing server", func() {
			It("adds an event with no error, truncating the request body", func() {
				client := &http.Client{Transport: NewTracingTransport()}
				data := make([]byte, 128*1024)
				for i := range data {
					data[i] = byte(1)
				}
				req, err := http.NewRequest(
					http.MethodPost,
					testServer.URL,
					bytes.NewReader(data))
				if err != nil {
					Fail("couldn't create request")
				}
				client.Do(req)
				Expect(requests).To(HaveLen(1))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				Expect([]byte(events[0].Resource.Metadata["request_body"])).To(HaveCap(epsagon.MaxMetadataSize))
				verifyTraceIDExists(events[0])
			})
		})
		Context("sending a request to existing server, no tracer", func() {
			It("adds an event with no error", func() {
				tracer.GlobalTracer = nil
				client := &http.Client{Transport: NewTracingTransport()}
				req, err := http.NewRequest(http.MethodGet, testServer.URL, nil)
				if err != nil {
					Fail("couldn't create request")
				}
				response, err := client.Do(req)
				verifyResponseSuccess(response, err)
			})
		})
		Context("request to whitelisted url", func() {
			It("Adds event with trace ID", func() {
				client := &http.Client{Transport: NewTracingTransport()}
				req, err := http.NewRequest(
					http.MethodGet,
					fmt.Sprintf("https://test.%s.com", APPSYNC_API_SUBDOMAIN),
					nil,
				)
				if err != nil {
					Fail("couldn't create request")
				}
				client.Do(req)
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
				verifyTraceIDExists(events[0])
			})
		})
		Context("request to blacklisted url", func() {
			It("Adds event with trace ID and the response truncated", func() {
				client := &http.Client{Transport: NewTracingTransport()}
				req, err := http.NewRequest(
					http.MethodGet,
					fmt.Sprintf("https://%s", EPSAGON_DOMAIN),
					nil,
				)
				if err != nil {
					Fail("couldn't create request")
				}
				client.Do(req)
				Expect(events).To(HaveLen(1))
				Expect([]byte(events[0].Resource.Metadata["response_body"])).To(HaveCap(10 * 1024))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				verifyTraceIDNotExists(events[0])
			})
		})
		Context("wrapping a custom transport, request created succesfully", func() {
			It("Adds event", func() {
				mock := &mockTransport{}
				client := &http.Client{Transport: NewWrappedTracingTransport(mock)}
				client.Head(testServer.URL)
				Expect(requests).To(HaveLen(1))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
				Expect(mock.called).To(BeTrue())
				verifyTraceIDExists(events[0])
			})
		})
	})
})
