package epsagon

import (
	"context"
	"encoding/json"
	"fmt"
	protocol "github.com/epsagon/epsagon-go/protocol"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"reflect"
)

var _ = Describe("GenericHandler suite", func() {
	Describe("validateArguments", func() {
		Context("Bad Handlers", func() {
			It("fails on too many arguments", func() {
				badHandler := func(ctx context.Context, a, b, c string) error {
					return nil
				}
				_, err := validateArguments(reflect.TypeOf(badHandler))
				Expect(err).NotTo(Equal(nil))
			})
			It("fails if two arguments but first is not context", func() {
				badHandler := func(a string, ctx context.Context) error {
					return nil
				}
				_, err := validateArguments(reflect.TypeOf(badHandler))
				Expect(err).NotTo(Equal(nil))
			})
		})
		Context("Happy Handlers", func() {
			It("accepts no arguments", func() {
				goodHandler := func() error {
					return nil
				}
				hasCtx, err := validateArguments(reflect.TypeOf(goodHandler))
				Expect(err).To(BeNil())
				Expect(hasCtx).To(Equal(false))
			})
			It("accepts one argument not context", func() {
				goodHandler := func(a string) error {
					return nil
				}
				hasCtx, err := validateArguments(reflect.TypeOf(goodHandler))
				Expect(err).To(BeNil())
				Expect(hasCtx).To(Equal(false))
			})
			It("accepts two arguments when the first is context", func() {
				goodHandler := func(ctx context.Context, a string) error {
					return nil
				}
				hasCtx, err := validateArguments(reflect.TypeOf(goodHandler))
				Expect(err).To(BeNil())
				Expect(hasCtx).To(Equal(true))
			})
		})
	})
	Describe("validateReturns", func() {
		Context("Bad Handlers", func() {
			It("fails on too many returns", func() {
				badHandler := func() (string, error, string) {
					return "", nil, ""
				}
				err := validateReturns(reflect.TypeOf(badHandler))
				Expect(err).NotTo(Equal(nil))
			})
			It("fails if last return is not error", func() {
				badHandler := func() (string, string) {
					return "", ""
				}
				err := validateReturns(reflect.TypeOf(badHandler))
				Expect(err).NotTo(Equal(nil))
			})
			It("fails if retuns one thing that is not error", func() {
				badHandler := func() string {
					return ""
				}
				err := validateReturns(reflect.TypeOf(badHandler))
				Expect(err).NotTo(Equal(nil))
			})
		})
		Context("Good Handlers", func() {
			It("succeeds if no returns", func() {
				goodHandler := func() {
					return
				}
				err := validateReturns(reflect.TypeOf(goodHandler))
				Expect(err).To(BeNil())
			})
			It("suceeds if only returns error", func() {
				goodHandler := func() error {
					return nil
				}
				err := validateReturns(reflect.TypeOf(goodHandler))
				Expect(err).To(BeNil())
			})
			It("succeeds if returns something and error", func() {
				goodHandler := func() (string, error) {
					return "", nil
				}
				err := validateReturns(reflect.TypeOf(goodHandler))
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("makeGenericHandler", func() {
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

		Context("Bad Handlers", func() {
			It("fails if handlers return types are bad", func() {
				badHandler := func() (string, string) {
					return "", ""
				}
				generic := makeGenericHandler(badHandler)
				bg := context.Background()
				_, err := generic(bg, []byte{})

				Expect(err).NotTo(BeNil())
				Expect(len(exceptions)).To(BeNumerically("==", 1))
			})
			It("fails if handler is not a function", func() {
				badHandler := "not a function"
				generic := makeGenericHandler(badHandler)
				bg := context.Background()
				_, err := generic(bg, []byte{})

				Expect(err).NotTo(BeNil())
				Expect(len(exceptions)).To(BeNumerically("==", 1))
			})
			It("fails if handler is nil", func() {
				generic := makeGenericHandler(nil)
				bg := context.Background()
				_, err := generic(bg, []byte{})

				Expect(err).NotTo(BeNil())
				Expect(len(exceptions)).To(BeNumerically("==", 1))
			})
			It("fails if handlers arguments are bad", func() {
				badHandler := func(string, string) {}
				generic := makeGenericHandler(badHandler)
				bg := context.Background()
				_, err := generic(bg, []byte{})

				Expect(err).NotTo(BeNil())
				Expect(len(exceptions)).To(BeNumerically("==", 1))
			})
		})
		Context("Handler arguments are not json compatible to input", func() {
			type t1 struct {
				a1 string
				b1 string
			}
			type t2 struct {
				a2 int
			}
			It("fails when target type is not JSONable", func() {
				msg, err := json.Marshal(t1{a1: "hello", b1: "world"})
				if err != nil {
					Fail(fmt.Sprintf("Failed to marshal json %v", err))
				}
				type myFuncType func(int, int) error
				generic := makeGenericHandler(func(f myFuncType) { f(1, 2) })
				bg := context.Background()
				_, err = generic(bg, msg)
				Expect(err).NotTo(BeNil())
				Expect(len(exceptions)).To(BeNumerically("==", 1))
			})
			It("constructs available fields when not everything is available in json", func() {
				msg, err := json.Marshal(t1{b1: "world"})
				if err != nil {
					Fail(fmt.Sprintf("Failed to marshal json %v", err))
				}
				generic := makeGenericHandler(func(x t1) {})
				bg := context.Background()
				_, err = generic(bg, msg)
				Expect(err).To(BeNil())
				Expect(len(exceptions)).To(BeNumerically("==", 0))
			})
		})
	})
})
