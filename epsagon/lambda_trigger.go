package epsagon

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"

	lambdaEvents "github.com/aws/aws-lambda-go/events"
	lambdaContext "github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/google/uuid"
)

type triggerFactory func(event interface{}, metadataOnly bool) *protocol.Event

func mapParametersToString(params map[string]string) string {
	buf, err := json.Marshal(params)
	if err != nil {
		tracer.AddException(&protocol.Exception{
			Type:      "trigger-creation",
			Message:   fmt.Sprintf("Failed to serialize %v", params),
			Traceback: string(debug.Stack()),
			Time:      tracer.GetTimestamp(),
		})
		return ""
	}
	return string(buf)
}

type APIGatewayEventFields struct {
	requestId             string
	headers               map[string]string
	host                  string
	httpMethod            string
	stage                 string
	queryStringParameters map[string]string
	pathParameters        map[string]string
	path                  string
	body                  string
}

func getEventAssertionException(rawEvent interface{}, assertedType string) *protocol.Exception {
	return &protocol.Exception{
		Type:    "trigger-creation",
		Message: fmt.Sprintf("failed to convert rawEvent to %s. %v", assertedType, rawEvent),
		Time:    tracer.GetTimestamp(),
	}
}

func getAPIGatewayTriggerEvent(eventFields *APIGatewayEventFields, metadataOnly bool) *protocol.Event {
	triggerEvent := getAPIGatewayBaseEvent(eventFields)
	if !metadataOnly {
		addAPIGatewayEventMetadata(triggerEvent, eventFields)
	}
	return triggerEvent
}

func getAPIGatewayBaseEvent(eventFields *APIGatewayEventFields) *protocol.Event {
	return &protocol.Event{
		Id:        eventFields.requestId,
		Origin:    "trigger",
		StartTime: tracer.GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      eventFields.host,
			Type:      "api_gateway",
			Operation: eventFields.httpMethod,
			Metadata: map[string]string{
				"stage":                   eventFields.stage,
				"query_string_parameters": mapParametersToString(eventFields.queryStringParameters),
				"path_parameters":         mapParametersToString(eventFields.pathParameters),
				"path":                    eventFields.path,
			},
		},
	}
}

func addAPIGatewayEventMetadata(triggerEvent *protocol.Event, eventFields *APIGatewayEventFields) {
	if bodyJSON, err := json.Marshal(eventFields.body); err != nil {
		tracer.AddException(&protocol.Exception{
			Type:      "trigger-creation",
			Message:   fmt.Sprintf("Failed to serialize body %s", eventFields.body),
			Traceback: string(debug.Stack()),
			Time:      tracer.GetTimestamp(),
		})
		triggerEvent.Resource.Metadata["body"] = ""
	} else {
		triggerEvent.Resource.Metadata["body"] = string(bodyJSON)
	}
	triggerEvent.Resource.Metadata["headers"] = mapParametersToString(eventFields.headers)
}

func triggerAPIGatewayProxyRequest(rawEvent interface{}, metadataOnly bool) *protocol.Event {
	event, ok := rawEvent.(lambdaEvents.APIGatewayProxyRequest)
	if !ok {
		assertionException := getEventAssertionException(rawEvent, "lambdaEvents.APIGatewayProxyRequest")
		tracer.AddException(assertionException)
		return nil
	}
	return getAPIGatewayTriggerEvent(&APIGatewayEventFields{
		requestId:             event.RequestContext.RequestID,
		headers:               event.Headers,
		host:                  event.Headers["Host"],
		httpMethod:            event.HTTPMethod,
		stage:                 event.RequestContext.Stage,
		queryStringParameters: event.QueryStringParameters,
		pathParameters:        event.PathParameters,
		path:                  event.Resource,
		body:                  event.Body,
	}, metadataOnly)
}

func triggerAPIGatewayV2HTTPRequest(rawEvent interface{}, metadataOnly bool) *protocol.Event {
	event, ok := rawEvent.(lambdaEvents.APIGatewayV2HTTPRequest)
	if !ok {
		assertionException := getEventAssertionException(rawEvent, "lambdaEvents.APIGatewayV2HTTPRequest")
		tracer.AddException(assertionException)
		return nil
	}
	return getAPIGatewayTriggerEvent(&APIGatewayEventFields{
		requestId:             event.RequestContext.RequestID,
		headers:               event.Headers,
		host:                  event.Headers["host"],
		httpMethod:            event.RequestContext.HTTP.Method,
		stage:                 event.RequestContext.Stage,
		queryStringParameters: event.QueryStringParameters,
		pathParameters:        event.PathParameters,
		path:                  event.RawPath,
		body:                  event.Body,
	}, metadataOnly)
}

