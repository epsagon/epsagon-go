package epsagon

import (
	// "fmt"
	"encoding/json"
	"time"

	lambdaEvents "github.com/aws/aws-lambda-go/events"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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
	exampleAPIGatewayV2HTTP = lambdaEvents.APIGatewayV2HTTPRequest{
		RawPath: "/hello",
		RequestContext: lambdaEvents.APIGatewayV2HTTPRequestContext{
			APIID: "test-api",
			HTTP: lambdaEvents.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "GET",
			},
		},
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
	exampleDDB = lambdaEvents.DynamoDBEvent{
		Records: []lambdaEvents.DynamoDBEventRecord{
			lambdaEvents.DynamoDBEventRecord{
				AWSRegion:      "us-east-1",
				EventSourceArn: "test/1/2",
				EventSource:    "aws:dynamodb",
				EventName:      "PutItem",
				Change: lambdaEvents.DynamoDBStreamRecord{
					SequenceNumber: "test_sequence_number",
					NewImage: map[string]lambdaEvents.DynamoDBAttributeValue{
						"test2": lambdaEvents.NewStringAttribute("2"),
					},
					OldImage: map[string]lambdaEvents.DynamoDBAttributeValue{
						"test1": lambdaEvents.NewStringAttribute("1"),
					},
				},
			},
		},
	}
	exampleInventedEvent = inventedEvent{
		Name:        "Erez Freiberger",
		Job:         "Software Engineer",
		DateOfBirth: time.Now(),
	}
)

func verifyLabelValue(key string, value string, labelsMap map[string]string) {
	labelValue, ok := labelsMap[key]
	Expect(ok).To(BeTrue())
	Expect(labelValue).To(Equal(value))
}

var _ = Describe("epsagon trigger suite", func() {
	Describe("addLambdaTrigger", func() {
		var (
			events     []*protocol.Event
			exceptions []*protocol.Exception
		)
		BeforeEach(func() {
			events = make([]*protocol.Event, 0)
			exceptions = make([]*protocol.Exception, 0)
			tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
				Events:     &events,
				Exceptions: &exceptions,
			}
		})

		Context("Handling of known trigger - API Gateway", func() {
			It("Identifies the first known handler, API Gateway - REST", func() {
				exampleJSON, err := json.Marshal(exampleAPIGateWay)
				if err != nil {
					Fail("Failed to marshal json")
				}
				addLambdaTrigger(json.RawMessage(exampleJSON), false, triggerFactories, tracer.GlobalTracer)
				Expect(len(events)).To(BeNumerically("==", 1))
				Expect(events[0].Resource.Type).To(Equal("api_gateway"))
			})

			It("Identifies the first known handler, API Gateway - HTTP", func() {
				exampleJSON, err := json.Marshal(exampleAPIGatewayV2HTTP)
				if err != nil {
					Fail("Failed to marshal json")
				}
				addLambdaTrigger(json.RawMessage(exampleJSON), false, triggerFactories, tracer.GlobalTracer)
				Expect(len(events)).To(BeNumerically("==", 1))
				Expect(events[0].Resource.Type).To(Equal("api_gateway"))
			})
		})
		Context("Handling of known trigger - DynamoDB", func() {
			It("Identifies the first known handler, DynamoDB", func() {
				exampleJSON, err := json.Marshal(exampleDDB)
				if err != nil {
					Fail("Failed to marshal json")
				}
				addLambdaTrigger(json.RawMessage(exampleJSON), false, triggerFactories, tracer.GlobalTracer)
				Expect(len(events)).To(BeNumerically("==", 1))
				Expect(events[0].Resource.Type).To(Equal("dynamodb"))
				verifyLabelValue("region", "us-east-1", events[0].Resource.Metadata)
				verifyLabelValue(
					"New Image",
					"{\"test2\":\"{\\n  S: \\\"2\\\"\\n}\"}",
					events[0].Resource.Metadata,
				)
				verifyLabelValue(
					"Old Image",
					"{\"test1\":\"{\\n  S: \\\"1\\\"\\n}\"}",
					events[0].Resource.Metadata,
				)
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
				addLambdaTrigger(json.RawMessage(exampleJSON), false, triggerFactories, tracer.GlobalTracer)
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
				addLambdaTrigger(json.RawMessage(exampleJSON), false, triggerFactories, tracer.GlobalTracer)
				Expect(len(events)).To(BeNumerically("==", 1))
				Expect(events[0].Resource.Type).To(Equal("json"))
			})
		})
	})
})
