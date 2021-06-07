package epsagongrpc

import (
	"bytes"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

func TestUnaryClientInterceptor (t* testing.T) {
}

var _ = Describe("gRPC Client Wrapper", func ()  {

		BeforeEach(func() {
			config = &epsagon.Config{Config: tracer.Config{
				Disable:  true,
				TestMode: true,
			}}

			events = make([]*protocol.Event, 0)
			exceptions = make([]*protocol.Exception, 0)

			tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
				Events:     &events,
				Exceptions: &exceptions,
				Labels:     make(map[string]interface{}),
				Config:     &config.Config,
			}

			body := []byte("hello")
			request = httptest.NewRequest("POST", "https://www.help.com", ioutil.NopCloser(bytes.NewReader(body)))
			reply  = httptest.NewRecorder()
		})

})