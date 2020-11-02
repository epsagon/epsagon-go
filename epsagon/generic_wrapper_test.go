package epsagon

import (
	"reflect"
	"testing"

	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGenericWrapper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generic Wrapper")
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
	Describe("GenericWrapper", func() {
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
				wrapper := &GenericWrapper{
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
				wrapper := &GenericWrapper{
					config:       &Config{},
					handler:      reflect.ValueOf(func() { called = true }),
					tracer:       tracer.GlobalTracer,
					resourceName: resourceName,
				}
				wrapper.Call()
				Expect(called).To(Equal(true))
				Expect(len(events)).To(Equal(1))
				Expect(events[0].Resource.Name).To(Equal(resourceName))
			})
			It("Retuns and accepts arguments", func() {
				called := false
				result := false
				wrapper := &GenericWrapper{
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
				wrapper := &GenericWrapper{
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
				wrapper := &GenericWrapper{
					config:  &Config{},
					handler: reflect.ValueOf(func() { called = true }),
					tracer:  tracer.GlobalTracer,
				}
				wrapper.Call()
				Expect(called).To(Equal(true))
				Expect(len(exceptions)).To(Equal(1))
			})
			It("User function panics", func() {
				wrapper := &GenericWrapper{
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
