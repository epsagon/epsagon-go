package tracer_test

import (
	"encoding/json"
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/wrappers/net/http"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestEpsagonMultipleTracers(t *testing.T) {
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

func waitForTraces(start uint, end uint, traceChannel chan *protocol.Trace, wg *sync.WaitGroup) {
	defer wg.Done()
	var trace *protocol.Trace
	ticker := time.NewTicker(10 * time.Second)
	count := 0
	for {
		select {
		case trace = <-traceChannel:
			func() {
				log.Printf("test")
				Expect(len(trace.Events)).To(Equal(2))
				count += 1
			}()
		case <-ticker.C:
			panic("timeout while receiving traces")
		}
	}
	log.Println("Got %d", count)
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
						//res.Write([]byte(req.RequestURI))
						log.Printf("boom trah")
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
						//Expect(len(receivedTrace.Events)).To(Equal(2))

						res.Write([]byte(""))
					},
				))
				config = epsagon.NewTracerConfig("test", "test token")
				config.CollectorURL = traceCollectorServer.URL
				epsagon.SwitchToMultipleTraces()
				testServer = httptest.NewServer(http.HandlerFunc(
					func(res http.ResponseWriter, req *http.Request) {
						epsagon.GoWrapper(
							config,
							func(res http.ResponseWriter, req *http.Request) {

								client := epsagonhttp.Wrap(http.Client{})
								client.Get(fmt.Sprintf("https://www.google.com/%s", req.RequestURI))
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
				for i := 0; i < 50; i++ {
					wg.Add(1)
					go sendRequest(&wg, fmt.Sprintf("/%d", i), testServer)

				}
				wg.Wait()
				//go waitForTraces(0, 50, traceChannel)
				//time.Sleep(3 * time.Second)
				go waitForTraces(51, 100, traceChannel, &wg)
				for i := 51; i < 100; i++ {
					wg.Add(1)
					go sendRequest(&wg, fmt.Sprintf("/%d", i), testServer)
				}
				wg.Wait()
				//time.Sleep(1 * time.Second)
				//waitForTraces(51, 100, traceChannel, &wg)
			})
		})
	})
})
