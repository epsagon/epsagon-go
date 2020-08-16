package epsagonawsv2factories

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/epsagon/epsagon-go/protocol"
	"reflect"
)

// STSEventDataFactory to create epsagon Resource from aws.Request to STS
func StsDataFactory(r *aws.Request, res *protocol.Resource, metadataOnly bool) {
	handleSpecificOperations := map[string]specificOperationHandler{
		"GetCallerIdentity": handleStsGetCallerIdentityRequest,
	}
	handleSpecificOperation(r, res, metadataOnly, handleSpecificOperations, nil)
}

func handleStsGetCallerIdentityRequest(r *aws.Request, res *protocol.Resource, metadataOnly bool) {
	if !metadataOnly {
		outputValue := reflect.ValueOf(r.Data).Elem()
		for _, key := range []string{"Account", "Arn", "UserId"} {
			updateMetadataField(outputValue, key, res)
		}
	}
}
