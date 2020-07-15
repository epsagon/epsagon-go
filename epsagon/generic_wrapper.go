package epsagon

import (
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/google/uuid"
)

// GenericFunction type to represent any function
type GenericFunction func(args ...interface{}) []interface{}

// epsagonGenericWrapper is a generic lambda function type
type epsagonGenericWrapper struct {
	handler  GenericFunction
	config   *Config
	invoked  bool
	invoking bool
}

func (wrapper *epsagonGenericWrapper) createTracer() {
	if wrapper.config == nil {
		wrapper.config = &Config{}
	}
	CreateTracer(wrapper.config)
}

// Call the wrapped function
func (wrapper *epsagonGenericWrapper) Call(args ...interface{}) []interface{} {
	wrapper.createTracer()
	defer StopTracer()

	simpleEvent := &protocol.Event{
		Id:        uuid.New().String(),
		Origin:    "runner",
		StartTime: GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      "test-go-generic-wrapper",
			Type:      "go-wrapper",
			Operation: "invoke",
		},
	}
	AddEvent(simpleEvent)

	wrapper.invoked = true
	return wrapper.handler(args...)
}

// GoWrapper wraps the function with epsagon's tracer
func GoWrapper(config *Config, wrappedFunction GenericFunction) GenericFunction {
	return func(args ...interface{}) []interface{} {
		wrapper := &epsagonGenericWrapper{
			config:  config,
			handler: wrappedFunction,
		}
		return wrapper.Call(args...)
	}
}
