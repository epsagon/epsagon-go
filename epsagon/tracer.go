package epsagon

import (
	"bytes"
	protocol "github.com/epsagon/epsagon-go/protocol"
	"github.com/golang/protobuf/jsonpb"
	"io"
	"log"
	"net/http"
	"runtime"
	"sync"
)

var (
	mutex        sync.Mutex
	globalTracer *epsagonTracer
)

type epsagonTracer struct {
	appName      string
	token        string
	collectorURL string
	metadataOnly bool

	eventsPipe     chan *protocol.Event
	events         []*protocol.Event
	exceptionsPipe chan *protocol.Exception
	exceptions     []*protocol.Exception

	closeCmd chan struct{}
	stopped  chan struct{}
}

func (tracer *epsagonTracer) sendTraces() {
	tracesReader, err := tracer.getTraceReader()
	if err != nil {
		// TODO create an exception and send a trace only with that
		log.Printf("Encountered an error while marshaling the traces: %v\n", err)
		log.Println("failed to Marshal json")
		return
	}
	resp, err := http.Post(tracer.collectorURL, "application/json", tracesReader)
	if err != nil {
		var respBody []byte
		resp.Body.Read(respBody)
		resp.Body.Close()
		log.Printf("Error while sending traces \n%v\n%v\n", err, respBody)
	}
}

func (tracer *epsagonTracer) getTraceReader() (io.Reader, error) {
	version := runtime.Version()
	trace := protocol.Trace{
		AppName:    tracer.appName,
		Token:      tracer.token,
		Events:     tracer.events,
		Exceptions: tracer.exceptions,
		Version:    "0.0.1",
		Platform:   version,
	}
	marshaler := jsonpb.Marshaler{
		EnumsAsInts: true, EmitDefaults: true, OrigName: true}
	traceJSON, err := marshaler.MarshalToString(&trace)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer([]byte(traceJSON)), nil
}

func (tracer *epsagonTracer) closed() bool {
	select {
	case <-tracer.stopped:
		return true
	default:
		return false
	}
}

// CreateTracer will initiallize a global epsagon tracer
func CreateTracer(appName, token, collectorURL string, metadataOnly bool) {
	mutex.Lock()
	defer mutex.Unlock()
	if globalTracer != nil && !globalTracer.closed() {
		log.Println("The tracer is already created")
		return
	}
	globalTracer = &epsagonTracer{
		appName:        appName,
		token:          token,
		collectorURL:   collectorURL,
		metadataOnly:   metadataOnly,
		eventsPipe:     make(chan *protocol.Event),
		events:         make([]*protocol.Event, 0, 0),
		exceptionsPipe: make(chan *protocol.Exception),
		exceptions:     make([]*protocol.Exception, 0, 0),
		closeCmd:       make(chan struct{}),
		stopped:        make(chan struct{}),
	}
	go globalTracer.worker()
}

// AddException adds a tracing exception to the tracer
func (tracer *epsagonTracer) AddException(exception *protocol.Exception) {
	tracer.exceptionsPipe <- exception
}

// AddEvent adds an event to the tracer
func (tracer *epsagonTracer) AddEvent(event *protocol.Event) {
	tracer.eventsPipe <- event
}

// AddEvent adds an event to the tracer
func AddEvent(event *protocol.Event) {
	if globalTracer == nil || globalTracer.closed() {
		// TODO
		log.Println("The tracer is not initialized!")
		return
	}
	globalTracer.AddEvent(event)
}

// AddException adds an exception to the tracer
func AddException(exception *protocol.Exception) {
	if globalTracer == nil || globalTracer.closed() {
		// TODO
		log.Println("The tracer is not initialized!")
		return
	}
	globalTracer.AddException(exception)
}

// StopTracer will close the tracer and send all the data to the collector
func StopTracer() {
	if globalTracer == nil || globalTracer.closed() {
		// TODO
		log.Println("The tracer is not initialized!")
		return
	}
	select {
	case <-globalTracer.stopped:
		return
	default:
		globalTracer.closeCmd <- struct{}{}
		<-globalTracer.stopped
	}
}

func (tracer *epsagonTracer) worker() {
	defer close(tracer.stopped)
	for {
		select {
		case event := <-tracer.eventsPipe:
			tracer.events = append(tracer.events, event)
		case exception := <-tracer.exceptionsPipe:
			tracer.exceptions = append(tracer.exceptions, exception)
		case <-tracer.closeCmd:
			tracer.sendTraces()
			return
		}
	}
}
