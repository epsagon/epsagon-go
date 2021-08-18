package epsagon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
				tracer.GlobalTracer = nil
				called := false
				wrapper := WrapLambdaHandler(
					&Config{Config: tracer.Config{Disable: true}},
					func() { called = true },
				)
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
				tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
					Events:     &events,
					Exceptions: &exceptions,
				}
			})
			It("Adds an Event, Trigger and calls handler", func() {
				called := false
				wrapper := &epsagonLambdaWrapper{
					config:  &Config{},
					handler: makeGenericHandler(func() { called = true }),
					tracer:  tracer.GlobalTracer,
				}

				ctx := context.Background()
				payload := json.RawMessage("{}")
				wrapper.Invoke(ctx, payload)

				Expect(called).To(Equal(true))
				Expect(exceptions).To(BeEmpty())
				Expect(events).To(HaveLen(2))
			})

			Context("Lambda timeout handling", func() {
				It("Marks event as success when timeout defined but not reached", func() {
					const lambdaTimeout = 5 * time.Minute

					called := false
					wrapper := &epsagonLambdaWrapper{
						config:  &Config{},
						handler: makeGenericHandler(func() { called = true }),
						tracer:  tracer.GlobalTracer,
					}

					lambdaDeadline := time.Now().Add(lambdaTimeout)
					ctx, cancel := context.WithDeadline(context.Background(), lambdaDeadline)
					defer cancel()

					payload := json.RawMessage("{}")
					wrapper.Invoke(ctx, payload)

					Expect(called).To(Equal(true))
					Expect(exceptions).To(BeEmpty())
					Expect(events).To(HaveLen(2))

					var trigger, runner *protocol.Event
					for _, event := range events {
						if event.Origin == "trigger" {
							trigger = event
						} else if event.Origin == "runner" {
							runner = event
						}
					}
					Expect(trigger).NotTo(BeNil())
					Expect(runner).NotTo(BeNil())
					Expect(trigger.ErrorCode).To(Equal(protocol.ErrorCode_OK))
					Expect(runner.ErrorCode).To(Equal(protocol.ErrorCode_OK))
				})

				It("Marks event as timeout when the default lambda timout threshold reached", func() {
					const lambdaTimeout = (tracer.DefaultLambdaTimeoutThresholdMs + 100) * time.Millisecond

					called := false
					wrapper := &epsagonLambdaWrapper{
						config: &Config{},
						handler: makeGenericHandler(func() {
							called = true
							time.Sleep(lambdaTimeout)
						}),
						tracer: tracer.GlobalTracer,
					}

					lambdaDeadline := time.Now().Add(lambdaTimeout)
					ctx, cancel := context.WithDeadline(context.Background(), lambdaDeadline)
					defer cancel()

					payload := json.RawMessage("{}")
					wrapper.Invoke(ctx, payload)

					Expect(called).To(Equal(true))
					Expect(exceptions).To(BeEmpty())
					Expect(events).To(HaveLen(2))

					var trigger, runner *protocol.Event
					for _, event := range events {
						if event.Origin == "trigger" {
							trigger = event
						} else if event.Origin == "runner" {
							runner = event
						}
					}
					Expect(trigger).NotTo(BeNil())
					Expect(runner).NotTo(BeNil())
					Expect(trigger.ErrorCode).To(Equal(protocol.ErrorCode_OK))
					Expect(runner.ErrorCode).To(Equal(TimeoutErrorCode))
				})

				It("Marks event as timeout when a user defined lambda timout threshold reached", func() {
					const userDefinedTimeoutThresholdMs = 50
					os.Setenv("EPSAGON_LAMBDA_TIMEOUT_THRESHOLD_MS", fmt.Sprint(userDefinedTimeoutThresholdMs))
					defer os.Unsetenv("EPSAGON_LAMBDA_TIMEOUT_THRESHOLD_MS")

					const lambdaTimeout = (userDefinedTimeoutThresholdMs + 100) * time.Millisecond

					called := false
					wrapper := &epsagonLambdaWrapper{
						config: &Config{},
						handler: makeGenericHandler(func() {
							called = true
							time.Sleep(lambdaTimeout)
						}),
						tracer: tracer.GlobalTracer,
					}

					lambdaDeadline := time.Now().Add(lambdaTimeout)
					ctx, cancel := context.WithDeadline(context.Background(), lambdaDeadline)
					defer cancel()

					payload := json.RawMessage("{}")
					wrapper.Invoke(ctx, payload)

					Expect(called).To(Equal(true))
					Expect(exceptions).To(BeEmpty())
					Expect(events).To(HaveLen(2))

					var trigger, runner *protocol.Event
					for _, event := range events {
						if event.Origin == "trigger" {
							trigger = event
						} else if event.Origin == "runner" {
							runner = event
						}
					}
					Expect(trigger).NotTo(BeNil())
					Expect(runner).NotTo(BeNil())
					Expect(trigger.ErrorCode).To(Equal(protocol.ErrorCode_OK))
					Expect(runner.ErrorCode).To(Equal(TimeoutErrorCode))
				})
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
				tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
					Events:     &events,
					Exceptions: &exceptions,
				}
				wrapper = &epsagonLambdaWrapper{
					config:  &Config{},
					handler: makeGenericHandler(func() { called = true }),
					tracer:  tracer.GlobalTracer,
				}
			})
			Context("Failed to add event", func() {
				It("Recovers and adds exception", func() {
					tracer.GlobalTracer.(*tracer.MockedEpsagonTracer).PanicAddEvent = true
					wrapper.Invoke(context.Background(), json.RawMessage("{}"))
					Expect(called).To(Equal(true))
					Expect(exceptions).To(HaveLen(2))
					Expect(events).To(BeEmpty())
				})
			})
			Context("Failed to add exception and event", func() {
				It("Recovers and does nothing becuase it can't", func() {
					tracer.GlobalTracer.(*tracer.MockedEpsagonTracer).PanicAddEvent = true
					tracer.GlobalTracer.(*tracer.MockedEpsagonTracer).PanicAddException = true
					wrapper.Invoke(context.Background(), json.RawMessage("{}"))
					Expect(called).To(Equal(true))
					Expect(exceptions).To(BeEmpty())
					Expect(events).To(BeEmpty())
				})
			})
			Context("Failed to stop tracer", func() {
				It("Recovers and does nothing because it can't", func() {
					tracer.GlobalTracer.(*tracer.MockedEpsagonTracer).PanicStop = true
					wrapper.Invoke(context.Background(), json.RawMessage("{}"))
					Expect(called).To(Equal(true))
					Expect(exceptions).To(BeEmpty())
					Expect(events).To(HaveLen(2))
				})
			})
		})
	})
})
