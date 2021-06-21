package inner

import (
	"ctp-downstream-dispatcher/internal/dispatcher"
	"ctp-downstream-dispatcher/internal/types"
	httpx "ctp-downstream-dispatcher/pkg/http"
	"encoding/json"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

    "github.com/epsagon/epsagon-go/epsagon"
)

const (
	defaultHTTPTimeout = 10 * time.Second
)

var (
	dispatcherSvc dispatcher.Service
)

func init() {
	// initialise dispatcher
	dispatcherSvc = dispatcher.New(
		httpx.NewClient(
			httpx.WithHTTPTimeout(defaultHTTPTimeout),
		),
	)
}

// handler is the entry point for the dispatcher lambda function
func handler(event events.SQSEvent) error { // check if lambda event payload is empty
	if len(event.Records) == 0 {
		err := errors.New("no SQS message found")
		log.Error().Err(err).Msg("")
		return err
	}

	if err := xray.Configure(xray.Config{LogLevel: "trace"}); err != nil {
		err = errors.Wrap(err, "unable to configure xray tracing")
		log.Error().Err(err).Msg("")
	}

	// loop through SQSMessage to pass message Body to target lambda function
	for _, msg := range event.Records {
		request := &types.SyncEvent{}

		if err := json.Unmarshal([]byte(msg.Body), request); err != nil {
			err = errors.Wrapf(err, "error occurred when unmarshalling the SQS message (%v)", msg.Body)
			log.Error().Err(err).Msg("")
			return err
		}

		// call dispatcher
		if err := dispatcherSvc.Dispatch(request); err != nil {
			err = errors.Wrapf(err, "error invoking dispatch method")
			log.Error().Err(err).Msg("")
			return err
		}
	}
	return nil
}
func main() {
	config := epsagon.NewTracerConfig("ctp-downstream-dispatcher-stgwest2", "9972dfef-8598-4dc8-8cea-5f9337762a9a")
	config.Debug = true
	lambda.Start(epsagon.WrapLambdaHandler(config, handler))

}