func triggerS3Event(rawEvent interface{}, metadataOnly bool) *protocol.Event {
	event, ok := rawEvent.(lambdaEvents.S3Event)
	if !ok {
		tracer.AddException(&protocol.Exception{
			Type: "trigger-creation",
			Message: fmt.Sprintf(
				"failed to convert rawEvent to lambdaEvents.S3Event %v",
				rawEvent),
			Time: tracer.GetTimestamp(),
		})
		return nil
	}

	triggerEvent := &protocol.Event{
		Id:        fmt.Sprintf("s3-trigger-%s", event.Records[0].ResponseElements["x-amz-request-id"]),
		Origin:    "trigger",
		StartTime: tracer.GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      event.Records[0].S3.Bucket.Name,
			Type:      "s3",
			Operation: event.Records[0].EventName,
			Metadata: map[string]string{
				"region":           event.Records[0].AWSRegion,
				"object_key":       event.Records[0].S3.Object.Key,
				"object_size":      strconv.FormatInt(event.Records[0].S3.Object.Size, 10),
				"object_etag":      event.Records[0].S3.Object.ETag,
				"object_sequencer": event.Records[0].S3.Object.Sequencer,
				"x-amz-request-id": event.Records[0].ResponseElements["x-amz-request-id"],
			},
		},
	}

	return triggerEvent
}

func triggerKinesisEvent(rawEvent interface{}, metadataOnly bool) *protocol.Event {
	event, ok := rawEvent.(lambdaEvents.KinesisEvent)
	if !ok {
		tracer.AddException(&protocol.Exception{
			Type: "trigger-creation",
			Message: fmt.Sprintf(
				"failed to convert rawEvent to lambdaEvents.KinesisEvent %v",
				rawEvent),
			Time: tracer.GetTimestamp(),
		})
		return nil
	}

	eventSourceArnSlice := strings.Split(event.Records[0].EventSourceArn, "/")

	triggerEvent := &protocol.Event{
		Id:        event.Records[0].EventID,
		Origin:    "trigger",
		StartTime: tracer.GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      eventSourceArnSlice[len(eventSourceArnSlice)-1],
			Type:      "kinesis",
			Operation: strings.Replace(event.Records[0].EventName, "aws:kinesis:", "", -1),
			Metadata: map[string]string{
				"region":          event.Records[0].AwsRegion,
				"invoke_identity": event.Records[0].InvokeIdentityArn,
				"sequence_number": event.Records[0].Kinesis.SequenceNumber,
				"partition_key":   event.Records[0].Kinesis.PartitionKey,
			},
		},
	}

	return triggerEvent
}

func triggerSNSEvent(rawEvent interface{}, metadataOnly bool) *protocol.Event {
	event, ok := rawEvent.(lambdaEvents.SNSEvent)
	if !ok {
		tracer.AddException(&protocol.Exception{
			Type: "trigger-creation",
			Message: fmt.Sprintf(
				"failed to convert rawEvent to lambdaEvents.SNSEvent %v",
				rawEvent),
			Time: tracer.GetTimestamp(),
		})
		return nil
	}

	eventSubscriptionArnSlice := strings.Split(event.Records[0].EventSubscriptionArn, ":")

	triggerEvent := &protocol.Event{
		Id:        event.Records[0].SNS.MessageID,
		Origin:    "trigger",
		StartTime: tracer.GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      eventSubscriptionArnSlice[len(eventSubscriptionArnSlice)-2],
			Type:      "sns",
			Operation: event.Records[0].SNS.Type,
			Metadata: map[string]string{
				"Notification Subject": event.Records[0].SNS.Subject,
			},
		},
	}

	if !metadataOnly {
		triggerEvent.Resource.Metadata["Notification Message"] = event.Records[0].SNS.Message
	}

	return triggerEvent
}

