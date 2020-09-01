package epsagonawsv2factories

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"reflect"
)

// STSEventDataFactory to create epsagon Resource from aws.Request to STS
func StsDataFactory(
	r *aws.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	handleSpecificOperations := map[string]specificOperationHandler{
		"GetCallerIdentity": handleStsGetCallerIdentityRequest,
	}
	handleSpecificOperation(r, res, metadataOnly, handleSpecificOperations, nil, currentTracer)
}

func handleStsGetCallerIdentityRequest(
	r *aws.Request,
	res *protocol.Resource,
	metadataOnly bool,
	_ tracer.Tracer,
) {
	if !metadataOnly {
		outputValue := reflect.ValueOf(r.Data).Elem()
		for _, key := range []string{"Account", "Arn", "UserId"} {
			updateMetadataField(outputValue, key, res)
		}
	}
}
