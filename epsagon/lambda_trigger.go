package epsagon

import (
	"bytes"
	"encoding/json"
	"fmt"
	lambdaEvents "github.com/aws/aws-lambda-go/events"
	"github.com/epsagon/epsagon-go/protocol"
	"reflect"
	"runtime/debug"
	"strings"
	"time"
)

type triggerFactory func(event interface{}, metadataOnly bool) *protocol.Event

func unknownTrigger(event interface{}, metadataOnly bool) *protocol.Event {
	return &protocol.Event{}
}

func getReflectType(i interface{}) reflect.Type {
	return reflect.TypeOf(i)
}

func mapParametersToString(params map[string]string) string {
	buf, err := json.Marshal(params)
	if err != nil {
		AddException(&protocol.Exception{
			Type:      "trigger-creation",
			Message:   fmt.Sprintf("Failed to serialize %v", params),
			Traceback: string(debug.Stack()),
			Time:      float64(time.Now().UTC().Unix()),
		})
		return ""
	}
	return string(buf)
}

func triggerAPIGatewayProxyRequest(rawEvent interface{}, metadataOnly bool) *protocol.Event {
	event, ok := rawEvent.(lambdaEvents.APIGatewayProxyRequest)
	if !ok {
		AddException(&protocol.Exception{
			Type: "trigger-creation",
			Message: fmt.Sprintf(
				"failed to convert rawEvent to lambdaEvents.APIGatewayProxyRequest %v",
				rawEvent),
			Time: float64(time.Now().UTC().Unix()),
		})
		return nil
	}
	triggerEvent := &protocol.Event{
		Id:        event.RequestContext.RequestID,
		Origin:    "trigger",
		StartTime: float64(time.Now().UTC().Unix()),
		Resource: &protocol.Resource{
			Name:      event.Resource,
			Type:      "api_gateway",
			Operation: event.HTTPMethod,
			Metadata: map[string]string{
				"stage":                   event.RequestContext.Stage,
				"query_string_parameters": mapParametersToString(event.QueryStringParameters),
				"path_parameters":         mapParametersToString(event.PathParameters),
			},
		},
	}
	if !metadataOnly {
		if bodyJSON, err := json.Marshal(event.Body); err != nil {
			AddException(&protocol.Exception{
				Type:      "trigger-creation",
				Message:   fmt.Sprintf("Failed to serialize body %s", event.Body),
				Traceback: string(debug.Stack()),
				Time:      float64(time.Now().UTC().Unix()),
			})
			triggerEvent.Resource.Metadata["body"] = ""
		} else {
			triggerEvent.Resource.Metadata["body"] = string(bodyJSON)
		}
		triggerEvent.Resource.Metadata["headers"] = mapParametersToString(event.Headers)
	}

	return triggerEvent
}

type factoryAndType struct {
	EventType reflect.Type
	Factory   triggerFactory
}

var (
	triggerFactories = map[string]factoryAndType{
		"api_gateway": factoryAndType{
			EventType: reflect.TypeOf(lambdaEvents.APIGatewayProxyRequest{}),
			Factory:   triggerAPIGatewayProxyRequest,
		},
	}
)

func decodeAndUnpackEvent(
	payload json.RawMessage,
	eventType reflect.Type,
	factory triggerFactory,
	metadataOnly bool,
	disallowUnknownFields bool,
) *protocol.Event {

	event := reflect.New(eventType)
	decoder := json.NewDecoder(bytes.NewReader(payload))
	if disallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(event.Interface()); err != nil {
		// fmt.Printf("DEBUG: addLambdaTrigger error in json decoder: %v\n", err)
		return nil
	}
	return factory(event.Elem().Interface(), metadataOnly)
}

type recordField struct {
	EventSource string
}

type interestingFields struct {
	Records    []recordField
	HTTPMethod string
	Context    map[string]interface{}
	MethodArn  string
	Source     string
}

func guessTriggerSource(payload json.RawMessage) string {
	var rawEvent interestingFields
	err := json.Unmarshal(payload, &rawEvent)
	if err != nil {
		AddException(&protocol.Exception{
			Type:      "trigger-identification",
			Message:   fmt.Sprintf("Failed to unmarshal json %v\n", err),
			Traceback: string(debug.Stack()),
			Time:      float64(time.Now().UTC().Unix()),
		})
		return ""
	}
	triggerSource := "json"
	if len(rawEvent.Records) > 0 {
		triggerSource = rawEvent.Records[0].EventSource
	} else if len(rawEvent.HTTPMethod) > 0 {
		triggerSource = "api_gateway"
	} else if _, ok := rawEvent.Context["http-method"]; ok {
		triggerSource = "api_gateway_no_proxy"
	} else if len(rawEvent.Source) > 0 {
		sourceSlice := strings.Split(rawEvent.Source, ".")
		triggerSource = sourceSlice[len(sourceSlice)-1]
	}
	return triggerSource
}

func addLambdaTrigger(
	payload json.RawMessage,
	metadataOnly bool,
	triggerFactories map[string]factoryAndType) {

	var triggerEvent *protocol.Event
	for _, factoryStruct := range triggerFactories {
		triggerEvent = decodeAndUnpackEvent(
			payload, factoryStruct.EventType, factoryStruct.Factory, metadataOnly, true)
	}

	triggerSource := guessTriggerSource(payload)
	factoryStruct, found := triggerFactories[triggerSource]
	if found {
		triggerEvent = decodeAndUnpackEvent(
			payload, factoryStruct.EventType, factoryStruct.Factory, metadataOnly, false)
	}

	// If a trigger was found
	if triggerEvent != nil {
		AddEvent(triggerEvent)
	}
}
