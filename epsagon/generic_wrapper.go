package epsagon

import (
	"context"
	"fmt"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/google/uuid"
	"reflect"
	"runtime"
	"runtime/debug"
)

type userError struct {
	exception interface{}
	stack     string
}

// epsagonGenericWrapper is a generic lambda function type
type epsagonGenericWrapper struct {
	handler       reflect.Value
	config        *Config
	tracer        tracer.Tracer
	runner        *protocol.Event
	thrownError   interface{}
	invoked       bool
	invoking      bool
	dontAddRunner bool
	concurrent    bool
}

// createRunner creates a runner event but does not add it to the tracer
// the runner is saved for further manipulations at wrapper.runner
func (wrapper *epsagonGenericWrapper) createRunner() {
	wrapper.runner = &protocol.Event{
		Id:        uuid.New().String(),
		Origin:    "runner",
		StartTime: tracer.GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      runtime.FuncForPC(wrapper.handler.Pointer()).Name(),
			Type:      "go-function",
			Operation: "invoke",
		},
		ErrorCode: protocol.ErrorCode_OK,
	}
}

// For instances when you want to add event but can't risk exception
func (wrapper *epsagonGenericWrapper) safeAddRunnerEvent() {
	defer func() {
		recover()
	}()
	wrapper.addRunnerEvent()
}

func (wrapper *epsagonGenericWrapper) addRunnerEvent() {
	if wrapper.dontAddRunner {
		return
	}
	endTime := tracer.GetTimestamp()
	wrapper.runner.Duration = endTime - wrapper.runner.StartTime
	wrapper.tracer.AddEvent(wrapper.runner)
}

// Change the arguments from interface{} to reflect.Value array
func (wrapper *epsagonGenericWrapper) transformArguments(args ...interface{}) []reflect.Value {
	actualLength := len(args)
	if wrapper.concurrent {
		actualLength += 1
	}
	if wrapper.handler.Type().NumIn() != actualLength {
		msg := fmt.Sprintf(
			"Wrong number of args: %d, expected: %d",
			actualLength, wrapper.handler.Type().NumIn())
		wrapper.createRunner()
		wrapper.runner.Exception = &protocol.Exception{
			Type:    "Runtime Error",
			Message: fmt.Sprintf("%v", msg),
			Time:    tracer.GetTimestamp(),
		}
		wrapper.safeAddRunnerEvent()
		panic(msg)
	}
	// add new context to inputs
	inputs := make([]reflect.Value, actualLength)
	argsInputs := inputs
	if wrapper.concurrent {
		inputs[0] = reflect.ValueOf(
			context.WithValue(context.Background(), "tracer", wrapper.tracer))
		argsInputs = argsInputs[1:]
	}
	for k, in := range args {
		argsInputs[k] = reflect.ValueOf(in)
	}
	return inputs
}

// Call the wrapped function
func (wrapper *epsagonGenericWrapper) Call(args ...interface{}) (results []reflect.Value) {
	inputs := wrapper.transformArguments(args...)
	defer func() {
		wrapper.thrownError = recover()
		if wrapper.thrownError != nil {
			exception := &protocol.Exception{
				Type:      "Runtime Error",
				Message:   fmt.Sprintf("%v", wrapper.thrownError),
				Traceback: string(debug.Stack()),
				Time:      tracer.GetTimestamp(),
			}
			if wrapper.invoking {
				wrapper.runner.Exception = exception
				wrapper.runner.ErrorCode = protocol.ErrorCode_EXCEPTION
				wrapper.safeAddRunnerEvent()
				panic(userError{
					exception: wrapper.thrownError,
					stack:     exception.Traceback,
				})
			} else {
				exception.Type = "GenericEpsagonWrapper"
				wrapper.tracer.AddException(exception)
				if !wrapper.invoked { // attempt to run the user's function untraced
					results = wrapper.handler.Call(inputs)
				}
			}
		}
	}()

	wrapper.createRunner()
	wrapper.invoking = true
	wrapper.invoked = true
	results = wrapper.handler.Call(inputs)
	wrapper.invoking = false
	wrapper.addRunnerEvent()
	return results
}

// GenericFunction type
type GenericFunction func(args ...interface{}) []reflect.Value

// GoWrapper wraps the function with epsagon's tracer
func GoWrapper(config *Config, wrappedFunction interface{}) GenericFunction {
	return func(args ...interface{}) []reflect.Value {
		if config == nil {
			config = &Config{}
		}
		wrapperTracer := tracer.CreateGlobalTracer(&config.Config)
		wrapperTracer.Start()
		defer wrapperTracer.Stop()

		wrapper := &epsagonGenericWrapper{
			config:  config,
			handler: reflect.ValueOf(wrappedFunction),
			tracer:  wrapperTracer,
		}
		return wrapper.Call(args...)
	}
}

// GoWrapper wraps the function with epsagon's tracer
func ConcurrentGoWrapper(config *Config, wrappedFunction interface{}) GenericFunction {
	return func(args ...interface{}) []reflect.Value {
		if config == nil {
			config = &Config{}
		}
		wrapperTracer := tracer.CreateTracer(&config.Config)
		wrapperTracer.Start()
		defer wrapperTracer.Stop()

		wrapper := &epsagonGenericWrapper{
			config:     config,
			handler:    reflect.ValueOf(wrappedFunction),
			tracer:     wrapperTracer,
			concurrent: true,
		}
		return wrapper.Call(args...)
	}
}
