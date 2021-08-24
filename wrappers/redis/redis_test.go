package epsagonredis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	epsagonredis "github.com/epsagon/epsagon-go/wrappers/redis"
	"github.com/go-redis/redis/v8"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
)

func TestRedisWrapper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Redis Wrapper")
}

var _ = Describe("Redis wrapper", func() {
	var (
		events      []*protocol.Event
		client      *redis.Client
		ctx         context.Context
		tracerMock  *tracer.MockedEpsagonTracer
		redisServer *miniredis.Miniredis
	)

	BeforeEach(func() {
		events = make([]*protocol.Event, 0)
		tracerMock = &tracer.MockedEpsagonTracer{
			Events: &events,
			Config: &tracer.Config{
				Disable:      true,
				TestMode:     true,
				MetadataOnly: true,
			},
		}
		ctx = epsagon.ContextWithTracer(tracerMock)

		server, err := miniredis.Run()
		if err != nil {
			panic(err)
		} else {
			redisServer = server
		}

		client = epsagonredis.NewClient(&redis.Options{
			Addr:     redisServer.Addr(),
			Password: "",
			DB:       0,
		}, ctx)
	})

	AfterEach(func() {
		redisServer.Close()
	})

	Context("Single operation", func() {
		It("Adds event, MetadataOnly=true", func() {
			tracerMock.Config.MetadataOnly = true
			const (
				key   = "the_key"
				value = "the_value"
			)
			err := redisServer.Set(key, value)
			if err != nil {
				Fail("Test setup failed")
			}

			result := client.Get(ctx, key)

			Expect(result.Val()).To(Equal(value))
			Expect(len(events)).To(Equal(1))
			Expect(*events[0]).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Id":        ContainSubstring("redis-"),
				"StartTime": BeNumerically(">", 0),
				"Duration":  BeNumerically(">", 0),
				"ErrorCode": Equal(protocol.ErrorCode_OK),
				"Exception": BeNil(),
			}))
			Expect(*events[0].Resource).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Name":      Equal(redisServer.Host()),
				"Type":      Equal("redis"),
				"Operation": Equal("get"),
				"Metadata": gstruct.MatchAllKeys(gstruct.Keys{
					"Redis Host":     Equal(redisServer.Host()),
					"Redis Port":     Equal(redisServer.Port()),
					"Redis DB Index": Equal("0"),
				}),
			}))
			Expect(events[0].Resource.Metadata["Command Arguments"]).To(BeEmpty())
			Expect(events[0].Resource.Metadata["redis.response"]).To(BeEmpty())
		})

		It("Adds event, MetadataOnly=false", func() {
			tracerMock.Config.MetadataOnly = false
			const (
				key   = "the_key"
				value = "the_value"
			)
			err := redisServer.Set(key, value)
			if err != nil {
				Fail("Test setup failed")
			}

			result := client.Get(ctx, key)

			Expect(result.Val()).To(Equal(value))
			Expect(len(events)).To(Equal(1))
			Expect(*events[0]).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Id":        ContainSubstring("redis-"),
				"StartTime": BeNumerically(">", 0),
				"Duration":  BeNumerically(">", 0),
				"ErrorCode": Equal(protocol.ErrorCode_OK),
				"Exception": BeNil(),
			}))

			cmdArgs, _ := json.Marshal([]string{"get", key})
			Expect(*events[0].Resource).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Name":      Equal(redisServer.Host()),
				"Type":      Equal("redis"),
				"Operation": Equal("get"),
				"Metadata": gstruct.MatchAllKeys(gstruct.Keys{
					"Redis Host":        Equal(redisServer.Host()),
					"Redis Port":        Equal(redisServer.Port()),
					"Redis DB Index":    Equal("0"),
					"Command Arguments": Equal(string(cmdArgs)),
					"redis.response":    Equal(fmt.Sprintf("get %s: %s", key, value)),
				}),
			}))
		})

		It("Adds multiple events", func() {
			const (
				key   = "the_key"
				value = "the_value"
			)
			err := redisServer.Set(key, value)
			if err != nil {
				Fail("Test setup failed")
			}

			getResult1 := client.Get(ctx, key)
			getResult2 := client.Get(ctx, key)

			Expect(getResult1.Val()).To(Equal(value))
			Expect(getResult2.Val()).To(Equal(value))
			Expect(len(events)).To(Equal(2))

			for _, event := range events {
				Expect(*event).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Id":        ContainSubstring("redis-"),
					"StartTime": BeNumerically(">", 0),
					"Duration":  BeNumerically(">", 0),
					"ErrorCode": Equal(protocol.ErrorCode_OK),
					"Exception": BeNil(),
				}))
				Expect(*event.Resource).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Name":      Equal(redisServer.Host()),
					"Type":      Equal("redis"),
					"Operation": Equal("get"),
					"Metadata": gstruct.MatchAllKeys(gstruct.Keys{
						"Redis Host":     Equal(redisServer.Host()),
						"Redis Port":     Equal(redisServer.Port()),
						"Redis DB Index": Equal("0"),
					}),
				}))
				Expect(event.Resource.Metadata["Command Arguments"]).To(BeEmpty())
				Expect(event.Resource.Metadata["redis.response"]).To(BeEmpty())
			}
		})

		It("Adds error event", func() {
			const (
				key           = "the_key"
				value         = "the_value"
				expectedError = "ERR value is not an integer or out of range"
			)
			err := redisServer.Set(key, value)
			if err != nil {
				Fail("Test setup failed")
			}

			// trying to increment string value by one
			result := client.Incr(ctx, key)

			Expect(result.Err().Error()).To(Equal(expectedError))
			Expect(len(events)).To(Equal(1))
			Expect(*events[0]).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Id":        ContainSubstring("redis-"),
				"StartTime": BeNumerically(">", 0),
				"Duration":  BeNumerically(">", 0),
				"ErrorCode": Equal(protocol.ErrorCode_EXCEPTION),
			}))
			Expect(*events[0].Exception).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Message":   Equal(expectedError),
				"Time":      BeNumerically(">", 0),
				"Traceback": Not(BeEmpty()),
			}))
			Expect(*events[0].Resource).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Name":      Equal(redisServer.Host()),
				"Type":      Equal("redis"),
				"Operation": Equal("incr"),
				"Metadata": gstruct.MatchAllKeys(gstruct.Keys{
					"Redis Host":     Equal(redisServer.Host()),
					"Redis Port":     Equal(redisServer.Port()),
					"Redis DB Index": Equal("0"),
				}),
			}))
			Expect(events[0].Resource.Metadata["Command Arguments"]).To(BeEmpty())
			Expect(events[0].Resource.Metadata["redis.response"]).To(BeEmpty())
		})
	})

	Context("Pipeline operations", func() {
		It("Adds pipeline event, MetadataOnly=true", func() {
			tracerMock.Config.MetadataOnly = true
			const (
				key                 = "the_key"
				value               = "the_value"
				expectedSetResponse = "set the_key the_value: OK"
				expectedGetResponse = "get the_key: the_value"
			)

			pipe := client.Pipeline()
			pipe.Set(ctx, key, value, 0)
			pipe.Get(ctx, key)
			result, err := pipe.Exec(ctx)
			if err != nil {
				Fail("Failed pipeline test setup")
			}

			Expect(result[0].String()).To(Equal(expectedSetResponse))
			Expect(result[1].String()).To(Equal(expectedGetResponse))
			Expect(len(events)).To(Equal(1))
			Expect(*events[0]).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Id":        ContainSubstring("redis-"),
				"StartTime": BeNumerically(">", 0),
				"Duration":  BeNumerically(">", 0),
				"ErrorCode": Equal(protocol.ErrorCode_OK),
				"Exception": BeNil(),
			}))
			Expect(*events[0].Resource).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Name":      Equal(redisServer.Host()),
				"Type":      Equal("redis"),
				"Operation": Equal("Pipeline"),
				"Metadata": gstruct.MatchAllKeys(gstruct.Keys{
					"Redis Host":     Equal(redisServer.Host()),
					"Redis Port":     Equal(redisServer.Port()),
					"Redis DB Index": Equal("0"),
				}),
			}))
			Expect(events[0].Resource.Metadata["Command Arguments"]).To(BeEmpty())
			Expect(events[0].Resource.Metadata["redis.response"]).To(BeEmpty())
		})

		It("Adds pipeline event, MetadataOnly=false", func() {
			tracerMock.Config.MetadataOnly = false
			const (
				key                 = "the_key"
				value               = "the_value"
				expectedSetResponse = "set the_key the_value: OK"
				expectedGetResponse = "get the_key: the_value"
			)

			pipe := client.Pipeline()
			pipe.Set(ctx, key, value, 0)
			pipe.Get(ctx, key)
			result, err := pipe.Exec(ctx)
			if err != nil {
				Fail("Failed pipeline test setup")
			}

			Expect(result[0].String()).To(Equal(expectedSetResponse))
			Expect(result[1].String()).To(Equal(expectedGetResponse))
			Expect(len(events)).To(Equal(1))
			Expect(*events[0]).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Id":        ContainSubstring("redis-"),
				"StartTime": BeNumerically(">", 0),
				"Duration":  BeNumerically(">", 0),
				"ErrorCode": Equal(protocol.ErrorCode_OK),
				"Exception": BeNil(),
			}))

			cmdArgs, _ := json.Marshal([][]string{
				{"set", key, value},
				{"get", key},
			})
			redisResponse, _ := json.Marshal([]string{
				fmt.Sprintf("set %s %s: OK", key, value),
				fmt.Sprintf("get %s: %s", key, value),
			})
			Expect(*events[0].Resource).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Name":      Equal(redisServer.Host()),
				"Type":      Equal("redis"),
				"Operation": Equal("Pipeline"),
				"Metadata": gstruct.MatchAllKeys(gstruct.Keys{
					"Redis Host":        Equal(redisServer.Host()),
					"Redis Port":        Equal(redisServer.Port()),
					"Redis DB Index":    Equal("0"),
					"Command Arguments": Equal(string(cmdArgs)),
					"redis.response":    Equal(string(redisResponse)),
				}),
			}))
		})

		It("Adds pipeline error event", func() {
			const (
				key           = "the_key"
				value         = "the_value"
				expectedError = "ERR value is not an integer or out of range"
			)

			pipe := client.Pipeline()
			pipe.Set(ctx, key, value, 0)
			// trying to increment string value by one
			pipe.Incr(ctx, key)
			pipe.Exec(ctx)

			Expect(len(events)).To(Equal(1))
			Expect(*events[0]).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Id":        ContainSubstring("redis-"),
				"StartTime": BeNumerically(">", 0),
				"Duration":  BeNumerically(">", 0),
				"ErrorCode": Equal(protocol.ErrorCode_EXCEPTION),
			}))

			errorMessage, _ := json.Marshal([]string{expectedError})
			Expect(*events[0].Exception).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Message":   Equal(string(errorMessage)),
				"Time":      BeNumerically(">", 0),
				"Traceback": Not(BeEmpty()),
			}))
			Expect(*events[0].Resource).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Name":      Equal(redisServer.Host()),
				"Type":      Equal("redis"),
				"Operation": Equal("Pipeline"),
				"Metadata": gstruct.MatchAllKeys(gstruct.Keys{
					"Redis Host":     Equal(redisServer.Host()),
					"Redis Port":     Equal(redisServer.Port()),
					"Redis DB Index": Equal("0"),
				}),
			}))
			Expect(events[0].Resource.Metadata["Command Arguments"]).To(BeEmpty())
			Expect(events[0].Resource.Metadata["redis.response"]).To(BeEmpty())
		})
	})
})
