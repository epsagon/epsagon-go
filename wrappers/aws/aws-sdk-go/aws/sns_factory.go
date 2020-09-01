package epsagonawswrapper

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"reflect"
	"strings"
)

func snsEventDataFactory(
	r *request.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	if !metadataOnly {
		inputValue := reflect.ValueOf(r.Params).Elem()
		updateMetadataFromValue(inputValue, "Message", "Notification Message", res.Metadata)
	}
	handleSpecificOperation(r, res, metadataOnly,
		map[string]specificOperationHandler{
			"CreateTopic": handleSNSCreateTopic,
			"Publish":     handlerSNSPublish,
		},
		handleSNSdefault, currentTracer,
	)
}

func handleSNSdefault(r *request.Request, res *protocol.Resource, metadataOnly bool, _ tracer.Tracer) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	topicArn, ok := getFieldStringPtr(inputValue, "TopicArn")
	if ok {
		splitTopic := strings.Split(topicArn, ":")
		res.Name = splitTopic[len(splitTopic)-1]
	}
}

func handleSNSCreateTopic(r *request.Request, res *protocol.Resource, metadataOnly bool, _ tracer.Tracer) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	name, ok := getFieldStringPtr(inputValue, "Name")
	if ok {
		res.Name = name
	}
}

func handlerSNSPublish(r *request.Request, res *protocol.Resource, metadataOnly bool, _ tracer.Tracer) {
	handleSNSdefault(r, res, metadataOnly)
	outputValue := reflect.ValueOf(r.Data).Elem()
	updateMetadataFromValue(outputValue, "MessageId", "Message ID", res.Metadata)
}
