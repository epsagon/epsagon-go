package epsagon

import (
	"context"
	"encoding/json"
	"fmt"
	lambdaEvents "github.com/aws/aws-lambda-go/events"
	"github.com/epsagon/epsagon-go/protocol"
	"reflect"
	"runtime/debug"
	"time"
)

type triggerFactory func(ctx context.Context, event interface{}, metadataOnly bool) *protocol.Event

func unknownTrigger(ctx context.Context, event interface{}, metadataOnly bool) *protocol.Event {
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

func triggerAPIGatewayProxyRequest(ctx context.Context, rawEvent interface{}, metadataOnly bool) *protocol.Event {
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
			triggerEvent.Resource.Metadata["body"] = string(bodyJson)
		}
		triggerEvent.Resource.Metadata["headers"] = mapParametersToString(event.Headers)
	}

	return triggerEvent
}

var (
	triggerFactories = map[reflect.Type]triggerFactory{
		getReflectType(lambdaEvents.APIGatewayProxyRequest{}): triggerAPIGatewayProxyRequest,
	}
)

func addLambdaTrigger(ctx context.Context, payload json.RawMessage, metadataOnly bool, triggerFactories map[reflect.Type]triggerFactory) {
	var triggerEvent *protocol.Event
	for eventType, factory := range triggerFactories {
		event := reflect.New(eventType)
		if err := json.Unmarshal(payload, event.Interface()); err == nil {
			// On Success:
			triggerEvent = factory(ctx, event.Elem().Interface(), metadataOnly)
		}
	}
	// If a trigger was found
	if triggerEvent != nil {
		AddEvent(triggerEvent)
	}
}
