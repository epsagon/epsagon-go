package epsagon_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	epsagonhttp "github.com/epsagon/epsagon-go/wrappers/net/http"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const TestLabelKey = "random_key"

func TestEpsagonCustomTraceFields(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Custom trace fields tests")
}

func waitForTrace(traceChannel chan *protocol.Trace, resourceName string, eventsCount int) *protocol.Event {
	var trace *protocol.Trace
	receivedTrace := false
	ticker := time.NewTicker(3 * time.Second)
	for !receivedTrace {
		select {
		case trace = <-traceChannel:
			func() {
				Expect(len(trace.Events)).To(Equal(eventsCount))
				if len(resourceName) > 0 {
					Expect(trace.Events[eventsCount-1].Resource.Name).To(Equal(resourceName))
				}
				receivedTrace = true
			}()
		case <-ticker.C:
			panic("timeout while receiving trace")
		}
	}
	return trace.Events[eventsCount-1]
}

func getRunnerLabels(runner *protocol.Event) map[string]interface{} {
	labels, ok := runner.Resource.Metadata[tracer.LabelsKey]
	Expect(ok).To(BeTrue())
	var labelsMap map[string]interface{}
	err := json.Unmarshal([]byte(labels), &labelsMap)
	Expect(err).To(BeNil())
	return labelsMap
}

func verifyException(errorType string, errorMessage string, exception *protocol.Exception) {
	Expect(errorType).To(Equal(exception.Type))
	Expect(errorMessage).To(Equal(exception.Message))
}

func verifyLabelValue(key string, value interface{}, labelsMap map[string]interface{}) {
	labelValue, ok := labelsMap[key]
	Expect(ok).To(BeTrue())
	switch value.(type) {
	case int:
		Expect(int(labelValue.(float64))).To(Equal(value))
	default:
		Expect(labelValue).To(Equal(value))
	}
}

