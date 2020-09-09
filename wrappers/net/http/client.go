package epsagonhttp

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/internal"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const EPSAGON_TRACE_ID_KEY = "epsagon-trace-id"

type ValidationFunction func(string, string) bool

var hasSuffix ValidationFunction = strings.HasSuffix
var contains ValidationFunction = strings.Contains

var blacklistURLs = map[*ValidationFunction][]string{
	&hasSuffix: {
		"epsagon.com",
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
		".appsync-api.",
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
	currentTracer := internal.ExtractTracer(args)
	return ClientWrapper{c, false, currentTracer}
}

func (c *ClientWrapper) getMetadataOnly() bool {
	return c.MetadataOnly || c.tracer.GetConfig().MetadataOnly
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
	return isBlacklistedURL(parsedURL)
}

func generateEpsagonTraceID() string {
	traceID := hex.EncodeToString(uuid.NewRandom().Bytes())
	spanID := hex.EncodeToString(uuid.NewRandom().Bytes()[:8])
	parentSpanID := hex.EncodeToString(uuid.NewRandom().Bytes()[:8])
	return fmt.Printf("%s:%s:%s:1", traceID, spanID, parentSpanID)
}

// Do wraps http.Client's Do
func (c *ClientWrapper) Do(req *http.Request) (resp *http.Response, err error) {
	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Do", c.tracer)

	startTime := tracer.GetTimestamp()
	epsagonTraceID := ""
	if !isBlacklistedURL(req) {
		epsagonTraceID = generateEpsagonTraceID()
		req.Header.Set(EPSAGON_TRACE_ID_KEY, epsagonTraceID)
	}
	resp, err = c.Client.Do(req)
	event := postSuperCall(startTime, req.URL.String(), req.Method, resp, err, c.getMetadataOnly())
	if !c.getMetadataOnly() {
		updateRequestData(req, event.Resource.Metadata)
	}
	c.tracer.AddEvent(event)
	return
}

// Get wraps http.Client.Get
func (c *ClientWrapper) Get(url string) (resp *http.Response, err error) {
	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Do", c.tracer)
	startTime := tracer.GetTimestamp()
	resp, err = c.Client.Get(url)
	event := postSuperCall(startTime, url, http.MethodGet, resp, err, c.getMetadataOnly())
	if resp != nil && !c.getMetadataOnly() {
		updateRequestData(resp.Request, event.Resource.Metadata)
	}
	c.tracer.AddEvent(event)
	return
}

// Post wraps http.Client.Post
func (c *ClientWrapper) Post(
	url string, contentType string, body io.Reader) (resp *http.Response, err error) {

	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Do", c.tracer)
	startTime := tracer.GetTimestamp()
	resp, err = c.Client.Post(url, contentType, body)
	event := postSuperCall(startTime, url, http.MethodPost, resp, err, c.getMetadataOnly())
	if resp != nil && !c.getMetadataOnly() {
		updateRequestData(resp.Request, event.Resource.Metadata)
	}
	c.tracer.AddEvent(event)
	return
}

// PostForm wraps http.Client.PostForm
func (c *ClientWrapper) PostForm(
	url string, data url.Values) (resp *http.Response, err error) {

	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Do", c.tracer)
	startTime := tracer.GetTimestamp()
	resp, err = c.Client.PostForm(url, data)
	event := postSuperCall(startTime, url, http.MethodPost, resp, err, c.getMetadataOnly())
	if resp != nil && !c.getMetadataOnly() {
		updateRequestData(resp.Request, event.Resource.Metadata)
		dataBytes, err := json.Marshal(data)
		if err == nil {
			event.Resource.Metadata["body"] = string(dataBytes)
		}
	}
	c.tracer.AddEvent(event)
	return
}

// Head wraps http.Client.Head
func (c *ClientWrapper) Head(url string) (resp *http.Response, err error) {

	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Do", c.tracer)
	startTime := tracer.GetTimestamp()
	resp, err = c.Client.Head(url)
	event := postSuperCall(startTime, url, http.MethodHead, resp, err, c.getMetadataOnly())
	if resp != nil && !c.getMetadataOnly() {
		updateRequestData(resp.Request, event.Resource.Metadata)
	}
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
	resource.Metadata["error_code"] = strconv.Itoa(resp.StatusCode)
	if _, ok := resp.Header["x-amzn-requestid"]; ok {
		resource.Type = "api_gateway"
		resource.Name = resp.Request.URL.Path
		resource.Metadata["request_trace_id"] = resp.Header["x-amzn-requestid"][0]
	}
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
	traceID := req.Header.Get(EPSAGON_TRACE_ID_KEY)
	if len(traceID) > 0 {
		metadata["http_trace_id"] = traceID
	}
	if req.Body == nil {
		return
	}
	bodyReader, err := req.GetBody()
	if err == nil {
		bodyBytes, err := ioutil.ReadAll(bodyReader)
		if err == nil {
			metadata["request_body"] = string(bodyBytes)
		}
	}
}
