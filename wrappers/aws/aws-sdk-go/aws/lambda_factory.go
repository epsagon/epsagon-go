package epsagonawswrapper

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"io"
	"reflect"
)

func lambdaEventDataFactory(
	r *request.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
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
		currentTracer.AddExceptionTypeAndMessage("aws-sdk-go",
			fmt.Sprintf("lambdaEventDataFactory: %v", err))
		return
	}

	_, err = invokeArgsReader.Read(invokeArgsBytes)
	if err != nil {
		currentTracer.AddExceptionTypeAndMessage("aws-sdk-go",
			fmt.Sprintf("lambdaEventDataFactory: %v", err))
		return
	}
	res.Metadata["invoke_args"] = string(invokeArgsBytes)
	_, err = invokeArgsReader.Seek(initialOffset, io.SeekStart)
	if err != nil {
		currentTracer.AddExceptionTypeAndMessage("aws-sdk-go",
			fmt.Sprintf("lambdaEventDataFactory: %v", err))
		return
	}
}
