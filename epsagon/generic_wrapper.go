package epsagon

import (
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"

	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/google/uuid"
)

type userError struct {
	exception interface{}
	stack     string
}

// GenericWrapper is a generic lambda function type
type GenericWrapper struct {
	handler       reflect.Value
	config        *Config
	tracer        tracer.Tracer
	runner        *protocol.Event
	thrownError   interface{}
	resourceName  string
	invoked       bool
	invoking      bool
	dontAddRunner bool
	injectContext bool
}

// WrapGenericFunction return an epsagon wrapper for a generic function
func WrapGenericFunction(
	handler interface{}, config *Config, tracer tracer.Tracer, injectContext bool, resourceName string) *GenericWrapper {
	return &GenericWrapper{
		config:        config,
		handler:       reflect.ValueOf(handler),
		tracer:        tracer,
		injectContext: injectContext,
		resourceName:  resourceName,
	}
}

// createRunner creates a runner event but does not add it to the tracer
// the runner is saved for further manipulations at wrapper.runner
func (wrapper *GenericWrapper) createRunner() {
	resourceName := wrapper.resourceName
	if len(resourceName) == 0 {
		resourceName = runtime.FuncForPC(wrapper.handler.Pointer()).Name()
	}
	wrapper.runner = &protocol.Event{
		Id:        uuid.New().String(),
		Origin:    "runner",
		StartTime: tracer.GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      resourceName,
			Type:      "go-function",
			Operation: "invoke",
			Metadata:  make(map[string]string),
		},
		ErrorCode: protocol.ErrorCode_OK,
	}
}

// For instances when you want to add event but can't risk exception
func (wrapper *GenericWrapper) safeAddRunnerEvent() {
	defer func() {
		recover()
	}()
	wrapper.addRunnerEvent()
}

func (wrapper *GenericWrapper) addRunnerEvent() {
	if wrapper.dontAddRunner {
		return
	}
	endTime := tracer.GetTimestamp()
	wrapper.runner.Duration = endTime - wrapper.runner.StartTime
	wrapper.tracer.AddEvent(wrapper.runner)
}

// Change the arguments from interface{} to reflect.Value array
func (wrapper *GenericWrapper) transformArguments(args ...interface{}) []reflect.Value {
	actualLength := len(args)
	if wrapper.injectContext {
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
	if wrapper.injectContext {
		inputs[0] = reflect.ValueOf(ContextWithTracer(wrapper.tracer))
		argsInputs = argsInputs[1:]
	}
	for k, in := range args {
		argsInputs[k] = reflect.ValueOf(in)
	}
	return inputs
}

// Call the wrapped function
func (wrapper *GenericWrapper) Call(args ...interface{}) (results []reflect.Value) {
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

func getResourceName(args []string) (resourceName string) {
	if len(args) > 0 {
		resourceName = args[0]
	}
	return
}

// GoWrapper wraps the function with epsagon's tracer
func GoWrapper(config *Config, wrappedFunction interface{}, args ...string) GenericFunction {
	resourceName := getResourceName(args)
	return func(args ...interface{}) []reflect.Value {
		if config == nil {
			config = &Config{}
		}
		wrapperTracer := tracer.CreateGlobalTracer(&config.Config)
		wrapperTracer.Start()
		defer wrapperTracer.Stop()

		wrapper := &GenericWrapper{
			config:       config,
			handler:      reflect.ValueOf(wrappedFunction),
			tracer:       wrapperTracer,
			resourceName: resourceName,
		}
		return wrapper.Call(args...)
	}
}

// ConcurrentGoWrapper wraps the function with epsagon's tracer
func ConcurrentGoWrapper(config *Config, wrappedFunction interface{}, args ...string) GenericFunction {
	resourceName := getResourceName(args)
	return func(args ...interface{}) []reflect.Value {
		if config == nil {
			config = &Config{}
		}
		wrapperTracer := tracer.CreateTracer(&config.Config)
		wrapperTracer.Start()
		defer wrapperTracer.Stop()

		wrapper := &GenericWrapper{
			config:        config,
			handler:       reflect.ValueOf(wrappedFunction),
			tracer:        wrapperTracer,
			injectContext: true,
			resourceName:  resourceName,
		}
		return wrapper.Call(args...)
	}
}
