package epsagon

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/wrappers/net/http"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"
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
	Expect(err == nil).To(Equal(true))
	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	Expect(err == nil).To(Equal(true))
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

func waitForTraces(start int, end int, traceChannel chan *protocol.Trace, wg *sync.WaitGroup) {
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

var _ = Describe("multiple_traces", func() {
	Describe("http_server_tests", func() {
		Context("Happy Flows", func() {
			var (
				traceCollectorServer *httptest.Server
				testServer           *httptest.Server
				config               *Config
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
				config = NewTracerConfig("test", "test token")
				config.CollectorURL = traceCollectorServer.URL
				testServer = httptest.NewServer(http.HandlerFunc(
					func(res http.ResponseWriter, req *http.Request) {
						ConcurrentGoWrapper(
							config,
							func(ctx context.Context, res http.ResponseWriter, req *http.Request) {

								client := epsagonhttp.Wrap(http.Client{}, ctx)
								client.Get(fmt.Sprintf("https://www.google.com%s", req.RequestURI))
								res.Write([]byte(req.RequestURI))
							},
						)(res, req)
					},
				))

			})
			AfterEach(func() {
				testServer.Close()
				traceCollectorServer.Close()
			})
			It("Multiple requests to test server", func() {
				var wg sync.WaitGroup
				go waitForTraces(0, 50, traceChannel, &wg)
				wg.Add(1)
				for i := 0; i < 50; i++ {
					wg.Add(1)
					go sendRequest(&wg, fmt.Sprintf("/%d", i), testServer)

				}
				wg.Wait()
				go waitForTraces(51, 100, traceChannel, &wg)
				wg.Add(1)
				for i := 51; i < 100; i++ {
					wg.Add(1)
					go sendRequest(&wg, fmt.Sprintf("/%d", i), testServer)
				}
				wg.Wait()
			})
		})
	})
})
