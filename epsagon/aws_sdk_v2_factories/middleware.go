
package epsagonawsv2factories

import (
	"context"
	awsMiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	smithyMiddleware "github.com/aws/smithy-go/middleware"
	smithyHttp "github.com/aws/smithy-go/transport/http"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"reflect"
	"strings"
)

type MiddlewareFunc func(stack *smithyMiddleware.Stack) error


func AddMiddlewareFuncs(apiOptions reflect.Value, option ...MiddlewareFunc) reflect.Value {
	if ! apiOptions.CanSet() {
		return apiOptions
	}

	for _, opt := range option {
		apiOptions.Set(reflect.Append(
			apiOptions,
			reflect.ValueOf(opt),
		))
	}

	return apiOptions
}


func InitializeMiddleware (
	awsCall *AWSCall, currentTracer tracer.Tracer, completeEvent func(*AWSCall, tracer.Tracer),
	) func(stack *smithyMiddleware.Stack) error {

	return func (stack *smithyMiddleware.Stack) error {
		return stack.Initialize.Add(
			smithyMiddleware.InitializeMiddlewareFunc(
				"epsagonInitialize",
				func (
					ctx context.Context, in smithyMiddleware.InitializeInput, next smithyMiddleware.InitializeHandler,
				) (
					out smithyMiddleware.InitializeOutput, metadata smithyMiddleware.Metadata, err error,
				) {

					awsCall.StartTime = tracer.GetTimestamp()
					awsCall.Input = in.Parameters

					out, metadata, err = next.HandleInitialize(ctx, in)
					if err != nil {
						tracer.AddException(&protocol.Exception{
							Type:                 "aws-sdk-go-v2",
							Message:              err.Error(),
							Traceback:            "",
							Time:                 tracer.GetTimestamp(),
						})
						return out, metadata, err
					}

					if out != (smithyMiddleware.InitializeOutput{}) {
						awsCall.Output = out.Result
					}

					completeEvent(awsCall, currentTracer)

					return out, metadata, nil
				},
			),
			smithyMiddleware.After,
		)
	}
}



func FinalizeMiddleware (awsCall *AWSCall, currentTracer tracer.Tracer) func(stack *smithyMiddleware.Stack) error {
	return func (stack *smithyMiddleware.Stack) error {
		return stack.Finalize.Add(
			smithyMiddleware.FinalizeMiddlewareFunc(
				"epsagonFinalize",
				func (
					ctx context.Context, in smithyMiddleware.FinalizeInput, next smithyMiddleware.FinalizeHandler,
				) (
					out smithyMiddleware.FinalizeOutput, metadata smithyMiddleware.Metadata, err error,
				) {
					out, metadata, err = next.HandleFinalize(ctx, in)


					if err != nil {
						tracer.AddException(&protocol.Exception{
							Type:                 "go-aws-sdk-v2",
							Message:              err.Error(),
							Traceback:            "",
							Time:                 tracer.GetTimestamp(),
						})
					}

					requestID, ok := awsMiddleware.GetRequestIDMetadata(metadata)
					if ok {
						awsCall.RequestID = requestID
					}

					awsCall.Region = awsMiddleware.GetRegion(ctx)
					awsCall.Operation = awsMiddleware.GetOperationName(ctx)

					service := awsMiddleware.GetSigningName(ctx)
					if len(service) == 0 {
						service = awsMiddleware.GetServiceID(ctx)
					}

					awsCall.Service = strings.ToLower(service)
					_ = awsMiddleware.AddRequestUserAgentMiddleware(stack)

					awsCall.Req = in.Request.(*smithyHttp.Request)
					awsCall.Res = awsMiddleware.GetRawResponse(metadata).(*smithyHttp.Response)

					return out, metadata, nil
				},
			),
			smithyMiddleware.After,
		)
	}
}
