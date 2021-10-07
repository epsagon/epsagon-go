package tracer_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"testing"
	"time"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEpsagonTracer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Epsagon Core Suite")
}

var _ = Describe("epsagonTracer suite", func() {
	Describe("Run/Stop", func() {
	})
	Describe("AddEvent", func() {
	})
	Describe("AddException", func() {
	})
	Describe("sendTraces", func() {
	})
})

func runWithTracer(endpoint string, operations func()) {
	tracer.CreateGlobalTracer(&tracer.Config{
		CollectorURL: endpoint,
		Token:        "1",
	})
	tracer.GlobalTracer.Start()
	defer tracer.StopGlobalTracer()
	operations()
}

// testWithTracer runs a test with
func testWithTracer(timeout *time.Duration, operations func()) *protocol.Trace {
	traceChannel := make(chan *protocol.Trace)
	fakeCollectorServer := httptest.NewServer(http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			buf, err := ioutil.ReadAll(req.Body)
			if err != nil {
				traceChannel <- nil
				return
			}
			var receivedTrace protocol.Trace
			err = json.Unmarshal(buf, &receivedTrace)
			if err != nil {
				traceChannel <- nil
				return
			}
			traceChannel <- &receivedTrace
			res.Write([]byte(""))
		},
	))
	go runWithTracer(fakeCollectorServer.URL, operations)
	if timeout == nil {
		defaultTimeout := time.Second * 10
		timeout = &defaultTimeout
	}
	timer := time.NewTimer(*timeout)
	select {
	case <-timer.C:
		return nil
	case trace := <-traceChannel:
		return trace
	}
}

type stubHTTPClient struct {
	httpClient *http.Client
	PostError  error
}

func (s stubHTTPClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	if s.PostError != nil {
		return nil, s.PostError
	}
	return s.httpClient.Post(url, contentType, body)
}

func Test_handleSendTracesResponse(t *testing.T) {
	tests := []struct {
		name          string
		apiResponse   string
		apiStatusCode int
		httpClient    stubHTTPClient
		expectedLog   string
	}{
		{
			name:          "No Log",
			apiResponse:   `{"test":"valid"}`,
			apiStatusCode: http.StatusOK,
			httpClient: stubHTTPClient{
				httpClient: &http.Client{Timeout: time.Duration(time.Second)},
			},
			expectedLog: "",
		},
		{
			name: "Error With No Response",
			httpClient: stubHTTPClient{
				httpClient: &http.Client{Timeout: time.Duration(time.Second)},
				PostError:  fmt.Errorf("Post http://not-valid-blackole.local.test: dial tcp: lookup not-valid-blackole.local.test: no such host"),
			},
			expectedLog: fmt.Sprintf("Error while sending traces \nPost http://not-valid-blackole.local.test: dial tcp: lookup not-valid-blackole.local.test: no such host"),
		},
		{
			name:        "Error With 5XX Response",
			apiResponse: `{"error":"failed to send traces"}`,
			httpClient: stubHTTPClient{
				httpClient: &http.Client{Timeout: time.Duration(time.Second)},
			},
			apiStatusCode: http.StatusInternalServerError,
			expectedLog:   fmt.Sprintf("Error while sending traces \n{\"error\":\"failed to send traces\"}"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//Read the logs to a buffer
			buf := bytes.Buffer{}
			log.SetOutput(&buf)
			defer func() {
				log.SetOutput(os.Stderr)
			}()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.apiStatusCode)
				w.Write([]byte(test.apiResponse))
			}))
			defer server.Close()
			resp, err := test.httpClient.Post(server.URL, "application/json", nil)
			tracer.HandleSendTracesResponse(resp, err)

			if !strings.Contains(buf.String(), test.expectedLog) {
				t.Errorf("Unexpected log: expected %s got %s", test.expectedLog, buf.String())
			}

		})
	}
}

func Test_AddLabel_sanity(t *testing.T) {
	defaultTimeout := time.Second * 5
	timeout := &defaultTimeout
	trace := testWithTracer(timeout, func() { epsagon.Label("test_key", "test_value") })
	Expect(trace).ToNot(BeNil())
}

func Test_AddError_sanity(t *testing.T) {
	defaultTimeout := time.Second * 5
	timeout := &defaultTimeout
	trace := testWithTracer(timeout, func() { epsagon.Error("some error") })
	Expect(trace).ToNot(BeNil())
}

func Test_AddTypeError(t *testing.T) {
	defaultTimeout := time.Second * 5
	timeout := &defaultTimeout
	trace := testWithTracer(timeout, func() { epsagon.TypeError("some error", "test error type") })
	Expect(trace).ToNot(BeNil())
}

func Test_MaxTraceSize_sanity(t *testing.T) {
	os.Setenv(tracer.MaxTraceSizeEnvVar, "2048")
	defaultTimeout := time.Second * 5
	timeout := &defaultTimeout
	trace := testWithTracer(timeout, func() { epsagon.Label("1", "2") })
	Expect(trace).ToNot(BeNil())
	os.Setenv(tracer.MaxTraceSizeEnvVar, "64")
	trace = testWithTracer(timeout, func() { epsagon.Label("1", "2") })
	Expect(trace).To(BeNil())
}
