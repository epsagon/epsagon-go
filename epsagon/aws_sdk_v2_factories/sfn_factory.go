package epsagonawsv2factories

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/epsagon/epsagon-go/protocol"
	"reflect"
	"strings"
)

// SfnEventDataFactory creates an Epsagon Resource from aws.Request
func SfnEventDataFactory(r *aws.Request, res *protocol.Resource, metadataOnly bool) {
	res.Type = "stepfunctions"
	handleSpecificOperation(r, res, metadataOnly,
		map[string]specificOperationHandler{
			"PutRecord": handleSFNStartExecution,
		}, nil,
	)
}

func handleSFNStartExecution(r *aws.Request, res *protocol.Resource, metadataOnly bool) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	arn, ok := getFieldStringPtr(inputValue, "StateMachineArn")
	if ok {
		arnParts := strings.Split(arn, ":")
		res.Name = arnParts[len(arnParts)]
		res.Metadata["State Machine ARN"] = arn
	}
	updateMetadataFromValue(inputValue, "Name", "Execution Name", res.Metadata)
	if !metadataOnly {
		updateMetadataFromValue(inputValue, "Input", "input", res.Metadata)
	}
	outputValue := reflect.ValueOf(r.Data).Elem()
	updateMetadataFromValue(outputValue, "ExecutionArn", "Execution ARN", res.Metadata)
}
