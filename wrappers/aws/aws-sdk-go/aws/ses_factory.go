package epsagonawswrapper

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"reflect"
)

func sesEventDataFactory(
	r *request.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	handleSpecificOperation(r, res, metadataOnly,
		map[string]specificOperationHandler{
			"SendEmail": handleSESSendEmail,
		},
		nil, currentTracer,
	)
}

func handleSESSendEmail(
	r *request.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	updateMetadataFromValue(inputValue, "Source", "source", res.Metadata)
	updateMetadataWithFieldToJSON(inputValue, "Destination", "destination", res.Metadata, currentTracer)
	messageField := inputValue.FieldByName("Message")
	if messageField != (reflect.Value{}) {
		updateMetadataWithFieldToJSON(messageField, "Subject", "subject", res.Metadata, currentTracer)
		if !metadataOnly {
			updateMetadataWithFieldToJSON(messageField, "Body", "body", res.Metadata, currentTracer)
		}
	}
	outputValue := reflect.ValueOf(r.Data).Elem()
	updateMetadataFromValue(outputValue, "MessageId", "message_id", res.Metadata)
}
