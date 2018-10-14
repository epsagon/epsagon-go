package epsagon

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/epsagon/epsagon-go/protocol"
	"os"
	"strconv"
	"time"
)

var (
	coldStart = true
)

type genericHandler func(context.Context, json.RawMessage) (interface{}, error)

// epsagonLambdaWrapper is a generic lambda function type
type epsagonLambdaWrapper struct {
	handler         genericHandler
	token           string
	applicationName string
	collectorURL    string
}

func (handler *epsagonLambdaWrapper) createTracer() {
	CreateTracer(
		handler.applicationName,
		handler.token,
		handler.collectorURL,
	)
}

// Invoke calls the handler, and creates a tracer for that duration.
func (handler epsagonLambdaWrapper) Invoke(ctx context.Context, payload json.RawMessage) (interface{}, error) {
	handler.createTracer()
	defer StopTracer()
	errorStatus := protocol.ErrorCode_OK

	startTime := float64(time.Now().UTC().Unix())

	lc, _ := lambdacontext.FromContext(ctx)

	metadata := map[string]string{
		"log_stream_name":  lambdacontext.LogStreamName,
		"log_group_name":   lambdacontext.LogGroupName,
		"function_version": lambdacontext.FunctionVersion,
		"memory":           strconv.Itoa(lambdacontext.MemoryLimitInMB),
		"cold_start":       strconv.FormatBool(coldStart),
		"region":           os.Getenv("AWS_REGION"),
	}
	coldStart = false

	// calling the actual function:
	result, err := handler.handler(ctx, payload)
	if err != nil {
		errorStatus = protocol.ErrorCode_ERROR
	}

	endTime := float64(time.Now().UTC().Unix())
	duration := endTime - startTime
	AddEvent(&protocol.Event{
		Id:        lc.AwsRequestID,
		StartTime: startTime,
		Resource: &protocol.Resource{
			Name:      lambdacontext.FunctionName,
			Type:      "lambda",
			Operation: "invoke",
			Metadata:  metadata,
		},
		Origin:    "runner",
		Duration:  duration,
		ErrorCode: errorStatus,
	})
	return result, err
}

// WrapLambdaHandler wraps a generic handler for lambda function with epsagon tracing
func WrapLambdaHandler(appName, token, collectorURL string, handler interface{}) interface{} {
	return func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
		wrapper := &epsagonLambdaWrapper{
			applicationName: appName,
			token:           token,
			collectorURL:    collectorURL,
			handler:         makeGenericHandler(handler),
		}
		return wrapper.Invoke(ctx, payload)
	}
}
