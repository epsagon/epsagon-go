package epsagonhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/internal"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/google/uuid"
)

const EPSAGON_TRACEID_HEADER_KEY = "epsagon-trace-id"
const EPSAGON_TRACEID_METADATA_KEY = "http_trace_id"
const EPSAGON_DOMAIN = "epsagon.com"
const APPSYNC_API_SUBDOMAIN = ".appsync-api."
const AMAZON_REQUEST_ID = "x-amzn-requestid"
const API_GATEWAY_RESOURCE_TYPE = "api_gateway"
const EPSAGON_REQUEST_TRACEID_METADATA_KEY = "request_trace_id"
const AWS_SERVICE_KEY = "aws.service"

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

// TracingTransport is the RoundTripper implementation that traces HTTP calls
type TracingTransport struct {
	// MetadataOnly flag overriding the configuration
	MetadataOnly bool
	tracer       tracer.Tracer
}

func NewTracingTransport(args ...context.Context) *TracingTransport {
	currentTracer := internal.ExtractTracer(args)
	return &TracingTransport{
		tracer: currentTracer,
	}
}

// RoundTrip implements the RoundTripper interface to trace HTTP calls
func (t *TracingTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	called := false
	defer func() {
		if !called {
			resp, err = http.DefaultTransport.RoundTrip(req)
		}
	}()
	defer epsagon.GeneralEpsagonRecover("net.http.RoundTripper", "RoundTrip", t.tracer)
	startTime := tracer.GetTimestamp()
	if !isBlacklistedURL(req.URL) {
		req.Header[EPSAGON_TRACEID_HEADER_KEY] = []string{generateEpsagonTraceID()}
	}

	resp, err = http.DefaultTransport.RoundTrip(req)

	called = true
	event := postSuperCall(startTime, req.URL.String(), req.Method, resp, err, t.getMetadataOnly())
	t.addDataToEvent(req, resp, event)
	t.tracer.AddEvent(event)
	return

}

func (t *TracingTransport) getMetadataOnly() bool {
	return t.MetadataOnly || t.tracer.GetConfig().MetadataOnly
}

func (t *TracingTransport) addDataToEvent(req *http.Request, resp *http.Response, event *protocol.Event) {
	if req != nil {
		addTraceIdToEvent(req, event)
	}
	if resp != nil {
		if !t.getMetadataOnly() {
			updateRequestData(resp.Request, event.Resource.Metadata)
		}
	}
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
		event.Resource.Metadata[EPSAGON_TRACEID_METADATA_KEY] = traceID
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
			resource.Metadata[AWS_SERVICE_KEY] = API_GATEWAY_RESOURCE_TYPE
		}
		resource.Metadata[EPSAGON_REQUEST_TRACEID_METADATA_KEY] = amzRequestID
	}
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
	headers, err := json.Marshal(resp.Header)
	if err == nil {
		resource.Metadata["response_headers"] = string(headers)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err == nil {
		// truncates response body to the first 64KB
		if len(body) > (64 * 1024) {
			body = body[0 : 64*1024]
		}
		resource.Metadata["response_body"] = string(body)
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
	headers, err := json.Marshal(req.Header)
	if err == nil {
		metadata["request_headers"] = string(headers)
	}
	if req.Body == nil {
		return
	}
	bodyReader, err := req.GetBody()
	if err == nil {
		bodyBytes, err := ioutil.ReadAll(bodyReader)
		if err == nil {
			// truncates request body to the first 64KB
			if len(bodyBytes) > (64 * 1024) {
				bodyBytes = bodyBytes[0 : 64*1024]
			}
			metadata["request_body"] = string(bodyBytes)
		}
	}
}
