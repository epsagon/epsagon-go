package epsagon

import (
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
