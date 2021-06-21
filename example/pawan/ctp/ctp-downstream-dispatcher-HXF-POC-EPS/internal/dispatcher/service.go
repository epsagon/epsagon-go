package dispatcher

import (
	"sync"

	"ctp-downstream-dispatcher/internal/endpoint"
	httpx "ctp-downstream-dispatcher/pkg/http"

	"ctp-downstream-dispatcher/internal/types"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Service interface for Dispatcher
type Service interface {
	Dispatch(*types.SyncEvent) error
}

// dispatcher holds the http client to use for
// invoking sync endpoint invoke function
type dispatcher struct {
	HTTPClient *httpx.Client
}

// New returns a new instance of dispatcher
func New(httpClient *httpx.Client) Service {
	return &dispatcher{
		HTTPClient: httpClient,
	}
}

// Dispatch implements Service
// it invokes target endpoint according
// to the CTP event type (see README for details)
func (d *dispatcher) Dispatch(request *types.SyncEvent) error {
	if request == nil {
		err := errors.New("no sync event found")
		log.Error().Err(err).Msg("")
		return err
	}

	if d.HTTPClient == nil {
		err := errors.New("no http client found")
		log.Error().Err(err).Msg("")
		return err
	}

	// collect target endpoints based on event type
	endpoints, err := endpoint.Collect(request.EventType)
	if err != nil {
		err = errors.Wrap(err, "error occurred when collecting target endpoints")
		log.Error().Err(err).Msg("")
		return err
	}

	wg := sync.WaitGroup{}
	muErr := struct {
		*sync.Mutex
		*multierror.Error
	}{
		Mutex: &sync.Mutex{},
		Error: &multierror.Error{},
	}

	// trigger all lambda functions called out in arns
	for _, targetEndpoint := range endpoints {
		wg.Add(1)

		go func(ep string) {
			defer wg.Done()

			log.Info().Fields(map[string]interface{}{
				"endpoint":       ep,
				"event_type":     request.EventType,
				"VIN":            request.VIN,
				"GUID":           request.GUID,
				"wifi":           request.Wifi,
                "wifiAcceptedDate": request.WifiAcceptedDate,
                "wifiDeclinedDate": request.WifiDeclinedDate,
				"CORRELATION_ID": request.Header.XCorrelationID,
				"header":         request.Header,
				"request":        request,
			}).Msgf("invoking endpoint")

			response, err := endpoint.Invoke(d.HTTPClient, ep, request)
			if err != nil {
				err = errors.Wrapf(err, "error processing endpoint: %s with request: %+v", ep, request)

				muErr.Lock()
				// capture errors from each endpoint request
				muErr.Error = multierror.Append(muErr.Error, err)
				muErr.Unlock()

				// log error from each endpoint request if any and return
				log.Error().Err(err).Fields(map[string]interface{}{
					"endpoint":       ep,
					"event_type":     request.EventType,
					"VIN":            request.VIN,
					"GUID":           request.GUID,
					"wifi":           request.Wifi,
					"wifiAcceptedDate": request.WifiAcceptedDate,
					"wifiDeclinedDate": request.WifiDeclinedDate,
					"CORRELATION_ID": request.Header.XCorrelationID,
					"header":         request.Header,
					"request":        request,
				}).Msgf("request to endpoint failure")

				return
			}

			log.Info().Fields(map[string]interface{}{
				"endpoint":       ep,
				"event_type":     request.EventType,
				"VIN":            request.VIN,
				"GUID":           request.GUID,
				"wifi":           request.Wifi,
                "wifiAcceptedDate": request.WifiAcceptedDate,
                "wifiDeclinedDate": request.WifiDeclinedDate,
				"CORRELATION_ID": request.Header.XCorrelationID,
				"header":         request.Header,
				"request":        request,
				"response": string(response),
			}).Msgf("request to endpoint successful")
		}(targetEndpoint)
	}
	wg.Wait()
	muErr.Lock()
	wg.Wait()
	defer muErr.Unlock()

	return muErr.ErrorOrNil()
}
