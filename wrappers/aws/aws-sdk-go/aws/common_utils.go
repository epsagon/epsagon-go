package epsagonawswrapper

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
)

type specificOperationHandler func(r *request.Request, res *protocol.Resource, metadataOnly bool, currenttracer tracer.Tracer)

func handleSpecificOperation(
	r *request.Request,
	res *protocol.Resource,
	metadataOnly bool,
	handlers map[string]specificOperationHandler,
	defaultHandler specificOperationHandler,
	currentTracer tracer.Tracer,
) {
	handler := handlers[res.Operation]
	if handler == nil {
		handler = defaultHandler
	}
	if handler != nil {
		handler(r, res, metadataOnly, currentTracer)
	}
}

func isValueZero(value reflect.Value) bool {
	return !value.IsValid() || value.IsZero()
}

func getFieldStringPtr(value reflect.Value, fieldName string) (string, bool) {
	field := value.FieldByName(fieldName)
	if isValueZero(field) {
		return "", false
	}
	return field.Elem().String(), true
}

func updateMetadataFromBytes(
	value reflect.Value, fieldName string, targetKey string, metadata map[string]string) {
	field := value.FieldByName(fieldName)
	if isValueZero(field) {
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
	if isValueZero(field) {
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
	if isValueZero(field) {
		return
	}
	stream, err := json.Marshal(field.Interface())
	if err != nil {
		currentTracer.AddExceptionTypeAndMessage("aws-sdk-go", fmt.Sprintf("%v", err))
		return
	}
	metadata[targetKey] = string(stream)
}
