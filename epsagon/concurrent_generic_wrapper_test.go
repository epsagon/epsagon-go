package epsagon_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	epsagonhttp "github.com/epsagon/epsagon-go/wrappers/net/http"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEpsagonConcurrentWrapper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multiple Traces")
}

func sendRequest(wg *sync.WaitGroup, path string, testServer *httptest.Server) {
	defer wg.Done()
	time.Sleep(time.Duration(rand.Intn(500)) * time.Microsecond)
	client := http.Client{}
	response, err := client.Get(testServer.URL + path)
	Expect(err).To(BeNil())
	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	Expect(err).To(BeNil())
	responseString := string(responseData)
	Expect(responseString).To(Equal(path))
}

func parseEventID(event *protocol.Event) (identifier int) {
	resourceName := event.Resource.GetName()
	resourceURL, err := url.Parse(resourceName)
	if err != nil {
		panic("failed to parse event URL - bad trace")
	}
	urlPath := resourceURL.RequestURI()
	identifier, err = strconv.Atoi(urlPath[1:])
	if err != nil {
		panic("failed to parse path - bad trace")
	}
	return
}

func waitForTraces(start int, end int, traceChannel chan *protocol.Trace, resourceName string, wg *sync.WaitGroup) {
	defer wg.Done()
	var trace *protocol.Trace
	receivedTraces := map[int]bool{}
	for i := start; i < end; i++ {
		receivedTraces[i] = false
	}
	ticker := time.NewTicker(8 * time.Second)
	for len(receivedTraces) > 0 {
		select {
		case trace = <-traceChannel:
			func() {
				Expect(len(trace.Events)).To(Equal(2))
				if len(resourceName) > 0 {
					Expect(trace.Events[1].Resource.Name).To(Equal(resourceName))
				}
				identifier := parseEventID(trace.Events[0])
				if identifier < start || identifier >= end {
					panic("received unexpected event")
				}
				_, exists := receivedTraces[identifier]
				if !exists {
					panic("received duplicated event")
				}
				delete(receivedTraces, identifier)
			}()
		case <-ticker.C:
			panic("timeout while receiving traces")
		}
	}
}

type HandlerFunc func(res http.ResponseWriter, req *http.Request)

func handleResponse(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	client := http.Client{Transport: epsagonhttp.NewTracingTransport(ctx)}
	client.Get(fmt.Sprintf("https://www.google.com%s", req.RequestURI))
	res.Write([]byte(req.RequestURI))
}

func createTestHTTPServer(config *epsagon.Config, resourceName string) *httptest.Server {
	var concurrentWrapper epsagon.GenericFunction
	if len(resourceName) > 0 {
		concurrentWrapper = epsagon.ConcurrentGoWrapper(
			config,
			handleResponse,
			resourceName,
		)
	} else {
		concurrentWrapper = epsagon.ConcurrentGoWrapper(
			config,
			handleResponse,
		)
	}
	return httptest.NewServer(http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			concurrentWrapper(res, req)
		},
	))
}

var _ = Describe("multiple_traces", func() {
	Describe("http_server_tests", func() {
		Context("Happy Flows", func() {
			var (
				traceCollectorServer *httptest.Server
				testServer           *httptest.Server
				config               *epsagon.Config
				traceChannel         chan *protocol.Trace
			)
			BeforeEach(func() {
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
			})
			AfterEach(func() {
				testServer.Close()
				traceCollectorServer.Close()
			})
			It("Multiple requests to test server", func() {
				resourceName := ""
				testServer = createTestHTTPServer(config, "")
				var wg sync.WaitGroup
				go waitForTraces(0, 50, traceChannel, resourceName, &wg)
				wg.Add(1)
				for i := 0; i < 50; i++ {
					wg.Add(1)
					go sendRequest(&wg, fmt.Sprintf("/%d", i), testServer)

				}
				wg.Wait()
				go waitForTraces(51, 100, traceChannel, resourceName, &wg)
				wg.Add(1)
				for i := 51; i < 100; i++ {
					wg.Add(1)
					go sendRequest(&wg, fmt.Sprintf("/%d", i), testServer)
				}
				wg.Wait()
			})
			It("Custom runner resource name", func() {
				resourceName := "test-resource-name"
				testServer = createTestHTTPServer(config, resourceName)
				var wg sync.WaitGroup
				go waitForTraces(0, 1, traceChannel, resourceName, &wg)
				wg.Add(1)
				for i := 0; i < 1; i++ {
					wg.Add(1)
					go sendRequest(&wg, fmt.Sprintf("/%d", i), testServer)

				}
				wg.Wait()
			})

		})
	})
})
