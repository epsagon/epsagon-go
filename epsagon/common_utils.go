package epsagon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/epsagon/epsagon-go/tracer"
	"github.com/onsi/gomega/types"
)

// DefaultErrorType Default custom error type
const DefaultErrorType = "Error"

// MaxMetadataSize Maximum size of event metadata
const MaxMetadataSize = 10 * 1024

// HTTP Content types to ignore
var ignoredContentTypes = [...]string{
	"image",
	"audio",
	"video",
	"font",
	"zip",
	"css",
}

// HTTP request file types to ignore
var ignoredFileTypes = [...]string{
	".js",
	".jsx",
	".woff",
	".woff2",
	".ttf",
	".eot",
	".ico",
}

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
	req.Body = NewReadCloser(buf, err)
	if err != nil {
		return
	}
	// truncates request body to the first 64KB
	trimmed := buf
	if len(buf) > MaxMetadataSize {
		trimmed = buf[0:MaxMetadataSize]
	}
	body = string(trimmed)
	return
}

// ShouldIgnoreRequest checks whether HTTP request should be ignored according
// to given content type and request path
func ShouldIgnoreRequest(contentType string, path string) bool {
	if len(contentType) > 0 {
		for _, ignoredContentType := range ignoredContentTypes {
			if strings.Contains(contentType, ignoredContentType) {
				return true
			}
		}
	}
	if len(path) > 0 {
		for _, ignoredFileSuffix := range ignoredFileTypes {
			if strings.HasSuffix(path, ignoredFileSuffix) {
				return true
			}
		}
	}
	return false
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

type matchUserError struct {
	exception interface{}
}

func (matcher *matchUserError) Match(actual interface{}) (bool, error) {
	uErr, ok := actual.(userError)
	if !ok {
		return false, fmt.Errorf("excpects userError, got %v", actual)
	}
	if !reflect.DeepEqual(uErr.exception, matcher.exception) {
		return false, fmt.Errorf("expected\n\t%v\nexception, got\n\t%v", matcher.exception, uErr.exception)
	}

	return true, nil
}

func (matcher *matchUserError) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto be userError with exception\n\t%#v", actual, matcher.exception)
}

func (matcher *matchUserError) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("NegatedFailureMessage")
}

// MatchUserError matches epsagon exceptions
func MatchUserError(exception interface{}) types.GomegaMatcher {
	return &matchUserError{
		exception: exception,
	}
}
