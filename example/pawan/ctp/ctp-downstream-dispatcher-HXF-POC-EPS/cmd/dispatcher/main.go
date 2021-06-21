package main

import (
	"ctp-downstream-dispatcher/internal/dispatcher"
	"ctp-downstream-dispatcher/internal/types"
	httpx "ctp-downstream-dispatcher/pkg/http"
	//"encoding/json"
	"fmt"
	"time"

	//"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	//"github.com/aws/aws-xray-sdk-go/xray"
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

}

// handler is the entry point for the dispatcher lambda function
func handler() error { // check if lambda event payload is empty

	dispatcherSvc = dispatcher.New(
		httpx.NewClient(
			httpx.WithHTTPTimeout(defaultHTTPTimeout),
		),
	)

	//if len(event.Records) == 0 {
	//	err := errors.New("no SQS message found")
	//	log.Error().Err(err).Msg("")
	//	return err
	//}

	//if err := xray.Configure(xray.Config{LogLevel: "trace"}); err != nil {
	//	err = errors.Wrap(err, "unable to configure xray tracing")
	//	log.Error().Err(err).Msg("")
	//}

	// loop through SQSMessage to pass message Body to target lambda function
	for i := 0; i < 3; i++ {
		fmt.Println(i)
		request := &types.SyncEvent{
			EventType: "CUSTOMER_CREATE",
			GUID: "test-guid",
			VIN: "test-vin",
			Wifi: "test-wifi",
			WifiAcceptedDate: "test-accepted",
			WifiDeclinedDate: "test-declined",
			Header: types.Header{},
		}

		//fmt.Println("MSG BODY")
		//fmt.Println(msg.Body)
		//if err := json.Unmarshal([]byte(msg.Body), request); err != nil {
		//	err = errors.Wrapf(err, "error occurred when unmarshalling the SQS message (%v)", msg.Body)
		//	log.Error().Err(err).Msg("")
		//	return err
		//}

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

	//lambda.Start(handler)
}