var _ = Describe("Custom trace fields", func() {
	Describe("ut_tests", func() {
		Context("Happy Flows", func() {
			var (
				traceCollectorServer *httptest.Server
				config               *epsagon.Config
				traceChannel         chan *protocol.Trace
			)
			BeforeEach(func() {
				tracer.GlobalTracer = nil
				traceChannel = make(chan *protocol.Trace)
				traceCollectorServer = httptest.NewServer(http.HandlerFunc(
					func(res http.ResponseWriter, req *http.Request) {
						buf, err := ioutil.ReadAll(req.Body)
						if err != nil {
							panic(err)
						}
						var receivedTrace protocol.Trace
						err = json.Unmarshal(buf, &receivedTrace)
						if err != nil {
							panic(err)
						}
						traceChannel <- &receivedTrace
						res.Write([]byte(""))
					},
				))
				config = epsagon.NewTracerConfig("test", "test token")
				config.CollectorURL = traceCollectorServer.URL
				config.MetadataOnly = false
			})
			AfterEach(func() {
				traceCollectorServer.Close()
			})
			It("Test custom label with integer value", func() {
				resourceName := "test-resource-name"
				value := 5
				epsagon.GoWrapper(
					config,
					func() {
						epsagon.Label(TestLabelKey, value)
					},
					resourceName,
				)()
				runnerEvent := waitForTrace(traceChannel, resourceName, 1)
				labelsMap := getRunnerLabels(runnerEvent)
				Expect(len(labelsMap)).To(Equal(1))
				verifyLabelValue(TestLabelKey, value, labelsMap)
			})
			It("Test custom label with float value", func() {
				resourceName := "test-resource-name"
				var value float64 = 6
				epsagon.GoWrapper(
					config,
					func() {
						epsagon.Label(TestLabelKey, value)
					},
					resourceName,
				)()
				runnerEvent := waitForTrace(traceChannel, resourceName, 1)
				labelsMap := getRunnerLabels(runnerEvent)
				Expect(len(labelsMap)).To(Equal(1))
				verifyLabelValue(TestLabelKey, value, labelsMap)
			})
			It("Test custom label with bool value", func() {
				resourceName := "test-resource-name"
				value := false
				epsagon.GoWrapper(
					config,
					func() {
						epsagon.Label(TestLabelKey, value)
					},
					resourceName,
				)()
				runnerEvent := waitForTrace(traceChannel, resourceName, 1)
				labelsMap := getRunnerLabels(runnerEvent)
				Expect(len(labelsMap)).To(Equal(1))
				verifyLabelValue(TestLabelKey, value, labelsMap)
			})
			It("Test custom label with string value", func() {
				resourceName := "test-resource-name"
				value := "test_value"
				epsagon.GoWrapper(
					config,
					func() {
						epsagon.Label(TestLabelKey, value)
					},
					resourceName,
				)()
				runnerEvent := waitForTrace(traceChannel, resourceName, 1)
				labelsMap := getRunnerLabels(runnerEvent)
				Expect(len(labelsMap)).To(Equal(1))
				verifyLabelValue(TestLabelKey, value, labelsMap)
			})
			It("Test multiple custom labels", func() {
				resourceName := "test-resource-name"
				value := "test_value"
				secondLabelKey := "second_key"
				secondLabelValue := 5
				epsagon.GoWrapper(
					config,
					func() {
						epsagon.Label(TestLabelKey, value)
						epsagon.Label(secondLabelKey, secondLabelValue)
					},
					resourceName,
				)()
				runnerEvent := waitForTrace(traceChannel, resourceName, 1)
				labelsMap := getRunnerLabels(runnerEvent)
				Expect(len(labelsMap)).To(Equal(2))
				verifyLabelValue(TestLabelKey, value, labelsMap)
				verifyLabelValue(secondLabelKey, secondLabelValue, labelsMap)
			})
			It("Test invalid label value", func() {
				resourceName := "test-resource-name"
				type NotSupportType struct {
					x int
				}
				value := NotSupportType{}
				epsagon.GoWrapper(
					config,
					func() {
						epsagon.Label(TestLabelKey, value)
					},
					resourceName,
				)()
				runnerEvent := waitForTrace(traceChannel, resourceName, 1)
				labelsMap := getRunnerLabels(runnerEvent)
				Expect(len(labelsMap)).To(Equal(0))
			})
			It("Test too big labels size", func() {
				resourceName := "test-resource-name"
				value := "test_value"
				epsagon.GoWrapper(
					config,
					func() {
						epsagon.Label(TestLabelKey, value)
						letterBytes := "abc"
						b := make([]byte, tracer.MaxLabelsSize)
						for i := range b {
							b[i] = letterBytes[rand.Intn(len(letterBytes))]
						}
						bigValue := string(b)
						epsagon.Label("big label value", bigValue)
					},
					resourceName,
				)()
				runnerEvent := waitForTrace(traceChannel, resourceName, 1)
				labelsMap := getRunnerLabels(runnerEvent)
				Expect(len(labelsMap)).To(Equal(1))
				verifyLabelValue(TestLabelKey, value, labelsMap)
			})
			It("Default custom error - string error message", func() {
				resourceName := "test-resource-name"
				errorMessage := "test_value"
				epsagon.GoWrapper(
					config,
					func() {
						epsagon.Error(errorMessage)
					},
					resourceName,
				)()
				runnerEvent := waitForTrace(traceChannel, resourceName, 1)
				Expect(runnerEvent.ErrorCode).To(Equal(protocol.ErrorCode_OK))
				exception := runnerEvent.Exception
				verifyException(epsagon.DefaultErrorType, errorMessage, exception)
			})
			It("Default custom error - Error input", func() {
				resourceName := "test-resource-name"
				errorMessage := "test_value"
				epsagon.GoWrapper(
					config,
					func() {
						epsagon.Error(errors.New(errorMessage))
					},
					resourceName,
				)()
				runnerEvent := waitForTrace(traceChannel, resourceName, 1)
				Expect(runnerEvent.ErrorCode).To(Equal(protocol.ErrorCode_OK))
				exception := runnerEvent.Exception
				verifyException(epsagon.DefaultErrorType, errorMessage, exception)
			})
			It("Custom error & type - string error message", func() {
				resourceName := "test-resource-name"
				errorMessage := "test_value"
				errorType := "test error type"
				epsagon.GoWrapper(
					config,
					func() {
						epsagon.TypeError(errorMessage, errorType)
					},
					resourceName,
				)()
				runnerEvent := waitForTrace(traceChannel, resourceName, 1)
				Expect(runnerEvent.ErrorCode).To(Equal(protocol.ErrorCode_OK))
				exception := runnerEvent.Exception
				verifyException(errorType, errorMessage, exception)
			})
			It("Custom error & type - Error input", func() {
				resourceName := "test-resource-name"
				errorMessage := "test_value"
				errorType := "test error type"
				epsagon.GoWrapper(
					config,
					func() {
						epsagon.TypeError(errors.New(errorMessage), errorType)
					},
					resourceName,
				)()
				runnerEvent := waitForTrace(traceChannel, resourceName, 1)
				Expect(runnerEvent.ErrorCode).To(Equal(protocol.ErrorCode_OK))
				exception := runnerEvent.Exception
				verifyException(errorType, errorMessage, exception)
			})
			It("Custom label - no tracer", func() {
				value := 3
				output := func() int {
					epsagon.Label("a", 3)
					return value
				}()
				Expect(output).To(Equal(value))
			})
			It("Custom error - no tracer", func() {
				errorMessage := "test_value"
				value := 3
				output := func() int {
					epsagon.Error(errorMessage)
					return value
				}()
				Expect(output).To(Equal(value))
			})
			It("Custom type error - no tracer", func() {
				errorMessage := "test_value"
				value := 3
				output := func() int {
					epsagon.TypeError("a", errorMessage)
					return value
				}()
				Expect(output).To(Equal(value))
			})
			It("Trimmed event metadata fields", func() {
				resourceName := "test-resource-name"
				epsagon.GoWrapper(
					config,
					func() {
						letterBytes := "abc"
						b := make([]byte, tracer.MaxTraceSize*2)
						for i := range b {
							b[i] = letterBytes[rand.Intn(len(letterBytes))]
						}
						testServer := httptest.NewServer(http.HandlerFunc(
							func(res http.ResponseWriter, req *http.Request) {
								buf, err := ioutil.ReadAll(req.Body)
								if err != nil {
									panic(err)
								}
								res.Write(buf)
							},
						))
						defer testServer.Close()
						for i := 0; i < 10; i++ {
							client := http.Client{Transport: epsagonhttp.NewTracingTransport()}
							req, err := http.NewRequest(http.MethodGet, testServer.URL, bytes.NewReader(b))
							if err != nil {
								panic(err)
							}
							_, err = client.Do(req)
							if err != nil {
								panic(err)
							}
						}
					},
					resourceName,
				)()
				runnerEvent := waitForTrace(traceChannel, resourceName, 11)
				Expect(runnerEvent.ErrorCode).To(Equal(protocol.ErrorCode_OK))
				isTrimmed, ok := runnerEvent.Resource.Metadata[tracer.IsTrimmedKey]
				Expect(ok).To(BeTrue())
				Expect(isTrimmed).To(Equal("true"))
			})
		})
	})
})
