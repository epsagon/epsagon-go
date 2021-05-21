package epsagon

import (
	"context"
	"encoding/json"
	"fmt"
	awsMiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	smithyMiddlware "github.com/aws/smithy-go/middleware"
	smithyHttp "github.com/aws/smithy-go/transport/http"
	awsFactories "github.com/epsagon/epsagon-go/epsagon/aws_sdk_v2_factories"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"log"
	"reflect"
	"runtime"
	"strings"
	"time"
	"unsafe"
)




// WrapAwsV2Service wrap aws service with epsgon
// svc := epsagon.WrapAwsV2Service(dynamodb.New(cfg)).(*dynamodb.Client)
func WrapAwsV2Service(svcClient awsFactories.AWSClient, args ...context.Context) awsFactories.AWSClient {

	optionsField := reflect.ValueOf(svcClient).Elem().FieldByName("options")
	options := reflect.NewAt(
		optionsField.Type(),
		unsafe.Pointer(optionsField.UnsafeAddr()),
	).Elem().Interface()

	apiOptionsField := reflect.ValueOf(options).FieldByName("APIOptions")
	apiOptions := reflect.NewAt(
		apiOptionsField.Type(),
		unsafe.Pointer(optionsField.UnsafeAddr()),
	).Elem()

	var awsEvent *awsFactories.AWSCall
	currentTracer := ExtractTracer(args)


	apiOptions.Set(reflect.Append(apiOptions, reflect.ValueOf(func (stack *smithyMiddlware.Stack) error {
		return stack.Finalize.Add(
			smithyMiddlware.FinalizeMiddlewareFunc(
				"epsagonFinalizeBefore",
				func(
					ctx context.Context, in smithyMiddlware.FinalizeInput, next smithyMiddlware.FinalizeHandler,
				) (
					out smithyMiddlware.FinalizeOutput, metadata smithyMiddlware.Metadata, err error,
				) {
					out, metadata, err = next.HandleFinalize(ctx, in)

					//req := in.Request.(*smithyHttp.Request)
					//fmt.Println("REQ")
					//fmt.Println(req)
					//b, err := ioutil.ReadAll(req.Body)
					//
					//defer req.Body.Close()
					//if err != nil {
					//	fmt.Println("BIG ERR:")
					//	fmt.Println(err)
					//}
					//fmt.Println(b)

					return
				},
			),
			smithyMiddlware.Before,
		)
	})))

	apiOptions.Set(reflect.Append(apiOptions, reflect.ValueOf(func (stack *smithyMiddlware.Stack) error {
		return stack.Finalize.Add(
			smithyMiddlware.FinalizeMiddlewareFunc(
				"epsagonFinalizeAfter",
				func (
					ctx context.Context, in smithyMiddlware.FinalizeInput, next smithyMiddlware.FinalizeHandler,
				) (
					out smithyMiddlware.FinalizeOutput, metadata smithyMiddlware.Metadata, err error,
				) {
					fmt.Println("--- Custom Finalize Option Called ---")

					out, metadata, err = next.HandleFinalize(ctx, in)
					if err != nil {
						return out, metadata, err
					}

					requestID, _ := awsMiddleware.GetRequestIDMetadata(metadata)
					region := awsMiddleware.GetRegion(ctx)
					operation := awsMiddleware.GetOperationName(ctx)
					service := awsMiddleware.GetSigningName(ctx)
					if len(service) == 0 {
						service = awsMiddleware.GetServiceID(ctx)
					}
					service = strings.ToLower(service)
					_ = awsMiddleware.AddRequestUserAgentMiddleware(stack)
					goos := runtime.GOOS
					requestTime, _ := awsMiddleware.GetServerTime(metadata)
					responseTime, _ := awsMiddleware.GetResponseAt(metadata)
					duration, _ := awsMiddleware.GetAttemptSkew(metadata)
					reqSmithy := in.Request.(*smithyHttp.Request)
					resSmithy := awsMiddleware.GetRawResponse(metadata).(*smithyHttp.Response)

					awsEvent = &awsFactories.AWSCall{
						RequestID: requestID,
						Service: service,
						Region: region,
						Operation: operation,
						Goos: goos,
						Endpoint: reqSmithy.RequestURI,
						Req: reqSmithy,
						Res: resSmithy,

						HTTPResponse: resSmithy.StatusCode,
						RequestTime: requestTime,
						ResponseTime: responseTime,
						Duration: duration,
					}

					completeEventData(awsEvent, currentTracer)
					fmt.Println("--- done Finalize Options ---")
					return out, metadata, nil
				},
			),
			smithyMiddlware.After,
		)
	})))

	return svcClient
}

