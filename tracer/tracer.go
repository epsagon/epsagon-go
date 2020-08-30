package tracer

import (
	"bytes"
	"fmt"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/golang/protobuf/jsonpb"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

var (
	mutex        sync.Mutex
	GlobalTracer Tracer
	// Tracers Global Tracers, mapped by goroutine ID
	Tracers        = map[uint64]Tracer{}
	useSingleTrace = true
)

// Tracer is what a general program tracer had to provide
type Tracer interface {
	AddEvent(*protocol.Event)
	AddException(*protocol.Exception)
	AddExceptionTypeAndMessage(string, string)
	Start()
	Running() bool
	Stop()
	Stopped() bool
	GetConfig() *Config
}

// Config is the configuration for Epsagon's tracer
type Config struct {
	ApplicationName string // Application name in Epsagon
	Token           string // Epsgaon Token
	CollectorURL    string // Epsagon collector url
	MetadataOnly    bool   // Only send metadata about the event
	Debug           bool   // Print Epsagon debug information
	SendTimeout     string // Timeout for sending traces to Epsagon
	Disable         bool   // Disable sending traces
}

type epsagonTracer struct {
	Config *Config

	eventsPipe     chan *protocol.Event
	events         []*protocol.Event
	exceptionsPipe chan *protocol.Exception
	exceptions     []*protocol.Exception

	closeCmd chan struct{}
	stopped  chan struct{}
	running  chan struct{}
}

// Start starts running the tracer in another goroutine and returns
// when it is ready, or after 1 second timeout
func (tracer *epsagonTracer) Start() {
	go tracer.Run()
	timer := time.NewTimer(time.Second)
	select {
	case <-tracer.running:
		return
	case <-timer.C:
		log.Println("Epsagon Tracer couldn't start after one second timeout")
	}
}

func (tracer *epsagonTracer) sendTraces() {
	tracesReader, err := tracer.getTraceReader()
	if err != nil {
		// TODO create an exception and send a trace only with that
		log.Printf("Epsagon: Encountered an error while marshaling the traces: %v\n", err)
		return
	}
	sendTimeout, err := time.ParseDuration(tracer.Config.SendTimeout)
	if err != nil {
		if tracer.Config.Debug {
			log.Printf("Epsagon: Encountered an error while parsing send timeout: %v, using '1s'\n", err)
		}
		sendTimeout, _ = time.ParseDuration("1s")
	}

	client := &http.Client{Timeout: sendTimeout}
	if !tracer.Config.Disable {
		HandleSendTracesResponse(client.Post(tracer.Config.CollectorURL, "application/json", tracesReader))
	}
}

// HandleSendTracesResponse handles responses from the trace collector
func HandleSendTracesResponse(resp *http.Response, err error) {
	if err != nil {
		log.Printf("Error while sending traces \n%v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusInternalServerError {
		//safe to ignore the error here
		respBody, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error while sending traces \n%v", string(respBody))
	}
}

func (tracer *epsagonTracer) getTraceReader() (io.Reader, error) {
	version := "go " + runtime.Version()
	trace := protocol.Trace{
		AppName:    tracer.Config.ApplicationName,
		Token:      tracer.Config.Token,
		Events:     tracer.events,
		Exceptions: tracer.exceptions,
		Version:    VERSION,
		Platform:   version,
	}
	if tracer.Config.Debug {
		log.Printf("EPSAGON DEBUG sending trace: %+v\n", trace)
	}

	marshaler := jsonpb.Marshaler{
		EnumsAsInts: true, EmitDefaults: true, OrigName: true}
	traceJSON, err := marshaler.MarshalToString(&trace)
	if err != nil {
		return nil, err
	}
	if tracer.Config.Debug {
		log.Printf("Final Traces: %s ", traceJSON)
	}
	return bytes.NewBuffer([]byte(traceJSON)), nil
}

func isChannelPinged(ch chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

// Running return true iff the tracer has been running
func (tracer *epsagonTracer) Running() bool {
	return isChannelPinged(tracer.running)
}

// Stopped return true iff the tracer has been closed
func (tracer *epsagonTracer) Stopped() bool {
	return isChannelPinged(tracer.stopped)
}

func fillConfigDefaults(config *Config) {
	if !config.Debug {
		if strings.ToUpper(os.Getenv("EPSAGON_DEBUG")) == "TRUE" {
			config.Debug = true
		}
	}
	if len(config.Token) == 0 {
		config.Token = os.Getenv("EPSAGON_TOKEN")
		if config.Debug {
			log.Println("EPSAGON DEBUG: setting token from environment variable")
		}
	}
	if config.MetadataOnly {
		if strings.ToUpper(os.Getenv("EPSAGON_METADATA")) == "FALSE" {
			config.MetadataOnly = false
		}
	}
	if len(config.CollectorURL) == 0 {
		envURL := os.Getenv("EPSAGON_COLLECTOR_URL")
		if len(envURL) != 0 {
			config.CollectorURL = envURL
		} else {
			region := os.Getenv("AWS_REGION")
			if len(region) != 0 {
				config.CollectorURL = fmt.Sprintf("https://%s.tc.epsagon.com", region)
			} else {
				config.CollectorURL = "https://us-east-1.tc.epsagon.com"
			}
		}
		if config.Debug {
			log.Printf("EPSAGON DEBUG: setting collector url to %s\n", config.CollectorURL)
		}
	}
	sendTimeout := os.Getenv("EPSAGON_SEND_TIMEOUT_SEC")
	if len(sendTimeout) != 0 {
		config.SendTimeout = sendTimeout
		if config.Debug {
			log.Println("EPSAGON DEBUG: setting send timeout from environment variable")
		}
	}
}

// Gets current goroutine tracer info
func getCurrentTracerInfo() (tracer Tracer, currentId uint64) {
	if useSingleTrace {
		return GlobalTracer, 0
	}
	currentId = curGoroutineID()
	tracer = Tracers[currentId]
	return tracer, currentId
}

// CreateTracer will initiallize a epsagon tracer for current goroutine
func CreateTracer(config *Config) Tracer {
	mutex.Lock()
	defer mutex.Unlock()
	if strings.ToUpper(os.Getenv("EPSAGON_MULTIPLE_TRACES")) == "TRUE" {
		useSingleTrace = false
	}
	tracer, currentId := getCurrentTracerInfo()
	if tracer != nil && !tracer.Stopped() {
		log.Println("The tracer is already created, Closing and Creating.")
		tracer.Stop()
	}
	if config == nil {
		config = &Config{}
	}
	fillConfigDefaults(config)
	tracer = &epsagonTracer{
		Config:         config,
		eventsPipe:     make(chan *protocol.Event),
		events:         make([]*protocol.Event, 0, 0),
		exceptionsPipe: make(chan *protocol.Exception),
		exceptions:     make([]*protocol.Exception, 0, 0),
		closeCmd:       make(chan struct{}),
		stopped:        make(chan struct{}),
		running:        make(chan struct{}),
	}
	if !useSingleTrace {
		Tracers[currentId] = tracer
	} else {
		GlobalTracer = tracer
	}

	if config.Debug {
		log.Println("EPSAGON DEBUG: Created a new tracer")
	}
	return tracer
}

// AddException adds a tracing exception to the tracer
func (tracer *epsagonTracer) AddException(exception *protocol.Exception) {
	defer func() {
		recover()
	}()
	tracer.exceptionsPipe <- exception
}

// AddEvent adds an event to the tracer
func (tracer *epsagonTracer) AddEvent(event *protocol.Event) {
	if tracer.Config.Debug {
		log.Println("EPSAGON DEBUG: Adding event: ", event)
	}
	tracer.eventsPipe <- event
}

// AddEvent adds an event to the tracer
func AddEvent(event *protocol.Event) {
	if !useSingleTrace {
		mutex.Lock()
		defer mutex.Unlock()
	}
	defer func() {
		if r := recover(); r != nil {
			log.Println("Epsagon: Failed to add event")
		}
	}()
	tracer, _ := getCurrentTracerInfo()
	if tracer == nil || tracer.Stopped() {
		log.Println("The tracer is not initialized!")
		return
	}
	tracer.AddEvent(event)
}

// AddException adds an exception to the tracer
func AddException(exception *protocol.Exception) {
	if !useSingleTrace {
		mutex.Lock()
		defer mutex.Unlock()
	}
	defer func() {
		if r := recover(); r != nil {
			log.Println("Epsagon: Failed to add exception")
		}
	}()
	tracer, _ := getCurrentTracerInfo()
	if tracer == nil || tracer.Stopped() {
		log.Println("The tracer is not initialized!")
		return
	}
	tracer.AddException(exception)
}

// Stop stops the tracer running routine
func (tracer *epsagonTracer) Stop() {
	select {
	case <-tracer.stopped:
		return
	default:
		tracer.closeCmd <- struct{}{}
		<-tracer.stopped
	}
}

// StopTracer will close the tracer and send all the data to the collector
func StopTracer() {
	if !useSingleTrace {
		mutex.Lock()
		defer mutex.Unlock()
	}
	tracer, currentId := getCurrentTracerInfo()
	if tracer == nil || tracer.Stopped() {
		log.Println("The tracer is not initialized!")
		return
	}
	tracer.Stop()
	if !useSingleTrace {
		delete(Tracers, currentId)
	}
}

// Run starts the runner background routine that will
// run until it
func (tracer *epsagonTracer) Run() {
	if tracer.Config.Debug {
		log.Println("EPSAGON DEBUG: tracer started running")
	}
	if tracer.Running() {
		return
	}
	close(tracer.running)
	defer func() { tracer.running = make(chan struct{}) }()
	defer close(tracer.stopped)
	for {
		select {
		case event := <-tracer.eventsPipe:
			tracer.events = append(tracer.events, event)
		case exception := <-tracer.exceptionsPipe:
			tracer.exceptions = append(tracer.exceptions, exception)
		case <-tracer.closeCmd:
			if tracer.Config.Debug {
				log.Println("EPSAGON DEBUG: tracer stops running, sending traces")
			}
			tracer.sendTraces()
			return
		}
	}
}

func (tracer *epsagonTracer) GetConfig() *Config {
	return tracer.Config
}

// GetGlobalTracerConfig returns the configuration of the global tracer of the current goroutine
func GetGlobalTracerConfig() *Config {
	if !useSingleTrace {
		mutex.Lock()
		defer mutex.Unlock()
	}
	tracer, _ := getCurrentTracerInfo()
	if tracer == nil || tracer.Stopped() {
		return &Config{}
	}
	return tracer.GetConfig()
}

// AddExceptionTypeAndMessage adds an exception to the current tracer with
// the current stack and time.
// exceptionType, msg are strings that will be added to the exception
func (tracer *epsagonTracer) AddExceptionTypeAndMessage(exceptionType, msg string) {
	stack := debug.Stack()
	tracer.AddException(&protocol.Exception{
		Type:      exceptionType,
		Message:   msg,
		Traceback: string(stack),
		Time:      GetTimestamp(),
	})
}

// SwitchToMultipleTraces switches tracer status to support multiple tracers
func SwitchToMultipleTraces() {
	mutex.Lock()
	defer mutex.Unlock()
	useSingleTrace = false
}