func triggerSQSEvent(rawEvent interface{}, metadataOnly bool) *protocol.Event {
	event, ok := rawEvent.(lambdaEvents.SQSEvent)
	if !ok {
		tracer.AddException(&protocol.Exception{
			Type: "trigger-creation",
			Message: fmt.Sprintf(
				"failed to convert rawEvent to lambdaEvents.SQSEvent %v",
				rawEvent),
			Time: tracer.GetTimestamp(),
		})
		return nil
	}

	eventSourceArnSlice := strings.Split(event.Records[0].EventSourceARN, ":")

	triggerEvent := &protocol.Event{
		Id:        event.Records[0].MessageId,
		Origin:    "trigger",
		StartTime: tracer.GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      eventSourceArnSlice[len(eventSourceArnSlice)-1],
			Type:      "sqs",
			Operation: "ReceiveMessage",
			Metadata: map[string]string{
				"MD5 Of Message Body":                 event.Records[0].Md5OfBody,
				"Sender ID":                           event.Records[0].Attributes["SenderId"],
				"Approximate Receive Count":           event.Records[0].Attributes["ApproximateReceiveCount"],
				"Sent Timestamp":                      event.Records[0].Attributes["SentTimestamp"],
				"Approximate First Receive Timestamp": event.Records[0].Attributes["ApproximateFirstReceiveTimestamp"],
			},
		},
	}

	if !metadataOnly {
		triggerEvent.Resource.Metadata["Message Body"] = event.Records[0].Body
		if strings.Contains(event.Records[0].Body, "TopicArn") {
			triggerEvent.Resource.Metadata["SNS Trigger"] = event.Records[0].Body
		}
	}

	return triggerEvent
}

func unmarshalToStringMap(dav map[string]lambdaEvents.DynamoDBAttributeValue) (map[string]string, error) {
	dbAttrMap := make(map[string]*dynamodb.AttributeValue)
	for k, v := range dav {
		var dbAttr dynamodb.AttributeValue
		bytes, marshalErr := v.MarshalJSON()
		if marshalErr != nil {
			return nil, marshalErr
		}
		json.Unmarshal(bytes, &dbAttr)
		dbAttrMap[k] = &dbAttr
	}
	serializedItems := make(map[string]string)
	for k, v := range dbAttrMap {
		serializedItems[k] = v.String()
	}
	return serializedItems, nil
}

func getImageMapBytes(imageMap map[string]lambdaEvents.DynamoDBAttributeValue) ([]byte, error) {
	itemMap, err := unmarshalToStringMap(imageMap)
	if err != nil {
		return nil, err
	}
	itemBytes, jsonError := json.Marshal(itemMap)
	if jsonError != nil {
		return nil, err
	}
	return itemBytes, nil
}

func triggerDynamoDBEvent(rawEvent interface{}, metadataOnly bool) *protocol.Event {
	event, ok := rawEvent.(lambdaEvents.DynamoDBEvent)
	if !ok {
		tracer.AddException(&protocol.Exception{
			Type: "trigger-creation",
			Message: fmt.Sprintf(
				"failed to convert rawEvent to lambdaEvents.DynamoDBEvent %v",
				rawEvent),
			Time: tracer.GetTimestamp(),
		})
		return nil
	}

	eventSourceArnSlice := strings.Split(event.Records[0].EventSourceArn, "/")

	triggerEvent := &protocol.Event{
		Id:        event.Records[0].EventID,
		Origin:    "trigger",
		StartTime: tracer.GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      eventSourceArnSlice[len(eventSourceArnSlice)-3],
			Type:      "dynamodb",
			Operation: event.Records[0].EventName,
			Metadata: map[string]string{
				"region":          event.Records[0].AWSRegion,
				"sequence_number": event.Records[0].Change.SequenceNumber,
				"item_hash":       "",
			},
		},
	}

	itemBytes, err := getImageMapBytes(event.Records[0].Change.NewImage)
	if err != nil {
		return triggerEvent
	}
	oldItemBytes, oldImageErr := getImageMapBytes(event.Records[0].Change.OldImage)
	if oldImageErr != nil {
		return triggerEvent
	}

	h := md5.New()
	h.Write(itemBytes)
	triggerEvent.Resource.Metadata["item_hash"] = hex.EncodeToString(h.Sum(nil))

	if !metadataOnly {
		triggerEvent.Resource.Metadata["New Image"] = string(itemBytes)
		triggerEvent.Resource.Metadata["Old Image"] = string(oldItemBytes)
	}

	return triggerEvent
}

func triggerJSONEvent(rawEvent json.RawMessage, metadataOnly bool) *protocol.Event {
	triggerEvent := &protocol.Event{
		Id:        uuid.New().String(),
		Origin:    "trigger",
		StartTime: tracer.GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      fmt.Sprintf("trigger-%s", lambdaContext.FunctionName),
			Type:      "json",
			Operation: "json",
			Metadata:  map[string]string{},
		},
	}

	if !metadataOnly {
		triggerEvent.Resource.Metadata["data"] = string(rawEvent)
	}

	return triggerEvent
}

