package internal

import (
	"encoding/json"
	"github.com/epsagon/epsagon-go/protocol"
	"net"
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
