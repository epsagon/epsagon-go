package epsagonhttp

import (
	"encoding/json"
	// "fmt"
	"bytes"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// ClientWrapper is Epsagon's wrapper for http.Client
type ClientWrapper struct {
	http.Client
}

// Wrap wraps an http.Client to Epsagon's ClientWrapper
func Wrap(c http.Client) ClientWrapper {
	return ClientWrapper{c}
}

// Do wraps http.Client's Do
func (c *ClientWrapper) Do(req *http.Request) (resp *http.Response, err error) {
	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Do")

	startTime := epsagon.GetTimestamp()
	resp, err = c.Client.Do(req)
	event := postSuperCall(startTime, req.URL.String(), req.Method, resp, err)
	if !epsagon.GetGlobalTracerConfig().MetadataOnly {
		updateRequestData(req, event.Resource.Metadata)
	}
	epsagon.AddEvent(event)
	return
}

// Get wraps http.Client.Get
func (c *ClientWrapper) Get(url string) (resp *http.Response, err error) {
	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Do")
	startTime := epsagon.GetTimestamp()
	resp, err = c.Client.Get(url)
	event := postSuperCall(startTime, url, http.MethodPost, resp, err)
	if resp != nil && !epsagon.GetGlobalTracerConfig().MetadataOnly {
		updateRequestData(resp.Request, event.Resource.Metadata)
	}
	epsagon.AddEvent(event)
	return
}

// Post wraps http.Client.Post
func (c *ClientWrapper) Post(
	url string, contentType string, body io.Reader) (resp *http.Response, err error) {

	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Do")
	startTime := epsagon.GetTimestamp()
	resp, err = c.Client.Post(url, contentType, body)
	event := postSuperCall(startTime, url, http.MethodPost, resp, err)
	if resp != nil && !epsagon.GetGlobalTracerConfig().MetadataOnly {
		updateRequestData(resp.Request, event.Resource.Metadata)
	}
	epsagon.AddEvent(event)
	return
}

// PostForm wraps http.Client.PostForm
func (c *ClientWrapper) PostForm(
	url string, data url.Values) (resp *http.Response, err error) {

	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Do")
	startTime := epsagon.GetTimestamp()
	resp, err = c.Client.PostForm(url, data)
	event := postSuperCall(startTime, url, http.MethodPost, resp, err)
	if resp != nil && !epsagon.GetGlobalTracerConfig().MetadataOnly {
		updateRequestData(resp.Request, event.Resource.Metadata)
		dataBytes, err := json.Marshal(data)
		if err == nil {
			event.Resource.Metadata["body"] = string(dataBytes)
		}
	}
	epsagon.AddEvent(event)
	return
}

// Head wraps http.Client.Head
func (c *ClientWrapper) Head(url string) (resp *http.Response, err error) {

	defer epsagon.GeneralEpsagonRecover("net.http.Client", "Client.Do")
	startTime := epsagon.GetTimestamp()
	resp, err = c.Client.Head(url)
	event := postSuperCall(startTime, url, http.MethodPost, resp, err)
	if resp != nil && !epsagon.GetGlobalTracerConfig().MetadataOnly {
		updateRequestData(resp.Request, event.Resource.Metadata)
	}
	epsagon.AddEvent(event)
	return
}

func postSuperCall(
	startTime float64,
	url string,
	method string,
	resp *http.Response,
	err error) *protocol.Event {

	endTime := epsagon.GetTimestamp()
	duration := endTime - startTime

	event := createHTTPEvent(url, method, err)
	event.StartTime = startTime
	event.Duration = duration
	if resp != nil {
		updateResponseData(resp, event.Resource)
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

func updateResponseData(resp *http.Response, resource *protocol.Resource) {
	resource.Metadata["error_code"] = strconv.Itoa(resp.StatusCode)
	if _, ok := resp.Header["x-amzn-requestid"]; ok {
		resource.Type = "api_gateway"
		resource.Name = resp.Request.URL.Path
		resource.Metadata["request_trace_id"] = resp.Header["x-amzn-requestid"][0]
	}
	if epsagon.GetGlobalTracerConfig().MetadataOnly {
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
