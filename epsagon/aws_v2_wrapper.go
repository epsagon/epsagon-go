package epsagon

import (
	"context"
	"encoding/json"
	"fmt"
	awsFactories "github.com/epsagon/epsagon-go/epsagon/aws_sdk_v2_factories"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"log"
)

// WrapAwsV2Service wrap aws service with epsgon
// example usage:
// svc := epsagon.WrapAwsV2Service(dynamodb.New(cfg)).(*dynamodb.Client)
func WrapAwsV2Service(svcClient awsFactories.AWSClient, args ...context.Context) awsFactories.AWSClient {
	apiOptions := extractAPIOptions(svcClient)
	awsCall := &awsFactories.AWSCall{}
	currentTracer := ExtractTracer(args)

	awsFactories.AddMiddlewareFuncs(
		apiOptions,
		awsFactories.InitializeMiddleware(awsCall, currentTracer, completeEventData),
		awsFactories.FinalizeMiddleware(awsCall, currentTracer),
	)
	return svcClient
}

func completeEventData(r *awsFactories.AWSCall, currentTracer tracer.Tracer) {
	defer GeneralEpsagonRecover("aws-sdk-go wrapper", "", currentTracer)
	if currentTracer.GetConfig().Debug {
		log.Printf("EPSAGON DEBUG OnComplete current tracer: %+v\n", currentTracer)
		log.Printf("EPSAGON DEBUG OnComplete request response: %+v\n", r.Res.StatusCode)
		log.Printf("EPSAGON DEBUG OnComplete request Operation: %+v\n", r.Operation)
		log.Printf("EPSAGON DEBUG OnComplete request Headers: %+v\n", r.Res.Header)
		log.Printf("EPSAGON DEBUG OnComplete request Input: %+v\n", r.Input)
		log.Printf("EPSAGON DEBUG OnComplete request Output: %+v\n", r.Output)
	}

	r.EndTime = tracer.GetTimestamp()

	event := protocol.Event{
		Id:        r.RequestID,
		StartTime: r.StartTime,
		Origin:    "aws-sdk",
		Resource:  extractResourceInformation(r, currentTracer),
	}
	event.Duration = r.EndTime - r.StartTime
	currentTracer.AddEvent(&event)
}

type factory func(*awsFactories.AWSCall, *protocol.Resource, bool, tracer.Tracer)

var awsResourceEventFactories = map[string]factory{
	"s3":       awsFactories.S3EventDataFactory,
	"dynamodb": awsFactories.DynamodbEventDataFactory,
	"sts":      awsFactories.StsEventDataFactory,
}

func extractResourceInformation(r *awsFactories.AWSCall, currentTracer tracer.Tracer) *protocol.Resource {
	res := protocol.Resource{
		Type:      r.Service,
		Operation: r.Operation,
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

func defaultFactory(r *awsFactories.AWSCall, res *protocol.Resource, metadataOnly bool, currentTracer tracer.Tracer) {
	if currentTracer.GetConfig().Debug {
		log.Println("EPSAGON DEBUG:: entering defaultFactory")
	}
	if !metadataOnly {
		extractInterfaceToMetadata(r.Input, res)
		extractInterfaceToMetadata(r.Output, res)
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
