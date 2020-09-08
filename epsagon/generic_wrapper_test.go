package epsagon

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func TestEpsagonTracer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generic Wrapper")
}

type matchUserError struct {
	exception interface{}
}

func (matcher *matchUserError) Match(actual interface{}) (bool, error) {
	uErr, ok := actual.(userError)
	if !ok {
		return false, fmt.Errorf("excpects userError, got %v", actual)
	}

	if !reflect.DeepEqual(uErr.exception, matcher.exception) {
		return false, fmt.Errorf("expected\n\t%v\nexception, got\n\t%v", matcher.exception, uErr.exception)
	}

	return true, nil
}

func (matcher *matchUserError) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto be userError with exception\n\t%#v", actual, matcher.exception)
}

func (matcher *matchUserError) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("NegatedFailureMessage")
}

func MatchUserError(exception interface{}) types.GomegaMatcher {
	return &matchUserError{
		exception: exception,
	}
}

var _ = Describe("generic_wrapper", func() {
	Describe("GoWrapper", func() {
		Context("called with nil config", func() {
			It("Returns a valid function", func() {
				wrapper := GoWrapper(nil, func() {})
				wrapperType := reflect.TypeOf(wrapper)
				Expect(wrapperType.Kind()).To(Equal(reflect.Func))
			})
		})
	})
	Describe("epsagonGenericWrapper", func() {
		Context("Happy Flows", func() {
			var (
				events     []*protocol.Event
				exceptions []*protocol.Exception
			)
			BeforeEach(func() {
				events = make([]*protocol.Event, 0)
				exceptions = make([]*protocol.Exception, 0)
				tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
					Events:     &events,
					Exceptions: &exceptions,
				}
			})
			It("Calls the user function", func() {
				called := false
				wrapper := &epsagonGenericWrapper{
					config:  &Config{},
					handler: reflect.ValueOf(func() { called = true }),
					tracer:  tracer.GlobalTracer,
				}
				wrapper.Call()
				Expect(called).To(Equal(true))
				Expect(len(events)).To(Equal(1))
			})
			It("Calls the user function with custom resource name", func() {
				called := false
				resourceName := "test-resource-name"
				wrapper := &epsagonGenericWrapper{
					config:       &Config{},
					handler:      reflect.ValueOf(func() { called = true }),
					tracer:       tracer.GlobalTracer,
					resourceName: resourceName,
				}
				wrapper.Call()
				Expect(called).To(Equal(true))
				Expect(len(events)).To(Equal(1))
				Expect(events[0].Name).To(Equal(resourceName))
			})
			It("Retuns and accepts arguments", func() {
				called := false
				result := false
				wrapper := &epsagonGenericWrapper{
					config: &Config{},
					handler: reflect.ValueOf(
						func(x bool) bool {
							called = x
							return x
						}),
					tracer: tracer.GlobalTracer,
				}
				result = wrapper.Call(true)[0].Bool()
				Expect(called).To(Equal(true))
				Expect(result).To(Equal(true))
				Expect(len(events)).To(Equal(1))
			})
		})
		Context("Error Flows", func() {
			var (
				events     []*protocol.Event
				exceptions []*protocol.Exception
			)
			BeforeEach(func() {
				events = make([]*protocol.Event, 0)
				exceptions = make([]*protocol.Exception, 0)
				tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
					Events:     &events,
					Exceptions: &exceptions,
				}
			})
			It("Panics for wrong number of arguments", func() {
				called := false
				wrapper := &epsagonGenericWrapper{
					config:  &Config{},
					handler: reflect.ValueOf(func(x bool) { called = x }),
					tracer:  tracer.GlobalTracer,
				}
				Expect(func() { wrapper.Call() }).To(Panic())
				Expect(called).To(Equal(false))
				Expect(len(events)).To(Equal(1))
				Expect(events[0].Exception).NotTo(Equal(nil))
			})
			It("Failed to add event", func() {
				tracer.GlobalTracer.(*tracer.MockedEpsagonTracer).PanicAddEvent = true
				called := false
				wrapper := &epsagonGenericWrapper{
					config:  &Config{},
					handler: reflect.ValueOf(func() { called = true }),
					tracer:  tracer.GlobalTracer,
				}
				wrapper.Call()
				Expect(called).To(Equal(true))
				Expect(len(exceptions)).To(Equal(1))
			})
			It("User function panics", func() {
				wrapper := &epsagonGenericWrapper{
					config:  &Config{},
					handler: reflect.ValueOf(func() { panic("boom") }),
					tracer:  tracer.GlobalTracer,
				}
				Expect(func() { wrapper.Call() }).To(
					PanicWith(MatchUserError("boom")))
				Expect(len(events)).To(Equal(1))
				Expect(events[0].Exception).NotTo(Equal(nil))
			})
		})
	})
})
