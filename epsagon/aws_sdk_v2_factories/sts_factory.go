package epsagonawsv2factories

import (
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"reflect"
)

// StsEventDataFactory to create epsagon Resource from aws.Request to STS
func StsEventDataFactory(
	r *AWSCall,
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
	r *AWSCall,
	res *protocol.Resource,
	metadataOnly bool,
	_ tracer.Tracer,
) {
	if ! metadataOnly {
		inputValue := reflect.ValueOf(r.Input).Elem()
		for _, key := range []string{"Account", "Arn", "UserId"} {
			updateMetadataField(inputValue, key, res)
		}
	}
}
