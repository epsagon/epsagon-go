package epsagon

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/epsagon/epsagon-go/epsagon/aws_sdk_v2_factories"
	"github.com/epsagon/epsagon-go/internal"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"log"
	"reflect"
	"time"
)

// WrapAwsV2Service wrap aws service with epsgon
// svc := epsagon.WrapAwsV2Service(dynamodb.New(cfg)).(*dynamodb.Client)
func WrapAwsV2Service(svcClient interface{}, args ...context.Context) interface{} {
	awsClient := reflect.ValueOf(svcClient).Elem().FieldByName("Client").Interface().(*aws.Client)
	awsClient.Handlers.Complete.PushFrontNamed(
		aws.NamedHandler{
			Name: "epsagon-aws-sdk-v2",
			Fn: func(r *aws.Request) {
				currentTracer := internal.ExtractTracer(args)
				completeEventData(r, currentTracer)
			},
		},
	)
	return svcClient
}

func getTimeStampFromRequest(r *aws.Request) float64 {
	return float64(r.Time.UTC().UnixNano()) / float64(time.Millisecond) / float64(time.Nanosecond) / 1000.0
}

func completeEventData(r *aws.Request, currentTracer tracer.Tracer) {
	defer GeneralEpsagonRecover("aws-sdk-go wrapper", "", currentTracer)
	if currentTracer.GetConfig().Debug {
		log.Printf("EPSAGON DEBUG OnComplete current tracer: %+v\n", currentTracer)
		log.Printf("EPSAGON DEBUG OnComplete request response: %+v\n", r.HTTPResponse)
		log.Printf("EPSAGON DEBUG OnComplete request Operation: %+v\n", r.Operation)
		log.Printf("EPSAGON DEBUG OnComplete request Endpoint: %+v\n", r.Endpoint)
		log.Printf("EPSAGON DEBUG OnComplete request Params: %+v\n", r.Params)
		log.Printf("EPSAGON DEBUG OnComplete request Data: %+v\n", r.Data)
	}

	endTime := tracer.GetTimestamp()
	event := protocol.Event{
		Id:        r.RequestID,
		StartTime: getTimeStampFromRequest(r),
		Origin:    "aws-sdk",
		Resource:  extractResourceInformation(r, currentTracer),
	}
	event.Duration = endTime - event.StartTime
	currentTracer.AddEvent(&event)
}

type factory func(*aws.Request, *protocol.Resource, bool, tracer.Tracer)

var awsResourceEventFactories = map[string]factory{
	"s3":       epsagonawsv2factories.S3EventDataFactory,
	"dynamodb": epsagonawsv2factories.DynamodbEventDataFactory,
	"sts":      epsagonawsv2factories.StsDataFactory,
}

func extractResourceInformation(r *aws.Request, currentTracer tracer.Tracer) *protocol.Resource {
	res := protocol.Resource{
		Type:      r.Endpoint.SigningName,
		Operation: r.Operation.Name,
		Metadata:  make(map[string]string),
	}
	factory := awsResourceEventFactories[res.Type]
	if factory != nil {
		factory(r, &res, currentTracer.GetConfig().MetadataOnly, currentTracer)
	} else {
		defaultFactory(r, &res, currentTracer.GetConfig().MetadataOnly, currentTracer)
	}
	return &res
}

func defaultFactory(r *aws.Request, res *protocol.Resource, metadataOnly bool, currentTracer tracer.Tracer) {
	if currentTracer.GetConfig().Debug {
		log.Println("EPSAGON DEBUG:: entering defaultFactory")
	}
	if !metadataOnly {
		extractInterfaceToMetadata(r.Data, res)
		extractInterfaceToMetadata(r.Params, res)
	}
}

func extractInterfaceToMetadata(input interface{}, res *protocol.Resource) {
	var data map[string]interface{}
	rawJSON, err := json.Marshal(input)
	if err != nil {
		log.Printf("EPSAGON DEBUG: Failed to marshal input: %+v\n", input)
		return
	}
	err = json.Unmarshal(rawJSON, &data)
	if err != nil {
		log.Printf("EPSAGON DEBUG: Failed to unmarshal input: %+v\n", rawJSON)
		return
	}
	for key, value := range data {
		res.Metadata[key] = fmt.Sprintf("%v", value)
	}
}
