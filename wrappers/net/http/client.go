package epsagonhttp

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/google/uuid"
)

// EpsagonTraceidHeaderKey header key for epsagon trace id
const EpsagonTraceidHeaderKey = "epsagon-trace-id"

// EpsagonDomain epsagon domain
const EpsagonDomain = "epsagon.com"

// AppsyncAPISubdomain appsync domain
const AppsyncAPISubdomain = ".appsync-api."

// AmazonRequestID amazon request id header
const AmazonRequestID = "x-amzn-requestid"

// APIGatewayResourceType api gateway resource type
const APIGatewayResourceType = "api_gateway"

// ValidationFunction is a type for a validation function
type ValidationFunction func(string, string) bool

var hasSuffix ValidationFunction = strings.HasSuffix
var contains ValidationFunction = strings.Contains

var blacklistURLs = map[*ValidationFunction][]string{
	&hasSuffix: {
		EpsagonDomain,
		".amazonaws.com",
	},
	&contains: {
		"accounts.google.com",
		"documents.azure.com",
		"169.254.170.2", // AWS Task Metadata Endpoint
	},
}
var whitelistURLs = map[*ValidationFunction][]string{
	&contains: {
		".execute-api.",
		".elb.amazonaws.com",
		AppsyncAPISubdomain,
	},
}

// ClientWrapper is Epsagon's wrapper for http.Client
type ClientWrapper struct {
	http.Client

	// MetadataOnly flag overriding the configuration
	MetadataOnly bool
	tracer       tracer.Tracer
}

// Wrap wraps an http.Client to Epsagon's ClientWrapper
func Wrap(c http.Client, args ...context.Context) ClientWrapper {
	currentTracer := epsagon.ExtractTracer(args)
	return ClientWrapper{c, false, currentTracer}
}

func (c *ClientWrapper) getMetadataOnly() bool {
	return c.MetadataOnly || c.tracer.GetConfig().MetadataOnly
}

// TracingTransport is the RoundTripper implementation that traces HTTP calls
type TracingTransport struct {
	// MetadataOnly flag overriding the configuration
	MetadataOnly bool
	tracer       tracer.Tracer
	transport    http.RoundTripper
}

// NewTracingTransport creates a new tracing transport
func NewTracingTransport(args ...context.Context) *TracingTransport {
	return NewWrappedTracingTransport(http.DefaultTransport, args...)
}

// NewWrappedTracingTransport creates a new wrapped tracing transport
func NewWrappedTracingTransport(rt http.RoundTripper, args ...context.Context) *TracingTransport {
	currentTracer := epsagon.ExtractTracer(args)
	return &TracingTransport{
		tracer:    currentTracer,
		transport: rt,
	}
}

// RoundTrip implements the RoundTripper interface to trace HTTP calls
func (t *TracingTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	// reference to the tracer
	tr := t.tracer
	// if the TracingTransport is created before the global tracer is created it will be nil
	if tr == nil {
		tr = epsagon.ExtractTracer(nil)
		if tr != nil && tr.GetConfig().Debug {
			log.Println("EPSAGON DEBUG: defaulting to global tracer in RoundTrip")
		}
	}

	called := false
	defer func() {
		if !called {
			resp, err = t.transport.RoundTrip(req)
		}
	}()
	defer epsagon.GeneralEpsagonRecover("net.http.RoundTripper", "RoundTrip", t.tracer)
	startTime := tracer.GetTimestamp()
	reqHeaders, reqBody := "", ""
	if !t.getMetadataOnly(tr) {
		reqHeaders, reqBody = epsagon.ExtractRequestData(req)
	}
	if !isBlacklistedURL(req.URL) {
		req.Header[EpsagonTraceidHeaderKey] = []string{generateEpsagonTraceID()}
	}

	resp, err = t.transport.RoundTrip(req)

	called = true
	event := postSuperCall(startTime, req.URL.String(), req.Method, resp, err, t.getMetadataOnly(tr))
	t.addDataToEvent(reqHeaders, reqBody, req, event, tr)
	tr.AddEvent(event)
	return

}

func (t *TracingTransport) getMetadataOnly(tr tracer.Tracer) bool {
	return t.MetadataOnly || tr.GetConfig().MetadataOnly
}

func (t *TracingTransport) addDataToEvent(reqHeaders, reqBody string, req *http.Request, event *protocol.Event, tr tracer.Tracer) {
	if req != nil {
		addTraceIDToEvent(req, event)
	}
	if !t.getMetadataOnly(tr) {
		event.Resource.Metadata["request_headers"] = reqHeaders
		event.Resource.Metadata["request_body"] = reqBody
	}
}

