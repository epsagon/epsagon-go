package main

import (
	"github.com/pkg/errors"
	"testing"

	"ctp-downstream-dispatcher/internal/types"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

type mockDispatcherFail struct{}

func (m *mockDispatcherFail) Dispatch(request *types.SyncEvent) error {
	return errors.New("mock error")
}

type mockDispatcherSuccess struct{}

func (m *mockDispatcherSuccess) Dispatch(request *types.SyncEvent) error {
	return nil
}

func TestHandler_NoSQSEventRecord__Error(t *testing.T) {

	event := events.SQSEvent{}
	err := handler(event)
	assert.Error(t, err, "expected error")

}

func TestHandler_NoSQSRecords__Error(t *testing.T) {
	event := events.SQSEvent{Records: []events.SQSMessage{}}
	err := handler(event)
	assert.Error(t, err, "expected error")

}

func TestHandler_InvalidSQSMessageBody__Error(t *testing.T) {

	event := getSQSEvent(`[{}]`)
	err := handler(event)
	assert.Error(t, err, "expected error")

}

func TestHandler_InvalidSQSMessageBodyValidEvent__Error(t *testing.T) {
	event := getSQSEvent(`{"eventType":"SOME_EVENT", "guid":"12345678901234567890"}`)
	err := handler(event)
	assert.Error(t, err, "unexpected error")

}

func TestHandler_ValidSQSMessageBodyValidEvent_DispatcherError__Error(t *testing.T) {
	event := getSQSEvent(`{"eventType":"CUSTOMER_CREATE", "guid":"12345678901234567890"}`)

	dispatcherSvc = &mockDispatcherFail{}
	err := handler(event)
	assert.Error(t, err, "expected error")

}

func TestHandler_ValidSQSMessageBodyValidEvent__Success(t *testing.T) {
	event := getSQSEvent(`{"eventType": "Subscription_Event","guid": "123asd123zxc456dfg456dfg7a366a1b","vin": "SOMETESTVIN000","headers": {"X-CorrelationId": "52DE4B5F-4E6B-4035-9AEB-593DAA9447AF"}}`)
	dispatcherSvc = &mockDispatcherSuccess{}
	err := handler(event)
	assert.NoError(t, err, "unexpected error")

}

func TestHandler_ValidSQSMessageBodyCustomerEvent__Success(t *testing.T) {
	event := getSQSEvent(`{"eventType":"CUSTOMER_UPDATE","guid":"1234test1234567895sometestf3b6f8","headers":{"content-length":["219"],"x-brand":["T"],"postman-token":["16255980-0be0-f4b1-ce14-fbf2e02d971e"],"accept":["*/*"],"authorization":["someauthtoken1234thatnotuse123bv8WwNGGKc"],"x-correlationid":["123e4567-e89b-12d3-a456-426655440000"],"host":["localhost:8080"],"content-type":["application/json"],"x-channel":["US_SELF_AZURE"],"connection":["keep-alive"],"cache-control":["no-cache"],"accept-encoding":["gzip, deflate"],"user-agent":["PostmanRuntime/7.15.2"]}}`)
	dispatcherSvc = &mockDispatcherSuccess{}
	err := handler(event)
	assert.NoError(t, err, "unexpected error")

}

func getSQSEvent(body string) events.SQSEvent {
	return events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId: "1",
				Body:      body,
			},
		},
	}
}
