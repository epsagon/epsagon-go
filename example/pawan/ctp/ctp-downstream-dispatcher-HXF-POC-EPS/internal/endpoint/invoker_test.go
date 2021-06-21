package endpoint

import (
	"ctp-downstream-dispatcher/internal/types"
	httpx "ctp-downstream-dispatcher/pkg/http"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// func TestInvoke_MissingEndpoint__Error(t *testing.T) {
// 	client := &httpx.Client{
// 		HTTPClient: &http.Client{},
// 	}
// 	payload := getSyncEvent("someGuid", "someEventType")
// 	response, err := Invoke(client, "", payload)
// 	assert.Error(t, err)
// 	assert.EqualError(t, err, "missing target endpoint")
// 	assert.Nil(t, response)
// }

// func TestInvoke_MissingPayload__Error(t *testing.T) {
// 	client := &httpx.Client{
// 		HTTPClient: &http.Client{},
// 	}
// 	response, err := Invoke(client, "someEndpoint", nil)
// 	assert.Error(t, err)
// 	assert.EqualError(t, err, "missing payload")
// 	assert.Nil(t, response)
// }

////TODO
//func TestInvoke_InvalidPayload__Error(t *testing.T) {
//	client := &httpx.Client{
//		HTTPClient: &http.Client{},
//	}
//	response, err := Invoke(client, "someEndpoint", nil)
//	assert.Error(t, err)
//	assert.EqualError(t, err, "invalid payload")
//	assert.Nil(t, response)
//
//}

////TODO
func TestInvoke_PostErrorResponse__Error(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer ts.Close()

	client := httpx.NewClient(
		httpx.WithHTTPTimeout(10 * time.Second),
	)
	//client := &httpx.Client{
	//	HTTPClient: &http.Client{},
	//}
	payload := getSyncEvent("someGuid", "someEventType")
	response, err := Invoke(client, "http://10.0.0.1/someendpoint", payload)
	assert.Error(t, err)
	//assert.EqualError(t, err, "failed to invoke sync endpoint")
	assert.Nil(t, response)

}

func getSyncEvent(guid string, eventType string) *types.SyncEvent {
	return &types.SyncEvent{
		EventType: eventType,
		GUID:      guid,
	}
}
