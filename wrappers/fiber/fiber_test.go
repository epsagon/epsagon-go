package epsagonfiber

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"

	"github.com/gofiber/fiber/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHandlerWrapper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fiber Wrapper")
}

const SanityPath = "/test"
const SanityHTTPMethod = "GET"
const ResponseData = "Hello world"

func verifyResponseSuccess(response *http.Response, err error) {
	Expect(err).To(BeNil())
	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	Expect(err).To(BeNil())
	responseString := string(responseData)
	Expect(responseString).To(Equal(ResponseData))
}

var _ = Describe("fiber_middleware", func() {
	Describe("FiberEpsagonMiddleware", func() {
		var (
			events     []*protocol.Event
			exceptions []*protocol.Exception
			config     *epsagon.Config
			app        *fiber.App
			called     bool
			request    *http.Request
		)
		BeforeEach(func() {
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
			request = httptest.NewRequest(SanityHTTPMethod, SanityPath, nil)
			app = fiber.New()
			called = false
			epsagonMiddleware := &FiberEpsagonMiddleware{
				Config: config,
			}
			app.Use(epsagonMiddleware.HandlerFunc())
			app.Get(SanityPath, func(c *fiber.Ctx) error {
				called = true
				return c.SendString(ResponseData)
			})
		})
		Context("Happy Flows", func() {
			It("calls the original handler", func() {
				_, err := app.Test(request)
				Expect(err).To(BeNil())
				Expect(called).To(Equal(true))
			})
			It("validates the original handler response", func() {
				resp, err := app.Test(request)
				verifyResponseSuccess(resp, err)
				Expect(called).To(Equal(true))
			})
			It("creates a runner & trigger events", func() {
				eventsRecievedChan := make(chan bool)
				tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
					Events:            &events,
					Exceptions:        &exceptions,
					Labels:            make(map[string]interface{}),
					Config:            &config.Config,
					DelayAddEvent:     true,
					DelayedEventsChan: eventsRecievedChan,
				}
				resp, err := app.Test(request)
				verifyResponseSuccess(resp, err)
				Expect(called).To(Equal(true))
				timer := time.NewTimer(time.Second * 10)
				for eventsRecieved := 0; eventsRecieved < 2; {
					select {
					case <-eventsRecievedChan:
						eventsRecieved++
					case <-timer.C:
						// timeout - events should have been recieved
						Expect(false).To(Equal(true))
					}
				}
				Expect(len(events)).To(Equal(2))
				var runnerEvent *protocol.Event
				for _, event := range events {
					if event.Origin == "runner" {
						runnerEvent = event
					}
				}
				Expect(runnerEvent).NotTo(Equal(nil))
				Expect(runnerEvent.Resource.Type).To(Equal("fiber"))
				Expect(runnerEvent.Resource.Name).To(Equal("/test"))
			})
			It("Validates runner event", func() {
				eventsRecievedChan := make(chan bool)
				tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
					Events:            &events,
					Exceptions:        &exceptions,
					Labels:            make(map[string]interface{}),
					Config:            &config.Config,
					DelayAddEvent:     true,
					DelayedEventsChan: eventsRecievedChan,
				}
				bodyString := "hello world"
				postRequest := httptest.NewRequest(
					"POST",
					"/test?hello=world&good=bye",
					strings.NewReader(bodyString))
				app.Post(SanityPath, func(c *fiber.Ctx) error {
					requestBody := c.Body()
					if string(requestBody) != bodyString {
						panic("unexpected request body")
					}
					called = true
					return c.SendString(ResponseData)
				})

				resp, err := app.Test(postRequest)
				verifyResponseSuccess(resp, err)
				Expect(called).To(Equal(true))
				timer := time.NewTimer(time.Second * 10)
				for eventsRecieved := 0; eventsRecieved < 2; {
					select {
					case <-eventsRecievedChan:
						eventsRecieved++
					case <-timer.C:
						// timeout - events should have been recieved
						Expect(false).To(Equal(true))
					}
				}
				Expect(len(events)).To(Equal(2))
				var triggerEvent *protocol.Event
				for _, event := range events {
					if event.Origin == "trigger" {
						triggerEvent = event
					}
				}
				Expect(triggerEvent).NotTo(Equal(nil))
				Expect(triggerEvent.Resource.Name).To(Equal("example.com"))
				Expect(triggerEvent.Resource.Type).To(Equal("http"))
				Expect(triggerEvent.Resource.Operation).To(Equal("POST"))
				expectedQuery, _ := json.Marshal(map[string]string{
					"hello": "world", "good": "bye"})
				Expect(triggerEvent.Resource.Metadata["query_string_params"]).To(
					Equal(string(expectedQuery)))
				Expect(triggerEvent.Resource.Metadata["path"]).To(
					Equal("/test"))
				Expect(triggerEvent.Resource.Metadata["request_body"]).To(
					Equal(bodyString))
				Expect(triggerEvent.Resource.Metadata["response_body"]).To(Equal(ResponseData))
				Expect(triggerEvent.Resource.Metadata["status_code"]).To(
					Equal("200"))
			})
			It("Validates runner event, metadataOnly is true", func() {
				eventsRecievedChan := make(chan bool)
				config := &epsagon.Config{Config: tracer.Config{
					Disable:      true,
					TestMode:     true,
					MetadataOnly: true,
				}}
				tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
					Events:            &events,
					Exceptions:        &exceptions,
					Labels:            make(map[string]interface{}),
					Config:            &config.Config,
					DelayAddEvent:     true,
					DelayedEventsChan: eventsRecievedChan,
				}
				bodyString := "hello world"
				postRequest := httptest.NewRequest(
					"POST",
					"/test?hello=world&good=bye",
					strings.NewReader(bodyString))
				app.Post(SanityPath, func(c *fiber.Ctx) error {
					requestBody := c.Body()
					if string(requestBody) != bodyString {
						panic("unexpected request body")
					}
					called = true
					return c.SendString(ResponseData)
				})
				resp, err := app.Test(postRequest)
				verifyResponseSuccess(resp, err)
				Expect(called).To(Equal(true))
				timer := time.NewTimer(time.Second * 10)
				for eventsRecieved := 0; eventsRecieved < 2; {
					select {
					case <-eventsRecievedChan:
						eventsRecieved++
					case <-timer.C:
						// timeout - events should have been recieved
						Expect(false).To(Equal(true))
					}
				}
				Expect(len(events)).To(Equal(2))
				var triggerEvent *protocol.Event
				for _, event := range events {
					if event.Origin == "trigger" {
						triggerEvent = event
					}
				}
				Expect(triggerEvent).NotTo(Equal(nil))
				Expect(triggerEvent.Resource.Name).To(Equal("example.com"))
				Expect(triggerEvent.Resource.Type).To(Equal("http"))
				Expect(triggerEvent.Resource.Operation).To(Equal("POST"))
				Expect(triggerEvent.Resource.Metadata).To(Not(ContainElement("query_string_params")))
				Expect(triggerEvent.Resource.Metadata).To(Not(ContainElement("request_body")))
				Expect(triggerEvent.Resource.Metadata).To(Not(ContainElement("response_body")))
				Expect(triggerEvent.Resource.Metadata["path"]).To(
					Equal("/test"))
				Expect(triggerEvent.Resource.Metadata["status_code"]).To(
					Equal("200"))

			})
		})
	})
})
