package tracer_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
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

// FakeCollector implements a fake trace collector that will
// listen on an endpoint untill a trace is received and then will
// return that parsed trace
type FakeCollector struct {
	Endpoint string
}

// Listen on the endpoint for one trace and push it to outChannel
func (fc *FakeCollector) Listen(outChannel chan *protocol.Trace) {
	ln, err := net.Listen("tcp", fc.Endpoint)
	if err != nil {
		outChannel <- nil
		return
	}
	defer ln.Close()
	conn, err := ln.Accept()
	if err != nil {
		outChannel <- nil
		return
	}
	defer conn.Close()
	var buf = make([]byte, 0)
	_, err = conn.Read(buf)
	if err != nil {
		outChannel <- nil
		return
	}
	var receivedTrace protocol.Trace
	err = json.Unmarshal(buf, &receivedTrace)
	if err != nil {
		outChannel <- nil
		return
	}
	outChannel <- &receivedTrace
}

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
	})
	tracer.GlobalTracer.Start()
	defer tracer.StopGlobalTracer()
	operations()
}

// testWithTracer runs a test with
func testWithTracer(timeout *time.Duration, operations func()) *protocol.Trace {
	endpoint := "127.0.0.1:54769"
	traceChannel := make(chan *protocol.Trace)
	fc := FakeCollector{Endpoint: endpoint}
	go fc.Listen(traceChannel)
	go runWithTracer("http://"+endpoint, operations)
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
	defaultTimeout := time.Second * 100
	timeout := &defaultTimeout
	trace := testWithTracer(timeout, func() { epsagon.Label("test_key", "test_value") })
	println(trace)
}

func Test_AddError_sanity(t *testing.T) {
	defaultTimeout := time.Second * 100
	timeout := &defaultTimeout
	trace := testWithTracer(timeout, func() { epsagon.Error("some error") })
	println(trace)
}

func Test_AddTypeError(t *testing.T) {
	defaultTimeout := time.Second * 100
	timeout := &defaultTimeout
	trace := testWithTracer(timeout, func() { epsagon.TypeError("some error", "test error type") })
	println(trace)
}
