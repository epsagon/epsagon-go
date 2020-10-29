package epsagon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/epsagon/epsagon-go/tracer"
)

// DefaultErrorType Default custom error type
const DefaultErrorType = "Error"

// MaxMetadataSize Maximum size of event metadata
const MaxMetadataSize = 10 * 1024

// Config is the configuration for Epsagon's tracer
type Config struct {
	tracer.Config
}

// GeneralEpsagonRecover recover function that will send exception to epsagon
// exceptionType, msg are strings that will be added to the exception
func GeneralEpsagonRecover(exceptionType, msg string, currentTracer tracer.Tracer) {
	if r := recover(); r != nil && currentTracer != nil {
		currentTracer.AddExceptionTypeAndMessage(exceptionType, fmt.Sprintf("%s:%+v", msg, r))
	}
}

// NewTracerConfig creates a new tracer Config
func NewTracerConfig(applicationName, token string) *Config {
	return &Config{
		Config: tracer.Config{
			ApplicationName: applicationName,
			Token:           token,
			MetadataOnly:    true,
			Debug:           false,
			SendTimeout:     "1s",
		},
	}
}

// Label adds a label to the sent trace
func Label(key string, value interface{}, args ...context.Context) {
	currentTracer := ExtractTracer(args)
	if currentTracer != nil {
		currentTracer.AddLabel(key, value)
	}
}

// Error adds an error to the sent trace
func Error(value interface{}, args ...context.Context) {
	currentTracer := ExtractTracer(args)
	if currentTracer != nil {
		currentTracer.AddError(DefaultErrorType, value)
	}
}

// TypeError adds an error to the sent trace with specific error type
func TypeError(value interface{}, errorType string, args ...context.Context) {
	currentTracer := ExtractTracer(args)
	if currentTracer != nil {
		currentTracer.AddError(errorType, value)
	}
}

// FormatHeaders format HTTP headers to string - using first header value, ignoring the rest
func FormatHeaders(headers http.Header) (string, error) {
	headersToFormat := make(map[string]string)
	for headerKey, headerValues := range headers {
		if len(headerValues) > 0 {
			headersToFormat[headerKey] = headerValues[0]
		}
	}
	headersJSON, err := json.Marshal(headersToFormat)
	if err != nil {
		return "", err
	}
	return string(headersJSON), nil
}

// ExtractRequestData extracts headers and body from http.Request
func ExtractRequestData(req *http.Request) (headers string, body string) {
	headers, err := FormatHeaders(req.Header)
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
	if len(buf) > MaxMetadataSize {
		trimmed = buf[0:MaxMetadataSize]
	}
	body = string(trimmed)
	return
}

// NewReadCloser returns an io.ReadCloser
// will mimick read from body depending on given error
func NewReadCloser(body []byte, err error) io.ReadCloser {
	if err != nil {
		return &errorReader{err: err}
	}
	return ioutil.NopCloser(bytes.NewReader(body))
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
