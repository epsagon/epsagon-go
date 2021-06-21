package dispatcher

import (
	"testing"

	"ctp-downstream-dispatcher/internal/endpoint"
	"ctp-downstream-dispatcher/internal/types"
	httpx "ctp-downstream-dispatcher/pkg/http"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

//TODO
type mockPostAllSuccess struct{}

//TODO
func (m *mockPostAllSuccess) Invoke(httpClient *httpx.Client, targetEndpoint string, payload *types.SyncEvent) ([]byte, error) {
	return nil, nil
}

//TODO
type mockOnePostSuccessRestFails struct{}

func (m *mockOnePostSuccessRestFails) Invoke(httpClient *httpx.Client, targetEndpoint string, payload *types.SyncEvent) ([]byte, error) {
	var err error
	if targetEndpoint == "SubscriptionIntegrationSyncURI" {
		err = errors.New("mock error")
	} else {
		err = nil
	}

	return nil, err
}

type mockTwoPostSuccessAndRestFails struct{}

func (m *mockTwoPostSuccessAndRestFails) Invoke(httpClient *httpx.Client, targetEndpoint string, payload *types.SyncEvent) ([]byte, error) {
	var err error
	targetEndpoint1 := "SubscriptionIntegrationSyncURI"
	targetEndpoint2 := "SubscriptionIntegrationLangURI"
	if targetEndpoint == targetEndpoint1 || targetEndpoint == targetEndpoint2 {
		err = errors.New("mock error")
	} else {
		err = nil
	}

	return nil, err
}

func TestDispatcher_NoSyncEvent__Error(t *testing.T) {
	dispatcher := dispatcher{}
	err := dispatcher.Dispatch(nil)
	assert.Error(t, err, "expected error")

}

func TestDispatcher_InvalidEventType__Error(t *testing.T) {
	request := getSyncEvent("12345678901234567890", "CUSTOMER_CREATE_something")

	//TODO
	//svc := &mockOnePostSuccessRestFails{}
	dispatcher := dispatcher{}

	err := dispatcher.Dispatch(&request)
	assert.Error(t, err, "expected error")

}

func TestDispatcher_MissingLambdaClient__Error(t *testing.T) {

	request := getSyncEvent("12345678901234567890", "CUSTOMER_UPDATE")
	dispatcher := dispatcher{}

	err := dispatcher.Dispatch(&request)
	assert.Error(t, err, "expected no error")

}

func TestDispatcher_LambdaFailure__Error(t *testing.T) {

	request := getSyncEvent("12345678901234567890", "CUSTOMER_UPDATE")
	setEnvironmentVariables()
	//TODO
	//p, _ := json.Marshal(request)
	//svc := &mockOnePostSuccessRestFails{
	//	payload: p,
	//}
	dispatcher := dispatcher{
		//TODO
		//HTTPClient: svc,
	}

	err := dispatcher.Dispatch(&request)
	assert.Error(t, err, "expected no error")

}

func TestDispatcher_TwoLambdaSuccessAndRestFail__Error(t *testing.T) {

	request := getSyncEvent("12345678901234567890", "CUSTOMER_UPDATE")
	setEnvironmentVariables()
	//TODO
	//p, _ := json.Marshal(request)
	//svc := &mockTwoPostSuccessAndRestFails{
	//	payload: p,
	//}
	dispatcher := dispatcher{
		//TODO
		//HTTPClient: svc,
	}

	err := dispatcher.Dispatch(&request)
	assert.Error(t, err, "expected no error")

}

//func TestDispatcher_Successfully_triggered_lambdas(t *testing.T) {
//
//	request := getSyncEvent("12345678901234567890", "CUSTOMER_CREATE")
//
//	setEnvironmentVariables()
//	//TODO
//	//svc := &mockPostAllSuccess{}
//	dispatcher := dispatcher{
//		//TODO
//		//HTTPClient: svc,
//	}
//
//	err := dispatcher.Dispatch(&request)
//	assert.NoError(t, err, "expected no error")
//
//}

func setEnvironmentVariables() {
	endpoint.SubscriptionIntegrationSyncURI = "SubscriptionIntegrationSyncURI"
	endpoint.SubscriptionIntegrationLangURI = "SubscriptionIntegrationLangURI"
	endpoint.CustomerCreateEvent = "CUSTOMER_CREATE"
	endpoint.CustomerUpdateEvent = "CUSTOMER_UPDATE"
	endpoint.SubscriptionEvent = "SUBSCRIPTION_EVENT"
	endpoint.CustomerEmailUpdateEvent = "CUSTOMER_EMAIL_UPDATE"
	endpoint.CustomerPhoneUpdateEvent = "CUSTOMER_PHONE_UPDATE"
	endpoint.CustomerAddressUpdateEvent = "CUSTOMER_ADDRESS_UPDATE"
	//endpoint.CustomerLanguageUpdateEvent = "CUSTOMER_UPDATE"
}

func getSyncEvent(guid string, eventType string) types.SyncEvent {
	return types.SyncEvent{
		EventType: eventType,
		GUID:      guid,
	}
}
