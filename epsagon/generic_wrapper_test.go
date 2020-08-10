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
			})
			It("Panics for wrong number of arguments", func() {
				called := false
				wrapper := &epsagonGenericWrapper{
					config:  &Config{},
					handler: reflect.ValueOf(func(x bool) { called = x }),
					tracer:  tracer.GlobalTracer,
				}
				Expect(wrapper.Call()).To(Panic())
				wrapper.Call(true)
				Expect(called).To(Equal(true))
			})
		})
		Context("called with nil config", func() {
		})
	})
})
