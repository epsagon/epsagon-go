package epsagon

import (
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"reflect"
	"testing"
)

func TestEpsagonTracer(t *testing.T) {
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
		})
	})
})
