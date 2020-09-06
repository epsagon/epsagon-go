package epsagonawswrapper

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"

	awsmetadata "github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
)

const (
	TestPanic = "test panic"
)

func TestEpsagonAWSWrappers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "epsagon aws sdk wrapper suite")
}

var _ = Describe("epsagon aws sdk wrapper suite", func() {
	Describe("WrapSession", func() {
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
				Config:     &tracer.Config{},
			}
		})
		AfterEach(func() {
			tracer.GlobalTracer = nil
		})

		Context("use of known aws operation", func() {
			It("adds an event with the correct data", func() {
				sess := WrapSession(session.Must(session.NewSession()))
				svcSQS := sqs.New(sess)
				sqsQueueName := "QueueName"
				_, _ = svcSQS.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: &sqsQueueName})
				Expect(len(events)).To(Equal(1))
				Expect(len(exceptions)).To(Equal(0))
				Expect(events[0].Resource.Type).To(Equal("sqs"))
				Expect(events[0].Resource.Operation).To(Equal("GetQueueUrl"))
				Expect(events[0].Resource.Name).To(Equal(sqsQueueName))
			})
		})
		Context("use of unknown operation", func() {
			It("adds an event with the available data", func() {
				// playing with internal structure to make all operations "unknown"
				original := awsResourceEventFactories
				defer func() { awsResourceEventFactories = original }()
				awsResourceEventFactories = map[string]factory{}

				sess := WrapSession(session.Must(session.NewSession()))
				svcSQS := sqs.New(sess)
				sqsQueueName := "QueueName"
				_, _ = svcSQS.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: &sqsQueueName})
				Expect(len(events)).To(Equal(1))
				Expect(len(exceptions)).To(Equal(0))
				Expect(events[0].Resource.Type).To(Equal("sqs"))
				Expect(events[0].Resource.Operation).To(Equal("GetQueueUrl"))
			})
		})
		Context("when errors occures in an internal mechanism", func() {
			It("Recovers and sends exception to tracer", func() {
				// playing with internal structure to make all operations "unknown"
				original := awsResourceEventFactories
				defer func() { awsResourceEventFactories = original }()
				awsResourceEventFactories = map[string]factory{
					"sqs": myFault,
				}
				sess := WrapSession(session.Must(session.NewSession()))
				svcSQS := sqs.New(sess)
				sqsQueueName := "QueueName"
				_, _ = svcSQS.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: &sqsQueueName})
				Expect(len(events)).To(Equal(0))
				Expect(len(exceptions)).To(Equal(1))
				Expect(exceptions[0].Message).To(Equal(fmt.Sprintf(":%s", TestPanic)))
			})
		})
	})
	// Describe("completeEventData", func() {
	// })
	Describe("defaultFactory", func() {
		Context("sanity with simple dynamodb data", func() {
			var (
				req       request.Request
				world     string
				tableName string
				param     dynamodb.GetItemInput
				data      dynamodb.GetItemOutput
			)
			BeforeEach(func() {
				world = "world"
				tableName = "erez-table"
				param = dynamodb.GetItemInput{
					ExpressionAttributeNames: map[string]*string{
						"hello": &world,
					},
					TableName: &tableName,
				}
				data = dynamodb.GetItemOutput{
					ConsumedCapacity: &dynamodb.ConsumedCapacity{
						TableName: &tableName,
					},
				}
				req = request.Request{
					Data:   &data,
					Params: &param,
				}
				tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
					Config: &tracer.Config{},
				}
			})
			It("Extracts basic data", func() {
				res := protocol.Resource{
					Metadata: make(map[string]string),
				}
				defaultFactory(&req, &res, false, tracer.GlobalTracer)
				Expect(res.Metadata["TableName"]).To(Equal(tableName))
				Expect(res.Metadata["ExpressionAttributeNames"]).To(
					Equal(fmt.Sprintf("%v", map[string]string{"hello": "world"})))
				Expect(res.Metadata["ConsumedCapacity"]).To(
					ContainSubstring(tableName))
			})
			It("Wont add data if MetadataOnly is set to true", func() {
				res := protocol.Resource{
					Metadata: make(map[string]string),
				}
				defaultFactory(&req, &res, true, tracer.GlobalTracer)
				Expect(res.Metadata["TableName"]).To(BeZero())
				Expect(res.Metadata["ExpressionAttributeNames"]).To(BeZero())
				Expect(res.Metadata["ConsumedCapacity"]).To(BeZero())
			})
		})
	})
	Describe("extractResourceInformation", func() {
		var (
			req request.Request
		)
		BeforeEach(func() {
			req = request.Request{
				ClientInfo: awsmetadata.ClientInfo{
					ServiceName: "sqs",
				},
				Operation: &request.Operation{
					Name: "SendMessage",
				},
				Data:   map[string]string{"hello": "world"},
				Params: map[string]string{"params": "output"},
			}
			tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
				Config: &tracer.Config{},
			}
		})

		Context("unrecognized input", func() {
			It("calls default factory on unknown service", func() {
				req.ClientInfo.ServiceName = "Non Existant Service"
				res := extractResourceInformation(&req, tracer.GlobalTracer)
				Expect(res.Metadata["hello"]).To(Equal("world"))
				Expect(res.Metadata["params"]).To(Equal("output"))
			})
		})
		Context("recognized input", func() {
			It("calls the correct data factory", func() {
				url := "prefix/QueueName"
				req.Params = &sqs.SendMessageInput{
					QueueUrl: &url,
				}
				messageID := "id"
				mD5OfMessageBody := "md4"
				req.Data = &sqs.SendMessageOutput{
					MessageId:        &messageID,
					MD5OfMessageBody: &mD5OfMessageBody,
				}
				res := extractResourceInformation(&req, tracer.GlobalTracer)
				Expect(res.Name).To(Equal("QueueName"))
			})
		})
	})
})

func myFault(*request.Request, *protocol.Resource, bool, tracer.Tracer) {
	panic(TestPanic)
}