func isBlacklistedURL(parsedURL *url.URL) bool {
	hostname := parsedURL.Hostname()
	for method, urls := range whitelistURLs {
		for _, whitelistURL := range urls {
			if (*method)(hostname, whitelistURL) {
				return false
			}
		}
	}
	for method, urls := range blacklistURLs {
		for _, blacklistURL := range urls {
			if (*method)(hostname, blacklistURL) {
				return true
			}
		}
	}
	return false
}

func shouldAddHeaderByURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return !isBlacklistedURL(parsedURL)
}

func generateRandomUUID() string {
	uuid, err := uuid.NewRandom()
	if err != nil {
		panic("failed to generate random UUID")
	}
	return strings.ReplaceAll(uuid.String(), "-", "")
}

func generateEpsagonTraceID() string {
	traceID := generateRandomUUID()
	spanID := generateRandomUUID()[:16]
	parentSpanID := generateRandomUUID()[:16]
	return fmt.Sprintf("%s:%s:%s:1", traceID, spanID, parentSpanID)
}

func addTraceIDToEvent(req *http.Request, event *protocol.Event) {
	traceIDs, ok := req.Header[EpsagonTraceidHeaderKey]
	if ok && len(traceIDs) > 0 {
		traceID := traceIDs[0]
		event.Resource.Metadata[tracer.EpsagonHTTPTraceIDKey] = traceID
	}
}

// update event data according to given response headers
// adds amazon request ID, if returned in response headers
// used for traces HTTP correlation (appsync / api gateway targets)
func updateByResponseHeaders(resp *http.Response, resource *protocol.Resource) {
	var amzRequestIDs []string
	for headerKey, headerValues := range resp.Header {
		if strings.ToLower(headerKey) == AmazonRequestID {
			amzRequestIDs = headerValues
			break
		}
	}
	if len(amzRequestIDs) > 0 {
		amzRequestID := amzRequestIDs[0]
		if !strings.Contains(resp.Request.URL.Hostname(), AppsyncAPISubdomain) {
			// api gateway
			resource.Metadata[tracer.AwsServiceKey] = APIGatewayResourceType
		}
		resource.Metadata[tracer.EpsagonRequestTraceIDKey] = amzRequestID
	}
}

func (c *ClientWrapper) addDataToEvent(req *http.Request, resp *http.Response, event *protocol.Event) {
	if req != nil {
		addTraceIDToEvent(req, event)
	}
	if resp != nil {
		if !c.getMetadataOnly() {
			updateRequestData(resp.Request, event.Resource.Metadata)
		}
	}
}

// Do wraps http.Client's Do
func (c *ClientWrapper) Do(req *http.Request) (resp *http.Response, err error) {
	called := false
	defer func() {
		if !called {
			resp, err = c.Client.Do(req)
		}
	}()
	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Do", c.tracer)
	startTime := tracer.GetTimestamp()
	if !isBlacklistedURL(req.URL) {
		req.Header[EpsagonTraceidHeaderKey] = []string{generateEpsagonTraceID()}
	}
	resp, err = c.Client.Do(req)
	called = true
	event := postSuperCall(startTime, req.URL.String(), req.Method, resp, err, c.getMetadataOnly())
	if req != nil {
		addTraceIDToEvent(req, event)
	}
	if !c.getMetadataOnly() {
		updateRequestData(req, event.Resource.Metadata)
	}
	c.tracer.AddEvent(event)
	return
}

// Get wraps http.Client.Get
func (c *ClientWrapper) Get(rawURL string) (resp *http.Response, err error) {
	called := false
	defer func() {
		if !called {
			resp, err = c.Client.Get(rawURL)
		}
	}()
	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Get", c.tracer)
	startTime := tracer.GetTimestamp()
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil || !shouldAddHeaderByURL(rawURL) {
		// err might be nil if rawURL is invalid. Then, wrapping without any HTTP trace correlation
		resp, err = c.Client.Get(rawURL)
	} else {
		req.Header[EpsagonTraceidHeaderKey] = []string{generateEpsagonTraceID()}
		resp, err = c.Client.Do(req)
	}
	called = true
	event := postSuperCall(startTime, rawURL, http.MethodGet, resp, err, c.getMetadataOnly())
	c.addDataToEvent(req, resp, event)
	c.tracer.AddEvent(event)
	return
}

// Post wraps http.Client.Post
func (c *ClientWrapper) Post(
	rawURL string, contentType string, body io.Reader) (resp *http.Response, err error) {
	called := false
	defer func() {
		if !called {
			resp, err = c.Client.Post(rawURL, contentType, body)
		}
	}()
	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Post", c.tracer)
	startTime := tracer.GetTimestamp()
	req, err := http.NewRequest(http.MethodPost, rawURL, body)
	if err != nil || !shouldAddHeaderByURL(rawURL) {
		// err might be nil if rawURL is invalid. Then, wrapping without any HTTP trace correlation
		resp, err = c.Client.Post(rawURL, contentType, body)
	} else {
		req.Header.Set("Content-Type", contentType)
		req.Header[EpsagonTraceidHeaderKey] = []string{generateEpsagonTraceID()}
		resp, err = c.Client.Do(req)
	}
	called = true
	event := postSuperCall(startTime, rawURL, http.MethodPost, resp, err, c.getMetadataOnly())
	c.addDataToEvent(req, resp, event)
	c.tracer.AddEvent(event)
	return
}

