package epsagonawsv2factories

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/epsagon/epsagon-go/protocol"
	"reflect"
	"strings"
)

// SnsEventDataFactory creates an Epsagon Resource from aws.Request
func SnsEventDataFactory(r *aws.Request, res *protocol.Resource, metadataOnly bool) {
	if !metadataOnly {
		inputValue := reflect.ValueOf(r.Params).Elem()
		updateMetadataFromValue(inputValue, "Message", "Notification Message", res.Metadata)
	}
	handleSpecificOperation(r, res, metadataOnly,
		map[string]specificOperationHandler{
			"CreateTopic": handleSNSCreateTopic,
			"Publish":     handlerSNSPublish,
		},
		handleSNSdefault,
	)
}

func handleSNSdefault(r *aws.Request, res *protocol.Resource, metadataOnly bool) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	topicArn, ok := getFieldStringPtr(inputValue, "TopicArn")
	if ok {
		splitTopic := strings.Split(topicArn, ":")
		res.Name = splitTopic[len(splitTopic)-1]
	}
}

func handleSNSCreateTopic(r *aws.Request, res *protocol.Resource, metadataOnly bool) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	name, ok := getFieldStringPtr(inputValue, "Name")
	if ok {
		res.Name = name
	}
}

func handlerSNSPublish(r *aws.Request, res *protocol.Resource, metadataOnly bool) {
	handleSNSdefault(r, res, metadataOnly)
	outputValue := reflect.ValueOf(r.Data).Elem()
	updateMetadataFromValue(outputValue, "MessageId", "Message ID", res.Metadata)
}
