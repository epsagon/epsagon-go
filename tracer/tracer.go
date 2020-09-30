package tracer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epsagon/epsagon-go/protocol"
	"github.com/golang/protobuf/jsonpb"
)

var (
	mutex sync.Mutex
	// GlobalTracer A global Tracer for all internal uses
	GlobalTracer Tracer
)

// MaxLabelsSize is the maximum allowed total labels size
const MaxLabelsSize = 10 * 1024

// Tracer is what a general program tracer had to provide
type Tracer interface {
	AddEvent(*protocol.Event)
	AddException(*protocol.Exception)
	AddExceptionTypeAndMessage(string, string)
	AddLabel(string, interface{})
	AddError(string, interface{})
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

type EpsagonLabel struct {
	key   string
	value interface{}
}

type epsagonTracer struct {
	Config *Config

	eventsPipe          chan *protocol.Event
	events              []*protocol.Event
	runnerExceptionPipe chan *protocol.Exception
	exceptionsPipe      chan *protocol.Exception
	labelsPipe          chan EpsagonLabel
	exceptions          []*protocol.Exception
	runnerException     *protocol.Exception
	labels              map[string]interface{}
	labelsSize          int

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

// getRunnerEvent Gets the runner event, nil if not found
func (tracer *epsagonTracer) getRunnerEvent() *protocol.Event {
	for _, event := range tracer.events {
		if event.Origin == "runner" {
			return event
		}
	}
	return nil
}

func (tracer *epsagonTracer) addRunnerLabels(event *protocol.Event) {
	jsonString, err := json.Marshal(tracer.labels)
	if err != nil {
		if tracer.Config.Debug {
			log.Printf("EPSAGON DEBUG failed appending labels")
		}
	} else {
		event.Resource.Metadata["labels"] = string(jsonString)
	}
}

func (tracer *epsagonTracer) addRunnerException(event *protocol.Event) {
	if tracer.runnerException != nil {
		event.Exception = tracer.runnerException
	}
}

func (tracer *epsagonTracer) getTraceReader() (io.Reader, error) {
	version := "go " + runtime.Version()
	runnerEvent := tracer.getRunnerEvent()
	if runnerEvent != nil {
		tracer.addRunnerLabels(runnerEvent)
		tracer.addRunnerException(runnerEvent)
	}
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

// CreateTracer will initiallize a new epsagon tracer
func CreateTracer(config *Config) Tracer {
	if config == nil {
		config = &Config{}
	}
	fillConfigDefaults(config)
	tracer := &epsagonTracer{
		Config:              config,
		eventsPipe:          make(chan *protocol.Event),
		events:              make([]*protocol.Event, 0, 0),
		exceptionsPipe:      make(chan *protocol.Exception),
		runnerExceptionPipe: make(chan *protocol.Exception),
		exceptions:          make([]*protocol.Exception, 0, 0),
		closeCmd:            make(chan struct{}),
		stopped:             make(chan struct{}),
		running:             make(chan struct{}),
		labels:              make(map[string]interface{}),
		labelsPipe:          make(chan EpsagonLabel),
	}
	if config.Debug {
		log.Println("EPSAGON DEBUG: Created a new tracer")
	}
	return tracer
}

// CreateTracer will initiallize a global epsagon tracer
func CreateGlobalTracer(config *Config) Tracer {
	mutex.Lock()
	defer mutex.Unlock()
	if GlobalTracer != nil && !GlobalTracer.Stopped() {
		log.Println("The tracer is already created, Closing and Creating.")
		GlobalTracer.Stop()
	}
	GlobalTracer = CreateTracer(config)
	return GlobalTracer
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
	if GlobalTracer == nil || GlobalTracer.Stopped() {
		// TODO
		log.Println("The tracer is not initialized!")
		return
	}
	GlobalTracer.AddEvent(event)
}

func (tracer *epsagonTracer) verifyLabel(label EpsagonLabel) bool {
	var valueSize = 0
	switch label.value.(type) {
	case int, float64, bool:
		valueSize = strconv.IntSize
	case string:
		valueSize = len(label.value.(string))
	default:
		if tracer.Config.Debug {
			log.Println("EPSAGON DEBUG: Supported label types are: int, float, string, bool")
		}
		return false
	}
	if len(label.key)+valueSize+tracer.labelsSize > MaxLabelsSize {
		return false
	}

	tracer.labelsSize += len(label.key) + valueSize
	return true
}

// AddLabel adds a label to the tracer
func (tracer *epsagonTracer) AddLabel(key string, value interface{}) {
	if tracer.Config.Debug {
		log.Println("EPSAGON DEBUG: Adding label: ", key, value)
	}
	label := EpsagonLabel{key, value}
	tracer.labelsPipe <- label
}

// AddLabel adds a label to the tracer
func AddLabel(key string, value interface{}) {
	if GlobalTracer == nil || GlobalTracer.Stopped() {
		log.Println("The tracer is not initialized!")
		return
	}
	GlobalTracer.AddLabel(key, value)
}

// AddException adds an exception to the tracer
func AddException(exception *protocol.Exception) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Epsagon: Failed to add exception")
		}
	}()
	if GlobalTracer == nil || GlobalTracer.Stopped() {
		// TODO
		log.Println("The tracer is not initialized!")
		return
	}
	GlobalTracer.AddException(exception)
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
func StopGlobalTracer() {
	if GlobalTracer == nil || GlobalTracer.Stopped() {
		// TODO
		log.Println("The tracer is not initialized!")
		return
	}
	GlobalTracer.Stop()
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
		case exception := <-tracer.runnerExceptionPipe:
			tracer.runnerException = exception
		case label := <-tracer.labelsPipe:
			if tracer.verifyLabel(label) {
				tracer.labels[label.key] = label.value
			}
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

// GetGlobalTracerConfig returns the configuration of the global tracer
func GetGlobalTracerConfig() *Config {
	if GlobalTracer == nil || GlobalTracer.Stopped() {
		return &Config{}
	}
	return GlobalTracer.GetConfig()
}

func (tracer *epsagonTracer) createException(exceptionType, msg string) *protocol.Exception {
	stack := debug.Stack()
	return &protocol.Exception{
		Type:      exceptionType,
		Message:   msg,
		Traceback: string(stack),
		Time:      GetTimestamp(),
	}
}

// AddExceptionTypeAndMessage adds an exception to the current tracer with
// the current stack and time.
// exceptionType, msg are strings that will be added to the exception
func (tracer *epsagonTracer) AddExceptionTypeAndMessage(exceptionType, msg string) {
	tracer.AddException(tracer.createException(exceptionType, msg))
}

func (tracer *epsagonTracer) AddError(errorType string, value interface{}) {
	var message string
	switch value.(type) {
	case string:
		message = value.(string)
	case error:
		message = value.(error).Error()
	default:
		if tracer.Config.Debug {
			log.Println("EPSAGON DEBUG: Supported error types are: string, error")
		}
		return
	}
	if tracer.Config.Debug {
		log.Println("EPSAGON DEBUG: Adding error message to trace: ", message)
	}
	exception := tracer.createException(errorType, message)
	tracer.runnerExceptionPipe <- exception
}
