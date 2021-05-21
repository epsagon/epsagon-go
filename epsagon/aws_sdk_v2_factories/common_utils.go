package epsagonawsv2factories

import (
	"encoding/json"
	"fmt"
	smithyHttp "github.com/aws/smithy-go/transport/http"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"reflect"
	"strconv"
	"time"
)

type (
	AWSClient interface {}
	AWSCall struct {
		RequestID string
		PartitionID string
		Service string
		Region string
		Operation string
		Goos string

		Req *smithyHttp.Request
		Res *smithyHttp.Response
		Endpoint string
		HTTPResponse int
		RequestTime time.Time
		ResponseTime time.Time
		Duration time.Duration
	}
)


type specificOperationHandler func(
	r *AWSCall,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
)

func handleSpecificOperation(
	r *AWSCall,
	res *protocol.Resource,
	metadataOnly bool,
	handlers map[string]specificOperationHandler,
	defaultHandler specificOperationHandler,
	currentTracer tracer.Tracer,
) {

	fmt.Println("HANDLING SPECIFIC s3 OP")
	handler := handlers[res.Operation]
	if handler == nil {
		handler = defaultHandler
	}
	if handler != nil {
		handler(r, res, metadataOnly, currentTracer)
	}
}

func getFieldStringPtr(value reflect.Value, fieldName string) (string, bool) {
	field := value.FieldByName(fieldName)
	if field == (reflect.Value{}) {
		return "", false
	}
	return field.Elem().String(), true
}

func updateMetadataField(data reflect.Value, key string, res *protocol.Resource) {
	value, ok := getFieldStringPtr(data, key)
	if ok {
		res.Metadata[key] = value
	}
}

func updateMetadataFromBytes(
	value reflect.Value, fieldName string, targetKey string, metadata map[string]string) {
	field := value.FieldByName(fieldName)
	if field == (reflect.Value{}) {
		return
	}
	metadata[targetKey] = string(field.Bytes())
}

func updateMetadataFromValue(
	value reflect.Value, fieldName string, targetKey string, metadata map[string]string) {
	fieldValue, ok := getFieldStringPtr(value, fieldName)
	if ok {
		metadata[targetKey] = fieldValue
	}
}

func updateMetadataFromInt64(
	value reflect.Value, fieldName string, targetKey string, metadata map[string]string) {
	field := value.FieldByName(fieldName)
	if field == (reflect.Value{}) {
		return
	}
	metadata[targetKey] = strconv.FormatInt(field.Elem().Int(), 10)
}

func updateMetadataWithFieldToJSON(
	value reflect.Value,
	fieldName string,
	targetKey string,
	metadata map[string]string,
	currentTracer tracer.Tracer,
) {
	field := value.FieldByName(fieldName)
	if field == (reflect.Value{}) {
		return
	}
	stream, err := json.Marshal(field.Interface())
	if err != nil {
		currentTracer.AddExceptionTypeAndMessage("aws-sdk-go", fmt.Sprintf("%v", err))
		return
	}
	metadata[targetKey] = string(stream)
}

func getResourceNameFromField(res *protocol.Resource, value reflect.Value, fieldName string) {
	fieldValue, ok := getFieldStringPtr(value, fieldName)
	if ok {
		res.Name = fieldValue
	}
}
