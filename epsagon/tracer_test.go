package epsagon

import (
	// "fmt"
	// protocol "github.com/epsagon/epsagon-go/protocol"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestEpsagonTracer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "epsagonTracer suite")
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
