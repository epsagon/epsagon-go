package epsagonawsv2factories

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/epsagon/epsagon-go/protocol"
	"reflect"
)

// SesEventDataFactory creates an Epsagon Resource from aws.Resource
func SesEventDataFactory(r *aws.Request, res *protocol.Resource, metadataOnly bool) {
	handleSpecificOperation(r, res, metadataOnly,
		map[string]specificOperationHandler{
			"SendEmail": handleSESSendEmail,
		},
		nil,
	)
}

func handleSESSendEmail(r *aws.Request, res *protocol.Resource, metadataOnly bool) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	updateMetadataFromValue(inputValue, "Source", "source", res.Metadata)
	updateMetadataWithFieldToJSON(inputValue, "Destination", "destination", res.Metadata)
	messageField := inputValue.FieldByName("Message")
	if messageField != (reflect.Value{}) {
		updateMetadataWithFieldToJSON(messageField, "Subject", "subject", res.Metadata)
		if !metadataOnly {
			updateMetadataWithFieldToJSON(messageField, "Body", "body", res.Metadata)
		}
	}
	outputValue := reflect.ValueOf(r.Data).Elem()
	updateMetadataFromValue(outputValue, "MessageId", "message_id", res.Metadata)
}
