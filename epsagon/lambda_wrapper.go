package epsagon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
)

var (
	coldStart = true
)

const TimeoutErrorCode protocol.ErrorCode = 3

type genericLambdaHandler func(context.Context, json.RawMessage) (interface{}, error)

// epsagonLambdaWrapper is a generic lambda function type
type epsagonLambdaWrapper struct {
	handler  genericLambdaHandler
	config   *Config
	tracer   tracer.Tracer
	invoked  bool
	invoking bool
	timeout  bool
}

type preInvokeData struct {
	InvocationMetadata map[string]string
	LambdaContext      *lambdacontext.LambdaContext
	StartTime          float64
}

type invocationData struct {
	ExceptionInfo *protocol.Exception
	errorStatus   protocol.ErrorCode
	result        interface{}
	err           error
	thrownError   interface{}
}

func getAWSAccount(lc *lambdacontext.LambdaContext) string {
	arnParts := strings.Split(lc.InvokedFunctionArn, ":")
	if len(arnParts) >= 4 {
		return arnParts[4]
	}
	return ""
}

func createLambdaEvent(preInvokeInfo *preInvokeData) *protocol.Event {
	endTime := tracer.GetTimestamp()
	duration := endTime - preInvokeInfo.StartTime

	return &protocol.Event{
		Id:        preInvokeInfo.LambdaContext.AwsRequestID,
		StartTime: preInvokeInfo.StartTime,
		Resource: &protocol.Resource{
			Name:      lambdacontext.FunctionName,
			Type:      "lambda",
			Operation: "invoke",
			Metadata:  preInvokeInfo.InvocationMetadata,
		},
		Origin:   "runner",
		Duration: duration,
	}
}

func (wrapper *epsagonLambdaWrapper) preInvokeOps(
	ctx context.Context, payload json.RawMessage) (info *preInvokeData) {
	startTime := tracer.GetTimestamp()
	metadata := map[string]string{}
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		lc = &lambdacontext.LambdaContext{}
	}
	defer func() {
		if r := recover(); r != nil {
			wrapper.tracer.AddExceptionTypeAndMessage("LambdaWrapper",
				fmt.Sprintf("preInvokeOps:%+v", r))
			info = &preInvokeData{
				LambdaContext:      lc,
				StartTime:          startTime,
				InvocationMetadata: metadata,
			}
		}
	}()

	metadata = map[string]string{
		"log_stream_name":  lambdacontext.LogStreamName,
		"log_group_name":   lambdacontext.LogGroupName,
		"function_version": lambdacontext.FunctionVersion,
		"memory":           strconv.Itoa(lambdacontext.MemoryLimitInMB),
		"cold_start":       strconv.FormatBool(coldStart),
		"aws_account":      getAWSAccount(lc),
		"region":           os.Getenv("AWS_REGION"),
	}
	coldStart = false

	addLambdaTrigger(payload, wrapper.config.MetadataOnly, triggerFactories, wrapper.tracer)

	return &preInvokeData{
		InvocationMetadata: metadata,
		LambdaContext:      lc,
		StartTime:          startTime,
	}
}

func (wrapper *epsagonLambdaWrapper) postInvokeOps(
	preInvokeInfo *preInvokeData,
	invokeInfo *invocationData) {
	defer func() {
		if r := recover(); r != nil {
			wrapper.tracer.AddExceptionTypeAndMessage("LambdaWrapper", fmt.Sprintf("postInvokeOps:%+v", r))
		}
	}()

	lambdaEvent := createLambdaEvent(preInvokeInfo)
	lambdaEvent.ErrorCode = invokeInfo.errorStatus
	lambdaEvent.Exception = invokeInfo.ExceptionInfo

	if !wrapper.config.MetadataOnly {
		result, err := json.Marshal(invokeInfo.result)
		if err == nil {
			lambdaEvent.Resource.Metadata["return_value"] = string(result)
		} else {
			lambdaEvent.Resource.Metadata["return_value"] = fmt.Sprintf("%+v", invokeInfo.result)
		}
	}

	wrapper.tracer.AddEvent(lambdaEvent)
}

