package epsagonawswrapper

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/epsagon/epsagon-go/protocol"
	"reflect"
)

type specificOperationHandler func(r *request.Request, res *protocol.Resource, metadataOnly bool)

func getFieldStringPtr(value reflect.Value, fieldName string) (string, bool) {
	field := value.FieldByName(fieldName)
	if field == (reflect.Value{}) {
		return "", false
	}
	return field.Elem().String(), true
}

func updateMetadataFromValue(
	value reflect.Value, fieldName string, targetKey string, metadata map[string]string) {
	fieldValue, ok := getFieldStringPtr(value, fieldName)
	if ok {
		metadata[targetKey] = fieldValue
	}
}
