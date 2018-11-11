package epsagon

import (
	// "fmt"
	"context"
	"encoding/json"
	"github.com/epsagon/epsagon-go/protocol"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"reflect"
)

var _ = Describe("lambda_wrapper", func() {
	Describe("WrapLambdaHandler", func() {
		Context("called with nil config", func() {
			It("Returns a function suitable for lambda", func() {
				wrapper := WrapLambdaHandler(nil, func() {})
				wrapperType := reflect.TypeOf(wrapper)
				Expect(wrapperType.Kind()).To(Equal(reflect.Func))
				_, err := validateArguments(wrapperType)
				Expect(err).To(BeNil())
				err = validateReturns(wrapperType)
				Expect(err).To(BeNil())
			})
			It("calls the wrapped function", func() {
				GlobalTracer = nil
				called := false
				wrapper := WrapLambdaHandler(nil, func() {
					called = true
				})
				wrapperValue := reflect.ValueOf(wrapper)
				ctx := context.Background()
				var args []reflect.Value
				args = append(args, reflect.ValueOf(ctx))
				args = append(args, reflect.ValueOf(json.RawMessage("{}")))
				wrapperValue.Call(args)
				Expect(called).To(Equal(true))
			})
		})
		Context("called with nil handler", func() {
			It("Returns a function suitable for lambda", func() {
				wrapper := WrapLambdaHandler(&Config{}, nil)
				wrapperType := reflect.TypeOf(wrapper)
				_, err := validateArguments(wrapperType)
				Expect(err).To(BeNil())
				err = validateReturns(wrapperType)
				Expect(err).To(BeNil())
			})
		})
	})
	Describe("Invoke", func() {
		Context("Happy Flows", func() {
			var (
				events     []*protocol.Event
				exceptions []*protocol.Exception
			)
			BeforeEach(func() {
				events = make([]*protocol.Event, 0)
				exceptions = make([]*protocol.Exception, 0)
				GlobalTracer = &MockedEpsagonTracer{
					Events:     &events,
					Exceptions: &exceptions,
				}
			})
			It("Adds an Event, Trigger and calls handler", func() {
				called := false
				wrapper := &epsagonLambdaWrapper{
					config:  &Config{},
					handler: makeGenericHandler(func() { called = true }),
				}

				ctx := context.Background()
				payload := json.RawMessage("{}")
				wrapper.Invoke(ctx, payload)

				Expect(called).To(Equal(true))
				Expect(exceptions).To(BeEmpty())
				Expect(events).To(HaveLen(2))
			})
		})
		Describe("Error Flows", func() {
			var (
				events     []*protocol.Event
				exceptions []*protocol.Exception
				called     bool
				wrapper    *epsagonLambdaWrapper
			)
			BeforeEach(func() {
				called = false
				events = make([]*protocol.Event, 0)
				exceptions = make([]*protocol.Exception, 0)
				wrapper = &epsagonLambdaWrapper{
					config:  &Config{},
					handler: makeGenericHandler(func() { called = true }),
				}
			})
			Context("Failed to add event", func() {
				It("Recovers and adds exception", func() {
					GlobalTracer = &MockedEpsagonTracer{
						Events:        &events,
						Exceptions:    &exceptions,
						panicAddEvent: true,
					}
					wrapper.Invoke(context.Background(), json.RawMessage("{}"))
					Expect(called).To(Equal(true))
					Expect(exceptions).To(HaveLen(2))
					Expect(events).To(BeEmpty())
				})
			})
			Context("Failed to add exception and event", func() {
				It("Recovers and does nothing becuase it can't", func() {
					GlobalTracer = &MockedEpsagonTracer{
						Events:            &events,
						Exceptions:        &exceptions,
						panicAddEvent:     true,
						panicAddException: true,
					}
					wrapper.Invoke(context.Background(), json.RawMessage("{}"))
					Expect(called).To(Equal(true))
					Expect(exceptions).To(BeEmpty())
					Expect(events).To(BeEmpty())
				})
			})
			Context("Failed to stop tracer", func() {
				It("Recovers and does nothing because it can't", func() {
					GlobalTracer = &MockedEpsagonTracer{
						Events:     &events,
						Exceptions: &exceptions,
						panicStop:  true,
					}
					wrapper.Invoke(context.Background(), json.RawMessage("{}"))
					Expect(called).To(Equal(true))
					Expect(exceptions).To(BeEmpty())
					Expect(events).To(HaveLen(2))
				})
			})
		})
	})
})
