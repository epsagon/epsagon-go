package endpoint

import (
	"ctp-downstream-dispatcher/internal/types"
	httpx "ctp-downstream-dispatcher/pkg/http"
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
)

var (
	apiKey = os.Getenv("SUB_INT_API_KEY")
)

// Invoke sends the message from SQS as payload to the target API
func Invoke(httpClient *httpx.Client, targetEndpoint string, payload *types.SyncEvent) ([]byte, error) {
	if targetEndpoint == "" {
		err := errors.New("missing target endpoint")
		log.Error().Err(err).Msg("")
		return nil, err
	}
	log.Info().Msg("setting http request input values")

	if payload == nil {
		err := errors.New("missing payload")
		log.Error().Err(err).Msg("")
		return nil, err
	}

	// ignoring marshalling error as marshalling struct is always successful
	body, err := json.Marshal(payload)
	if err != nil {
		err := errors.Wrapf(err, "fail to encode request payload")
		log.Error().Err(err).Msg("fail to encode request payload")
		return nil, err
	}

	log.Info().Msgf("http request input: (%s)", payload)
	log.Info().Msgf("invoking endpoint : (%s)", targetEndpoint)

	headers := getHeaders(payload) // TODO

	// send downstream call
	response, err := httpClient.Post(targetEndpoint, bytes.NewBuffer(body), headers)
	if err != nil {
		err = errors.Wrap(err, "failed to invoke sync endpoint")
		log.Error().Err(err).Msg("")
		return nil, err
	}
	log.Info().Msg("sync endpoint successfully invoked")

	return response, nil
}

// TODO
func getHeaders(request *types.SyncEvent) http.Header {
	headers := http.Header{}
	headers.Set("Accept", "*/*")
	headers.Set("Authorization", apiKey)
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Connection", "keep-alive")
	headers.Set("Content-Type", "application/json")
	headers.Set("Host", "localhost:8084")
	headers.Set("X-CHANNEL", "US_DEALER")
	headers.Set("accept-encoding", "gzip, deflate")
	headers.Set("cache-control", "no-cache")
	headers.Set("content-length", "50")
	headers.Set("guid", request.GUID)
	headers.Set("vin", request.VIN)
	headers.Set("event_type", request.EventType)
	headers.Set("X-CORRELATIONID", request.Header.XCorrelationID)
	headers.Set("wifi", request.Wifi)
	headers.Set("wifiAcceptedDate", request.WifiAcceptedDate)
	headers.Set("wifiDeclinedDate", request.WifiDeclinedDate)
	return headers
}
