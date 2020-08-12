package tracer

import (
	"github.com/epsagon/epsagon-go/protocol"
)

// MockedEpsagonTracer will not send traces if closed
type MockedEpsagonTracer struct {
	Exceptions *[]*protocol.Exception
	Events     *[]*protocol.Event
	Config     *Config

	PanicStart        bool
	PanicAddEvent     bool
	PanicAddException bool
	PanicStop         bool
}

// Start implementes mocked Start
func (t *MockedEpsagonTracer) Start() {
	if t.PanicStart {
		panic("panic in Start()")
	}
}

// Running implementes mocked Running
func (t *MockedEpsagonTracer) Running() bool {
	return false
}

// Stop implementes mocked Stop
func (t *MockedEpsagonTracer) Stop() {
	if t.PanicStop {
		panic("panic in Stop()")
	}
}

// Stopped implementes mocked Stopped
func (t *MockedEpsagonTracer) Stopped() bool {
	return false
}

// AddEvent implementes mocked AddEvent
func (t *MockedEpsagonTracer) AddEvent(e *protocol.Event) {
	if t.PanicAddEvent {
		panic("panic in AddEvent()")
	}
	*t.Events = append(*t.Events, e)
}

// AddException implementes mocked AddEvent
func (t *MockedEpsagonTracer) AddException(e *protocol.Exception) {
	if t.PanicAddException {
		panic("panic in AddException()")
	}
	*t.Exceptions = append(*t.Exceptions, e)
}

// GetConfig implementes mocked AddEvent
func (t *MockedEpsagonTracer) GetConfig() *Config {
	return t.Config
}

// AddExceptionTypeAndMessage implements AddExceptionTypeAndMessage
func (t *MockedEpsagonTracer) AddExceptionTypeAndMessage(exceptionType, msg string) {
	t.AddException(&protocol.Exception{
		Type:    exceptionType,
		Message: msg,
		Time:    GetTimestamp(),
	})
}