// PostForm wraps http.Client.PostForm
func (c *ClientWrapper) PostForm(
	rawURL string, data url.Values) (resp *http.Response, err error) {
	called := false
	defer func() {
		if !called {
			resp, err = c.Client.PostForm(rawURL, data)
		}
	}()
	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.PostForm", c.tracer)
	startTime := tracer.GetTimestamp()
	req, err := http.NewRequest(http.MethodPost, rawURL, strings.NewReader(data.Encode()))
	if err != nil || !shouldAddHeaderByURL(rawURL) {
		// err might be nil if rawURL is invalid. Then, wrapping without any HTTP trace correlation
		resp, err = c.Client.PostForm(rawURL, data)
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header[EpsagonTraceidHeaderKey] = []string{generateEpsagonTraceID()}
		resp, err = c.Client.Do(req)
	}
	called = true
	event := postSuperCall(startTime, rawURL, http.MethodPost, resp, err, c.getMetadataOnly())
	c.addDataToEvent(req, resp, event)
	c.tracer.AddEvent(event)
	return
}

// Head wraps http.Client.Head
func (c *ClientWrapper) Head(rawURL string) (resp *http.Response, err error) {
	called := false
	defer func() {
		if !called {
			resp, err = c.Client.Head(rawURL)
		}
	}()
	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Head", c.tracer)
	startTime := tracer.GetTimestamp()
	req, err := http.NewRequest(http.MethodHead, rawURL, nil)
	if err != nil || !shouldAddHeaderByURL(rawURL) {
		// err might be nil if rawURL is invalid. Then, wrapping without any HTTP trace correlation
		resp, err = c.Client.Head(rawURL)
	} else {
		req.Header[EpsagonTraceidHeaderKey] = []string{generateEpsagonTraceID()}
		resp, err = c.Client.Do(req)
	}
	called = true
	event := postSuperCall(startTime, rawURL, http.MethodHead, resp, err, c.getMetadataOnly())
	c.addDataToEvent(req, resp, event)
	c.tracer.AddEvent(event)
	return
}

func postSuperCall(
	startTime float64,
	url string,
	method string,
	resp *http.Response,
	err error,
	metadataOnly bool) *protocol.Event {

	endTime := tracer.GetTimestamp()
	duration := endTime - startTime

	event := createHTTPEvent(url, method, err)
	event.StartTime = startTime
	event.Duration = duration
	if resp != nil {
		updateResponseData(resp, event.Resource, metadataOnly)
	}
	return event
}

func createHTTPEvent(url, method string, err error) *protocol.Event {
	errorcode := protocol.ErrorCode_OK
	if err != nil {
		errorcode = protocol.ErrorCode_ERROR
	}
	return &protocol.Event{
		Id:        "http.Client-" + uuid.New().String(),
		Origin:    "http.Client",
		ErrorCode: errorcode,
		Resource: &protocol.Resource{
			Name:      url,
			Type:      "http",
			Operation: method,
			Metadata:  map[string]string{},
		},
	}
}

func updateResponseData(resp *http.Response, resource *protocol.Resource, metadataOnly bool) {
	resource.Metadata["status_code"] = strconv.Itoa(resp.StatusCode)
	updateByResponseHeaders(resp, resource)
	if metadataOnly {
		return
	}
	headers, err := epsagon.FormatHeaders(resp.Header)
	if err == nil {
		resource.Metadata["response_headers"] = headers
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err == nil {
		// truncates response body to the first 64KB
		if len(body) > epsagon.MaxMetadataSize {
			resource.Metadata["response_body"] = string(body[0:epsagon.MaxMetadataSize])
		} else {
			resource.Metadata["response_body"] = string(body)
		}
	}
	resp.Body = epsagon.NewReadCloser(body, err)
}

func updateRequestData(req *http.Request, metadata map[string]string) {
	headers, err := epsagon.FormatHeaders(req.Header)
	if err == nil {
		metadata["request_headers"] = headers
	}
	if req.Body == nil {
		return
	}
	bodyReader, err := req.GetBody()
	if err == nil {
		bodyBytes, err := ioutil.ReadAll(bodyReader)
		if err == nil {
			// truncates request body to the first 64KB
			if len(bodyBytes) > epsagon.MaxMetadataSize {
				bodyBytes = bodyBytes[0:epsagon.MaxMetadataSize]
			}
			metadata["request_body"] = string(bodyBytes)
		}
	}
}