func getTimeStampFromRequest(r *awsFactories.AWSCall) float64 {
	return float64(r.RequestTime.UTC().UnixNano()) / float64(time.Millisecond) / float64(time.Nanosecond) / 1000.0
}

func completeEventData(r *awsFactories.AWSCall, currentTracer tracer.Tracer) {
	defer GeneralEpsagonRecover("aws-sdk-go wrapper", "", currentTracer)
	if currentTracer.GetConfig().Debug {
		log.Printf("EPSAGON DEBUG OnComplete current tracer: %+v\n", currentTracer)
		log.Printf("EPSAGON DEBUG OnComplete request response: %+v\n", r.HTTPResponse)
		log.Printf("EPSAGON DEBUG OnComplete request Operation: %+v\n", r.Operation)
		log.Printf("EPSAGON DEBUG OnComplete request Endpoint: %+v\n", r.Endpoint)
		//log.Printf("EPSAGON DEBUG OnComplete request Params: %+v\n", r.Params)
		//log.Printf("EPSAGON DEBUG OnComplete request Data: %+v\n", r.Data)
	}

	endTime := tracer.GetTimestamp()
	fmt.Println("endtime - requestime Unix")
	fmt.Println(endTime - float64(r.RequestTime.Unix()))

	fmt.Println("duration seconds - ")
	fmt.Println(r.Duration.Milliseconds())

	fmt.Println("request time unix")
	fmt.Println(float64(r.RequestTime.Unix()))

	fmt.Println("respons time unix")
	fmt.Println(float64(r.ResponseTime.Unix()))

	event := protocol.Event{
		Id:        r.RequestID,
		StartTime: getTimeStampFromRequest(r),
		Origin:    "aws-sdk",
		Resource:  extractResourceInformation(r, currentTracer),
	}
	event.Duration = r.Duration.Seconds()
	currentTracer.AddEvent(&event)
}

type factory func(*awsFactories.AWSCall, *protocol.Resource, bool, tracer.Tracer)

var awsResourceEventFactories = map[string]factory{
	"s3":       awsFactories.S3EventDataFactory,
	"dynamodb": awsFactories.DynamodbEventDataFactory,
	"sts":      awsFactories.StsDataFactory,
}

func extractResourceInformation(input *awsFactories.AWSCall, currentTracer tracer.Tracer) *protocol.Resource {
	res := protocol.Resource{
		Type:      input.Service,
		Operation: input.Operation,
		Metadata:  make(map[string]string),
	}
	factory := awsResourceEventFactories[res.Type]
	if factory != nil {
		factory(input, &res, currentTracer.GetConfig().MetadataOnly, currentTracer)
	} else {
		defaultFactory(input, &res, currentTracer.GetConfig().MetadataOnly, currentTracer)
	}
	return &res
}

func defaultFactory(input interface{}, res *protocol.Resource, metadataOnly bool, currentTracer tracer.Tracer) {
	if currentTracer.GetConfig().Debug {
		log.Println("EPSAGON DEBUG:: entering defaultFactory")
	}
	if !metadataOnly {
		extractInterfaceToMetadata(input, res)
		//extractInterfaceToMetadata(input, res)
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
