package epsagonhttp

import (
	"bytes"
	"context"
	"encoding/json"
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

const EPSAGON_TRACEID_HEADER_KEY = "epsagon-trace-id"
const EPSAGON_DOMAIN = "epsagon.com"
const APPSYNC_API_SUBDOMAIN = ".appsync-api."
const AMAZON_REQUEST_ID = "x-amzn-requestid"
const API_GATEWAY_RESOURCE_TYPE = "api_gateway"
const MAX_METADATA_SIZE = 10 * 1024

type ValidationFunction func(string, string) bool

var hasSuffix ValidationFunction = strings.HasSuffix
var contains ValidationFunction = strings.Contains

var blacklistURLs = map[*ValidationFunction][]string{
	&hasSuffix: {
		EPSAGON_DOMAIN,
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
		APPSYNC_API_SUBDOMAIN,
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

func NewTracingTransport(args ...context.Context) *TracingTransport {
	return NewWrappedTracingTransport(http.DefaultTransport, args...)
}

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
	reqHeaders, reqBody := t.extractRequestData(req, tr)
	if !isBlacklistedURL(req.URL) {
		req.Header[EPSAGON_TRACEID_HEADER_KEY] = []string{generateEpsagonTraceID()}
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
		addTraceIdToEvent(req, event)
	}
	if !t.getMetadataOnly(tr) {
		event.Resource.Metadata["request_headers"] = reqHeaders
		event.Resource.Metadata["request_body"] = reqBody
	}
}

func (t *TracingTransport) extractRequestData(req *http.Request, tr tracer.Tracer) (headers string, body string) {
	if t.getMetadataOnly(tr) {
		return
	}

	headers, err := formatHeaders(req.Header)
	if err != nil {
		headers = ""
	}

	if req.Body == nil {
		return
	}

	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return
	}
	req.Body = ioutil.NopCloser(bytes.NewReader(buf))
	// truncates request body to the first 64KB
	trimmed := buf
	if len(buf) > MAX_METADATA_SIZE {
		trimmed = buf[0:MAX_METADATA_SIZE]
	}
	body = string(trimmed)
	return
}

func isBlacklistedURL(parsedUrl *url.URL) bool {
	hostname := parsedUrl.Hostname()
	for method, urls := range whitelistURLs {
		for _, whitelistUrl := range urls {
			if (*method)(hostname, whitelistUrl) {
				return false
			}
		}
	}
	for method, urls := range blacklistURLs {
		for _, blacklistUrl := range urls {
			if (*method)(hostname, blacklistUrl) {
				return true
			}
		}
	}
	return false
}

func shouldAddHeaderByURL(rawUrl string) bool {
	parsedURL, err := url.Parse(rawUrl)
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

func addTraceIdToEvent(req *http.Request, event *protocol.Event) {
	traceIDs, ok := req.Header[EPSAGON_TRACEID_HEADER_KEY]
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
		if strings.ToLower(headerKey) == AMAZON_REQUEST_ID {
			amzRequestIDs = headerValues
			break
		}
	}
	if len(amzRequestIDs) > 0 {
		amzRequestID := amzRequestIDs[0]
		if !strings.Contains(resp.Request.URL.Hostname(), APPSYNC_API_SUBDOMAIN) {
			// api gateway
			resource.Metadata[tracer.AwsServiceKey] = API_GATEWAY_RESOURCE_TYPE
		}
		resource.Metadata[tracer.EpsagonRequestTraceIDKey] = amzRequestID
	}
}

func (c *ClientWrapper) addDataToEvent(req *http.Request, resp *http.Response, event *protocol.Event) {
	if req != nil {
		addTraceIdToEvent(req, event)
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
		req.Header[EPSAGON_TRACEID_HEADER_KEY] = []string{generateEpsagonTraceID()}
	}
	resp, err = c.Client.Do(req)
	called = true
	event := postSuperCall(startTime, req.URL.String(), req.Method, resp, err, c.getMetadataOnly())
	if req != nil {
		addTraceIdToEvent(req, event)
	}
	if !c.getMetadataOnly() {
		updateRequestData(req, event.Resource.Metadata)
	}
	c.tracer.AddEvent(event)
	return
}

// Get wraps http.Client.Get
func (c *ClientWrapper) Get(rawUrl string) (resp *http.Response, err error) {
	called := false
	defer func() {
		if !called {
			resp, err = c.Client.Get(rawUrl)
		}
	}()
	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Get", c.tracer)
	startTime := tracer.GetTimestamp()
	req, err := http.NewRequest(http.MethodGet, rawUrl, nil)
	if err != nil || !shouldAddHeaderByURL(rawUrl) {
		// err might be nil if rawUrl is invalid. Then, wrapping without any HTTP trace correlation
		resp, err = c.Client.Get(rawUrl)
	} else {
		req.Header[EPSAGON_TRACEID_HEADER_KEY] = []string{generateEpsagonTraceID()}
		resp, err = c.Client.Do(req)
	}
	called = true
	event := postSuperCall(startTime, rawUrl, http.MethodGet, resp, err, c.getMetadataOnly())
	c.addDataToEvent(req, resp, event)
	c.tracer.AddEvent(event)
	return
}

// Post wraps http.Client.Post
func (c *ClientWrapper) Post(
	rawUrl string, contentType string, body io.Reader) (resp *http.Response, err error) {
	called := false
	defer func() {
		if !called {
			resp, err = c.Client.Post(rawUrl, contentType, body)
		}
	}()
	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Post", c.tracer)
	startTime := tracer.GetTimestamp()
	req, err := http.NewRequest(http.MethodPost, rawUrl, body)
	if err != nil || !shouldAddHeaderByURL(rawUrl) {
		// err might be nil if rawUrl is invalid. Then, wrapping without any HTTP trace correlation
		resp, err = c.Client.Post(rawUrl, contentType, body)
	} else {
		req.Header.Set("Content-Type", contentType)
		req.Header[EPSAGON_TRACEID_HEADER_KEY] = []string{generateEpsagonTraceID()}
		resp, err = c.Client.Do(req)
	}
	called = true
	event := postSuperCall(startTime, rawUrl, http.MethodPost, resp, err, c.getMetadataOnly())
	c.addDataToEvent(req, resp, event)
	c.tracer.AddEvent(event)
	return
}

// PostForm wraps http.Client.PostForm
func (c *ClientWrapper) PostForm(
	rawUrl string, data url.Values) (resp *http.Response, err error) {
	called := false
	defer func() {
		if !called {
			resp, err = c.Client.PostForm(rawUrl, data)
		}
	}()
	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.PostForm", c.tracer)
	startTime := tracer.GetTimestamp()
	req, err := http.NewRequest(http.MethodPost, rawUrl, strings.NewReader(data.Encode()))
	if err != nil || !shouldAddHeaderByURL(rawUrl) {
		// err might be nil if rawUrl is invalid. Then, wrapping without any HTTP trace correlation
		resp, err = c.Client.PostForm(rawUrl, data)
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header[EPSAGON_TRACEID_HEADER_KEY] = []string{generateEpsagonTraceID()}
		resp, err = c.Client.Do(req)
	}
	called = true
	event := postSuperCall(startTime, rawUrl, http.MethodPost, resp, err, c.getMetadataOnly())
	c.addDataToEvent(req, resp, event)
	c.tracer.AddEvent(event)
	return
}

// Head wraps http.Client.Head
func (c *ClientWrapper) Head(rawUrl string) (resp *http.Response, err error) {
	called := false
	defer func() {
		if !called {
			resp, err = c.Client.Head(rawUrl)
		}
	}()
	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Head", c.tracer)
	startTime := tracer.GetTimestamp()
	req, err := http.NewRequest(http.MethodHead, rawUrl, nil)
	if err != nil || !shouldAddHeaderByURL(rawUrl) {
		// err might be nil if rawUrl is invalid. Then, wrapping without any HTTP trace correlation
		resp, err = c.Client.Head(rawUrl)
	} else {
		req.Header[EPSAGON_TRACEID_HEADER_KEY] = []string{generateEpsagonTraceID()}
		resp, err = c.Client.Do(req)
	}
	called = true
	event := postSuperCall(startTime, rawUrl, http.MethodHead, resp, err, c.getMetadataOnly())
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
	headers, err := formatHeaders(resp.Header)
	if err == nil {
		resource.Metadata["response_headers"] = headers
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err == nil {
		// truncates response body to the first 64KB
		if len(body) > MAX_METADATA_SIZE {
			resource.Metadata["response_body"] = string(body[0:MAX_METADATA_SIZE])
		} else {
			resource.Metadata["response_body"] = string(body)
		}
	}
	resp.Body = newReadCloser(body, err)
}

type errorReader struct {
	err error
}

func (er *errorReader) Read([]byte) (int, error) {
	return 0, er.err
}
func (er *errorReader) Close() error {
	return er.err
}

func newReadCloser(body []byte, err error) io.ReadCloser {
	if err != nil {
		return &errorReader{err: err}
	}
	return ioutil.NopCloser(bytes.NewReader(body))
}

func updateRequestData(req *http.Request, metadata map[string]string) {
	headers, err := formatHeaders(req.Header)
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
			if len(bodyBytes) > MAX_METADATA_SIZE {
				bodyBytes = bodyBytes[0:MAX_METADATA_SIZE]
			}
			metadata["request_body"] = string(bodyBytes)
		}
	}
}

// format HTTP headers to string - using first header value, ignoring the rest
func formatHeaders(headers http.Header) (string, error) {
	headersToFormat := make(map[string]string)
	for headerKey, headerValues := range headers {
		if len(headerValues) > 0 {
			headersToFormat[headerKey] = headerValues[0]
		}
	}
	headersJson, err := json.Marshal(headersToFormat)
	if err != nil {
		return "", err
	}
	return string(headersJson), nil
}
