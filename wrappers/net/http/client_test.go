package epsagonhttp

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
	"testing"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"net/http"
	"net/http/httptest"
)

func TestEpsagonHTTPWrappers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "epsagon http wrapper suite")
}

var _ = Describe("ClientWrapper", func() {
	var (
		events     []*protocol.Event
		exceptions []*protocol.Exception
		requests   []*http.Request
		testServer *httptest.Server
	)
	BeforeEach(func() {
		requests = make([]*http.Request, 0)
		events = make([]*protocol.Event, 0)
		exceptions = make([]*protocol.Exception, 0)
		epsagon.GlobalTracer = &epsagon.MockedEpsagonTracer{
			Events:     &events,
			Exceptions: &exceptions,
			Config:     &epsagon.Config{},
		}
		testServer = httptest.NewServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				requests = append(requests, req)
				res.Write([]byte("body"))
			}))
	})
	AfterEach(func() {
		epsagon.GlobalTracer = nil
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
					Fail("WTF couldn't create request")
				}
				client.Do(req)
				Expect(requests).To(HaveLen(1))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
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
			})
		})
		Context("bad input failing to create request", func() {
			It("Adds event with error code error", func() {
				client := Wrap(http.Client{})
				client.Get(testServer.URL + "balbla")
				Expect(requests).To(HaveLen(0))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
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
				client.Post(
					testServer.URL,
					"application/json",
					strings.NewReader("{\"hello\":\"world\"}"))
				Expect(requests).To(HaveLen(1))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
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
			})
		})
		Context("bad input failing to create request", func() {
			It("Adds event with error code error", func() {
				client := Wrap(http.Client{})
				client.Head(testServer.URL + "blabla")
				Expect(requests).To(HaveLen(0))
				Expect(events).To(HaveLen(1))
				Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
			})
		})
	})
})
