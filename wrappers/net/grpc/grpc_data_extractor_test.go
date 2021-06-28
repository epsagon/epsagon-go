package epsagongrpc


import (
	"github.com/epsagon/epsagon-go/protocol"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestGetURL(t *testing.T) {
	testcases := []struct {
		method   string
		target   string
		expected string
	}{
		{
			method:   "/TestApplication/DoUnaryUnary",
			target:   "",
			expected: "grpc:///TestApplication/DoUnaryUnary",
		},
		{
			method:   "TestApplication/DoUnaryUnary",
			target:   "",
			expected: "grpc://TestApplication/DoUnaryUnary",
		},
		{
			method:   "/TestApplication/DoUnaryUnary",
			target:   ":8080",
			expected: "grpc://:8080/TestApplication/DoUnaryUnary",
		},
		{
			method:   "/TestApplication/DoUnaryUnary",
			target:   "localhost:8080",
			expected: "grpc://localhost:8080/TestApplication/DoUnaryUnary",
		},
		{
			method:   "TestApplication/DoUnaryUnary",
			target:   "localhost:8080",
			expected: "grpc://localhost:8080/TestApplication/DoUnaryUnary",
		},
		{
			method:   "/TestApplication/DoUnaryUnary",
			target:   "dns:///localhost:8080",
			expected: "grpc://localhost:8080/TestApplication/DoUnaryUnary",
		},
		{
			method:   "/TestApplication/DoUnaryUnary",
			target:   "unix:/path/to/socket",
			expected: "grpc://localhost/TestApplication/DoUnaryUnary",
		},
		{
			method:   "/TestApplication/DoUnaryUnary",
			target:   "unix:///path/to/socket",
			expected: "grpc://localhost/TestApplication/DoUnaryUnary",
		},
	}

	for _, test := range testcases {
		actual := getURL(test.method, test.target)
		if actual.String() != test.expected {
			t.Errorf("incorrect URL:\n\tmethod=%s,\n\ttarget=%s,\n\texpected=%s,\n\tactual=%s",
				test.method, test.target, test.expected, actual.String())
		}
	}
}

var _ = Describe("createGRPCEvent", func() {
	var (
		event      		*protocol.Event
		TestMethod 		string
	)
	BeforeEach(func() {

		TestMethod = "TestMethod"
	})

	Context("Test Create gRPC event sanity", func() {
		It("test sanity", func() {

			event = createGRPCEvent("trigger", TestMethod, "TestEventID")

			Expect(event.ErrorCode).To(Equal(protocol.ErrorCode_OK))
			Expect(event.Origin).To(Equal("trigger"))
			Expect(event.Resource.Type).To(Equal("grpc"))
			Expect(event.Resource.Operation).To(Equal(TestMethod))
			Expect(event.Resource.Metadata).To(Equal(map[string]string{}))
		})
	})
})