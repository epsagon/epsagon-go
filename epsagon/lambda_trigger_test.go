package epsagon

import (
	// "fmt"
	"encoding/json"
	lambdaEvents "github.com/aws/aws-lambda-go/events"
	protocol "github.com/epsagon/epsagon-go/protocol"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

func TestEpsagonLambdaTrigger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "epsagon trigger suite")
}

type inventedEvent struct {
	Name        string
	Job         string
	DateOfBirth time.Time
}

var (
	exampleAPIGateWay = lambdaEvents.APIGatewayProxyRequest{
		Resource:        "test-resource",
		Path:            "/hello",
		HTTPMethod:      "GET",
		Body:            "<b>hello world</b>",
		IsBase64Encoded: false,
		Headers: map[string]string{
			"hello": "world",
		},
		StageVariables: map[string]string{
			"hello": "world",
		},
		PathParameters: map[string]string{
			"hello": "world",
		},
		QueryStringParameters: map[string]string{
			"hello": "world",
		},
	}
	exampleInventedEvent = inventedEvent{
		Name:        "Erez Freiberger",
		Job:         "Software Engineer",
		DateOfBirth: time.Now(),
	}
)

var _ = Describe("epsagon trigger suite", func() {
	Describe("addLambdaTrigger", func() {
		var (
			events     []*protocol.Event
			exceptions []*protocol.Exception
		)
		BeforeEach(func() {
			events = make([]*protocol.Event, 0)
			exceptions = make([]*protocol.Exception, 0)
			globalTracer = &mockedEpsagonTracer{
				events:     &events,
				exceptions: &exceptions,
			}
		})

		Context("Handling of known trigger", func() {
			It("Identifies the first known handler", func() {
				exampleJSON, err := json.Marshal(exampleAPIGateWay)
				if err != nil {
					Fail("Failed to marshal json")
				}
				addLambdaTrigger(json.RawMessage(exampleJSON), false, triggerFactories)
				Expect(len(events)).To(BeNumerically("==", 1))
				Expect(events[0].Resource.Type).To(Equal("api_gateway"))
			})
		})

		Context("Handling of known trigger with extra fields", func() {
			It("Identifies the first known handler", func() {
				exampleJSON, err := json.Marshal(exampleAPIGateWay)
				if err != nil {
					Fail("Failed to marshal json")
				}
				var rawEvent map[string]interface{}
				err = json.Unmarshal(exampleJSON, &rawEvent)
				if err != nil {
					Fail("Failed to unmarshal json into rawEvent map[string]interface{}")
				}
				rawEvent["ExtraFields"] = "BOOOM"
				exampleJSON, err = json.Marshal(rawEvent)
				if err != nil {
					Fail("Failed to marshal json from rawEvent")
				}
				addLambdaTrigger(json.RawMessage(exampleJSON), false, triggerFactories)
				Expect(len(events)).To(BeNumerically("==", 1))
				Expect(events[0].Resource.Type).To(Equal("api_gateway"))
			})
		})

		Context("Handling of unknown trigger", func() {
			It("Adds a JSON Event", func() {
				exampleJSON, err := json.Marshal(exampleInventedEvent)
				if err != nil {
					Fail("Failed to marshal json")
				}
				addLambdaTrigger(json.RawMessage(exampleJSON), false, triggerFactories)
				Expect(len(events)).To(BeNumerically("==", 1))
				Expect(events[0].Resource.Type).To(Equal("json"))
			})
		})
	})
})
