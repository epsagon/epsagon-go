package epsagonawswrapper

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"reflect"
	"strings"
)

const InvalidFieldValue = "<invalid Value>"

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

// gets the target name
func getSNStargetName(inputValue reflect.Value, targetKey string) (string, bool) {
	arnString, ok := getFieldStringPtr(inputValue, targetKey)
	if !ok {
		return "", false
	}
	arnSplit := strings.Split(arnString, ":")
	targetName := arnSplit[len(arnSplit)-1]
	return targetName, targetName != InvalidFieldValue
}

func handleSNSdefault(r *request.Request, res *protocol.Resource, metadataOnly bool, _ tracer.Tracer) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	targetName, ok := getSNStargetName(inputValue, "TopicArn")
	if ok {
		res.Name = targetName
		return
	}
	targetName, ok = getSNStargetName(inputValue, "TargetArn")
	if ok {
		res.Name = targetName
	}
	return
	topicArn, ok := getFieldStringPtr(inputValue, "TopicArn")
	if ok {
		splitTopic := strings.Split(topicArn, ":")
		topicName := splitTopic[len(splitTopic)-1]
		if topicName == InvalidFieldValue {
			targetArn, ok := getFieldStringPtr(inputValue, "TargetArn")
			if ok {
				splitTarget := strings.Split(targetArn, ":")
				targetName := splitTopic[len(splitTarget)-1]
				if targetName != InvalidFieldValue {
					res.Name = targetName
				}

			}
		} else {
			res.Name = splitTopic[len(splitTopic)-1]
		}
	}
}

func handleSNSCreateTopic(r *request.Request, res *protocol.Resource, metadataOnly bool, _ tracer.Tracer) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	name, ok := getFieldStringPtr(inputValue, "Name")
	if ok {
		res.Name = name
	}
}

func handlerSNSPublish(r *request.Request, res *protocol.Resource, metadataOnly bool, currentTracer tracer.Tracer) {
	handleSNSdefault(r, res, metadataOnly, currentTracer)
	outputValue := reflect.ValueOf(r.Data).Elem()
	updateMetadataFromValue(outputValue, "MessageId", "Message ID", res.Metadata)
}
