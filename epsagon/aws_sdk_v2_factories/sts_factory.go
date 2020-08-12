package epsagonawsv2factories

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/epsagon/epsagon-go/protocol"
	"reflect"
)

// DynamodbEventDataFactory to create epsagon Resource from aws.Request to DynamoDB
func StsDataFactory(r *aws.Request, res *protocol.Resource, metadataOnly bool) {
	handleSpecificOperations := map[string]specificOperationHandler{
		"GetCallerIdentity": handleStsGetCallerIdentityRequest,
	}
	handleSpecificOperation(r, res, metadataOnly, handleSpecificOperations, nil)
}

func updateMetadataField(data reflect.Value, key string, res *protocol.Resource) {
	value, ok := getFieldStringPtr(data, key)
	if ok {
		res.Metadata[key] = value
	}
}

func handleStsGetCallerIdentityRequest(r *aws.Request, res *protocol.Resource, metadataOnly bool) {
	outputValue := reflect.ValueOf(r.Data).Elem()
	if !metadataOnly {
		for _, key := range []string{"Account", "Arn", "UserId"} {
			updateMetadataField(outputValue, key, res)
		}
	}
}
