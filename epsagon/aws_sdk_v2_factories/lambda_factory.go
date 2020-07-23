package epsagonawsv2factories

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"io"
	"reflect"
)

// LambdaEventDataFactory creates an Epsagon Resource from aws.Reqeust to lambda
func LambdaEventDataFactory(r *aws.Request, res *protocol.Resource, metadataOnly bool) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	functionName, ok := getFieldStringPtr(inputValue, "FunctionName")
	if ok {
		res.Name = functionName
	}
	if metadataOnly {
		return
	}
	updateMetadataFromBytes(inputValue, "Payload", "payload", res.Metadata)
	invokeArgsField := inputValue.FieldByName("InvokeArgs")
	if invokeArgsField == (reflect.Value{}) {
		return
	}
	invokeArgsReader := invokeArgsField.Interface().(io.ReadSeeker)
	invokeArgsBytes := make([]byte, 100)

	initialOffset, err := invokeArgsReader.Seek(int64(0), io.SeekStart)
	if err != nil {
		tracer.AddExceptionTypeAndMessage("aws-sdk-go",
			fmt.Sprintf("lambdaEventDataFactory: %v", err))
		return
	}

	_, err = invokeArgsReader.Read(invokeArgsBytes)
	if err != nil {
		tracer.AddExceptionTypeAndMessage("aws-sdk-go",
			fmt.Sprintf("lambdaEventDataFactory: %v", err))
		return
	}
	res.Metadata["invoke_args"] = string(invokeArgsBytes)
	_, err = invokeArgsReader.Seek(initialOffset, io.SeekStart)
	if err != nil {
		tracer.AddExceptionTypeAndMessage("aws-sdk-go",
			fmt.Sprintf("lambdaEventDataFactory: %v", err))
		return
	}
}
