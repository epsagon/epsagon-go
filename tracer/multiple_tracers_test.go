package tracer

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
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

func validateTraceExists() {
	mutex.Lock()
	defer mutex.Unlock()
	current_tracer, _ := getCurrentTracerInfo()
	Expect(current_tracer != nil).To(Equal(true))
	Expect(current_tracer.Stopped()).To(Equal(false))
}

var _ = Describe("multiple_traces", func() {
	Describe("http_server_tests", func() {
		Context("Happy Flows", func() {
			var (
				testServer *httptest.Server
			)
			BeforeEach(func() {
				epsagon.SwitchToMultipleTraces()
				testServer = httptest.NewServer(http.HandlerFunc(
					func(res http.ResponseWriter, req *http.Request) {
						epsagon.GoWrapper(
							nil,
							func(res http.ResponseWriter, req *http.Request) {
								// validate a new Trace has been created for current goroutine ID
								validateTraceExists()
								res.Write([]byte(req.RequestURI))
							},
						)(res, req)
					},
				))

			})
			AfterEach(func() {
				testServer.Close()
			})
			It("Multiple requests to test server", func() {
				var wg sync.WaitGroup
				for i := 0; i < 30; i++ {
					wg.Add(1)
					go sendRequest(&wg, fmt.Sprintf("/%d", i), testServer)

				}
				wg.Wait()
				time.Sleep(3 * time.Second)
				mutex.Lock()
				Expect(0).To(Equal(len(Tracers)))
				mutex.Unlock()
				for i := 90; i < 100; i++ {
					wg.Add(1)
					go sendRequest(&wg, fmt.Sprintf("/%d", i), testServer)
				}
				wg.Wait()
				time.Sleep(1 * time.Second)
				mutex.Lock()
				Expect(0).To(Equal(len(Tracers)))
				mutex.Unlock()
			})
		})
	})
})