// Invoke calls the wrapper, and creates a tracer for that duration.
func (wrapper *epsagonLambdaWrapper) Invoke(ctx context.Context, payload json.RawMessage) (result interface{}, err error) {
	invokeInfo := &invocationData{}
	wrapper.invoked = false
	wrapper.invoking = false
	defer func() {
		if !wrapper.invoking {
			recover()
		}
		if !wrapper.invoked {
			result, err = wrapper.handler(ctx, payload)
		}
		if invokeInfo.thrownError != nil {
			panic(userError{
				exception: invokeInfo.thrownError,
				stack:     invokeInfo.ExceptionInfo.Traceback,
			})
		}
	}()

	preInvokeInfo := wrapper.preInvokeOps(ctx, payload)
	go wrapper.trackTimeout(ctx, preInvokeInfo)
	wrapper.InvokeClientLambda(ctx, payload, invokeInfo)
	if !wrapper.timeout {
		wrapper.postInvokeOps(preInvokeInfo, invokeInfo)
	}

	return invokeInfo.result, invokeInfo.err
}

func (wrapper *epsagonLambdaWrapper) trackTimeout(ctx context.Context, preInvokeInfo *preInvokeData) {
	deadline, isDeadlineSet := ctx.Deadline()
	if isDeadlineSet {
		thresholdDuration := time.Duration(tracer.GetLambdaTimeoutThresholdMs())
		deadline = deadline.Add(-thresholdDuration * time.Millisecond)
		timeoutChannel := time.After(time.Until(deadline))

		for range timeoutChannel {
			if wrapper.invoking {
				wrapper.timeout = true

				lambdaEvent := createLambdaEvent(preInvokeInfo)
				lambdaEvent.ErrorCode = TimeoutErrorCode

				wrapper.tracer.AddEvent(lambdaEvent)
				wrapper.tracer.Stop()
			}
		}
	}
}

func (wrapper *epsagonLambdaWrapper) InvokeClientLambda(
	ctx context.Context, payload json.RawMessage, invokeInfo *invocationData) {
	defer func() {
		invokeInfo.thrownError = recover()
		if invokeInfo.thrownError != nil {
			invokeInfo.ExceptionInfo = &protocol.Exception{
				Type:      "Runtime Error",
				Message:   fmt.Sprintf("%v", invokeInfo.thrownError),
				Traceback: string(debug.Stack()),
				Time:      tracer.GetTimestamp(),
			}
			invokeInfo.errorStatus = protocol.ErrorCode_EXCEPTION
		}
	}()

	invokeInfo.errorStatus = protocol.ErrorCode_OK
	// calling the actual function:
	wrapper.invoked = true
	wrapper.invoking = true
	result, err := wrapper.handler(ctx, payload)
	wrapper.invoking = false
	if err != nil {
		invokeInfo.errorStatus = protocol.ErrorCode_ERROR
		invokeInfo.ExceptionInfo = &protocol.Exception{
			Type:      "Error Result",
			Message:   err.Error(),
			Traceback: "",
			Time:      tracer.GetTimestamp(),
		}
	}
	invokeInfo.result = result
	invokeInfo.err = err
}

// WrapLambdaHandler wraps a generic wrapper for lambda function with epsagon tracing
func WrapLambdaHandler(config *Config, handler interface{}) interface{} {
	return func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
		wrapperTracer := tracer.CreateGlobalTracer(&config.Config)
		wrapperTracer.Start()

		wrapper := &epsagonLambdaWrapper{
			config:  config,
			handler: makeGenericHandler(handler),
			tracer:  wrapperTracer,
		}

		defer func() {
			if !wrapper.timeout {
				wrapperTracer.Stop()
			}
		}()

		return wrapper.Invoke(ctx, payload)
	}
}
