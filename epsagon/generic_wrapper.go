package epsagon

import (
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/google/uuid"
	"reflect"
	"runtime"
)

// epsagonGenericWrapper is a generic lambda function type
type epsagonGenericWrapper struct {
	handler  reflect.Value
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
func (wrapper *epsagonGenericWrapper) Call(args ...interface{}) []reflect.Value {
	wrapper.createTracer()
	defer StopTracer()

	simpleEvent := &protocol.Event{
		Id:        uuid.New().String(),
		Origin:    "runner",
		StartTime: GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      runtime.FuncForPC(wrapper.handler.Pointer()).Name(),
			Type:      "go-function",
			Operation: "invoke",
		},
	}
	AddEvent(simpleEvent)

	if wrapper.handler.Type().NumIn() != len(args) {
		panic("wrong number of args")
	}
	inputs := make([]reflect.Value, len(args))
	for k, in := range args {
		inputs[k] = reflect.ValueOf(in)
	}

	wrapper.invoked = true
	return wrapper.handler.Call(inputs)
}

// GenericFunction type
type GenericFunction func(args ...interface{}) []reflect.Value

// GoWrapper wraps the function with epsagon's tracer
func GoWrapper(config *Config, wrappedFunction interface{}) GenericFunction {
	return func(args ...interface{}) []reflect.Value {
		wrapper := &epsagonGenericWrapper{
			config:  config,
			handler: reflect.ValueOf(wrappedFunction),
		}
		return wrapper.Call(args...)
	}
}
