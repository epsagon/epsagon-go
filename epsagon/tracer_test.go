package epsagon

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	// "fmt"
	"github.com/epsagon/epsagon-go/internal"
	"github.com/epsagon/epsagon-go/protocol"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"time"
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
	CreateTracer(&Config{
		CollectorURL: endpoint,
	})
	defer StopTracer()
	operations()
}

// testWithTracer runs a test with
func testWithTracer(timeout *time.Duration, operations func()) *protocol.Trace {
	endpoint := "localhost:547698"
	traceChannel := make(chan *protocol.Trace)
	fc := internal.FakeCollector{Endpoint: endpoint}
	go fc.Listen(traceChannel)
	go runWithTracer(endpoint, operations)
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


func Test_handleSendTracesResponse(t *testing.T) {
	tests := []struct {
		name          string
		apiResponse   string
		apiStatusCode int
		customURL     string
		expectedLog   string
	}{
		{
			name:          "No Log",
			apiResponse:   `{"test":"valid"}`,
			apiStatusCode: http.StatusOK,
			expectedLog:   "",
		},
		{
			name:        "Error With No Response",
			customURL:   "http://not-valid-blackole.local.test",
			expectedLog: fmt.Sprintf("Error while sending traces \nPost http://not-valid-blackole.local.test: dial tcp: lookup not-valid-blackole.local.test: no such host"),
		},
		{
			name:          "Error With 5XX Response",
			apiResponse:   `{"error":"failed to send traces"}`,
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
			url := server.URL
			if test.customURL != "" {
				url = test.customURL
			}
			client := &http.Client{Timeout: time.Duration(time.Second)}
			resp, err := client.Post(url, "application/json", nil)
			handleSendTracesResponse(resp, err)

			if !strings.Contains(buf.String(), test.expectedLog) {
				t.Errorf("Unexpected log: expected %s got %s", test.expectedLog, buf.String())
			}

		})
	}
}

