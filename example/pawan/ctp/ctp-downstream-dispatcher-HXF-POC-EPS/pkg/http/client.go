package http

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"time"
	"net/http/httputil"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
   httpEpsagon "github.com/epsagon/epsagon-go/wrappers/net/http"

)

const (
	defaultHTTPTimeout = 10 * time.Second
	//CoRelationID co-relation id for tracing
	CoRelationID = "X-CORRELATIONID"
)

var (
	// We need to consume response bodies to maintain http connections, but
	// limit the size we consume to respReadLimit.
	respReadLimit = int64(4096)
)

// Option represents the client options
type Option func(*Client)

// WithHTTPTimeout sets the http timeout
func WithHTTPTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithTLS sets the http client with custom transport
func WithTLS(tlsConfig *tls.Config) Option {
	return func(c *Client) {
		if tlsConfig != nil {
			transport := DefaultPooledTransport()
			transport.TLSClientConfig = tlsConfig
			wrappedClient := httpEpsagon.Wrap(http.Client{
				Transport: transport,
				Timeout:   c.timeout,
			})
			c.HTTPClient = &wrappedClient
		}
	}
}

// Client http client implementation
type Client struct {
	// HTTPClient *http.Client
	HTTPClient *httpEpsagon.ClientWrapper
	timeout    time.Duration
}

// NewClient returns a new instance of http Client
func NewClient(opts ...Option) *Client {
	client := &Client{
		timeout: defaultHTTPTimeout,
	}

	for _, opt := range opts {
		opt(client)
	}

	//set default client if there is no existing client
	if client.HTTPClient == nil {
		wrappedClient := httpEpsagon.Wrap(http.Client{
            Timeout:   client.timeout,
            Transport: DefaultPooledTransport(),
        })
		client.HTTPClient = &wrappedClient
	}

	return client
}

// DefaultClient returns a new http.Client with a non-shared Transport, idle connections disabled, and
// keepalives disabled.
func DefaultClient() *http.Client {
	return &http.Client{
		Transport: DefaultPooledTransport(),
	}
}

//DefaultPooledTransport returns a new http.Transport with defaults
func DefaultPooledTransport() *http.Transport {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 90 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
	return transport
}

// Get makes a HTTP GET request to provided URL
func (c *Client) Get(url string, headers http.Header) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GET request")
	}
	return c.Do(req, headers)
}

// Post makes a HTTP POST request to provided URL with the requestBody
func (c *Client) Post(url string, body io.Reader, headers http.Header) ([]byte, error) {

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create POST request")
	}

	return c.Do(req, headers)
}

// Do makes an HTTP request with the native `http.Do` interface
func (c *Client) Do(req *http.Request, headers http.Header) (body []byte, err error) {

	if headers == nil {
		headers = make(http.Header)
	}
	//adding xCorrelationId if not already passed in headers
	//X-CORRELATIONID is used instead of X-CORRELATION-ID because TSC JP accepts only the wrong header
	if headers.Get(CoRelationID) == "" {
		headers.Set(CoRelationID, uuid.NewV4().String())
	}
	//add the given headers
	req.Header = headers

	res, err := c.HTTPClient.Do(req.WithContext(req.Context()))
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}
	//seg.Close(err)
	//log request for debugging
	reqDump, _ := httputil.DumpRequestOut(req, true)
	log.Info().Msgf("request: %s", string(reqDump))
	
	defer c.drainBody(res.Body)

	if body, err = ioutil.ReadAll(res.Body); err != nil {
		err = errors.Wrap(err, "error reading http response")
		log.Error().Err(err).Msg("")
		return
	}

	if sc := res.StatusCode; sc < 200 || sc > 299 {
		//to-do proper logging later
		dat, _ := json.Marshal(json.RawMessage(body))
		err = fmt.Errorf("invalid response %v: %+v", res.Status, string(dat))
		return
	}

	return
}

// Try to read the response body so we can reuse this connection.
func (c *Client) drainBody(body io.ReadCloser) {
	defer body.Close()
	//to-do proper logging later
	io.Copy(ioutil.Discard, io.LimitReader(body, respReadLimit))
}