type factoryAndType struct {
	EventType reflect.Type
	Factory   triggerFactory
}

var (
	triggerFactories = map[string]factoryAndType{
		"api_gateway": {
			EventType: reflect.TypeOf(lambdaEvents.APIGatewayProxyRequest{}),
			Factory:   triggerAPIGatewayProxyRequest,
		},
		"api_gateway_http2": {
			EventType: reflect.TypeOf(lambdaEvents.APIGatewayV2HTTPRequest{}),
			Factory:   triggerAPIGatewayV2HTTPRequest,
		},
		"aws:s3": {
			EventType: reflect.TypeOf(lambdaEvents.S3Event{}),
			Factory:   triggerS3Event,
		},
		"aws:kinesis": {
			EventType: reflect.TypeOf(lambdaEvents.KinesisEvent{}),
			Factory:   triggerKinesisEvent,
		},
		"aws:sns": {
			EventType: reflect.TypeOf(lambdaEvents.SNSEvent{}),
			Factory:   triggerSNSEvent,
		},
		"aws:sqs": {
			EventType: reflect.TypeOf(lambdaEvents.SQSEvent{}),
			Factory:   triggerSQSEvent,
		},
		"aws:dynamodb": {
			EventType: reflect.TypeOf(lambdaEvents.DynamoDBEvent{}),
			Factory:   triggerDynamoDBEvent,
		},
	}
)

func decodeAndUnpackEvent(
	payload json.RawMessage,
	eventType reflect.Type,
	factory triggerFactory,
	metadataOnly bool,
) *protocol.Event {

	event := reflect.New(eventType)
	decoder := json.NewDecoder(bytes.NewReader(payload))

	if err := decoder.Decode(event.Interface()); err != nil {
		return nil
	}
	return factory(event.Elem().Interface(), metadataOnly)
}

type recordField struct {
	EventSource string
}

type httpDescription struct {
	Method string
}

type requestContext struct {
	APIID string
	HTTP  httpDescription
}

type interestingFields struct {
	Records        []recordField
	HTTPMethod     string
	Context        map[string]interface{}
	MethodArn      string
	Source         string
	RequestContext requestContext
}

func guessTriggerSource(payload json.RawMessage) string {
	var rawEvent interestingFields
	err := json.Unmarshal(payload, &rawEvent)
	if err != nil {
		tracer.AddException(&protocol.Exception{
			Type:      "trigger-identification",
			Message:   fmt.Sprintf("Failed to unmarshal json %v\n", err),
			Traceback: string(debug.Stack()),
			Time:      tracer.GetTimestamp(),
		})
		return ""
	}
	triggerSource := "json"
	if len(rawEvent.Records) > 0 {
		triggerSource = rawEvent.Records[0].EventSource
	} else if len(rawEvent.HTTPMethod) > 0 {
		triggerSource = "api_gateway"
	} else if _, ok := rawEvent.Context["http-method"]; ok {
		triggerSource = "api_gateway_no_proxy"
	} else if len(rawEvent.RequestContext.APIID) > 0 && len(rawEvent.RequestContext.HTTP.Method) > 0 {
		triggerSource = "api_gateway_http2"
	} else if len(rawEvent.Source) > 0 {
		sourceSlice := strings.Split(rawEvent.Source, ".")
		triggerSource = sourceSlice[len(sourceSlice)-1]
	}
	return triggerSource
}

func addLambdaTrigger(
	payload json.RawMessage,
	metadataOnly bool,
	triggerFactories map[string]factoryAndType,
	currentTracer tracer.Tracer,
) {
	var triggerEvent *protocol.Event

	triggerSource := guessTriggerSource(payload)

	if triggerSource == "json" {
		triggerEvent = triggerJSONEvent(payload, metadataOnly)
	} else if triggerSource == "api_gateway_no_proxy" {
		// currently not supported, needs to extract data from json
	} else {
		factoryStruct, found := triggerFactories[triggerSource]
		if found {
			triggerEvent = decodeAndUnpackEvent(
				payload, factoryStruct.EventType, factoryStruct.Factory, metadataOnly)
		}
	}

	// If a trigger was found
	if triggerEvent != nil {
		currentTracer.AddEvent(triggerEvent)
	}
}
