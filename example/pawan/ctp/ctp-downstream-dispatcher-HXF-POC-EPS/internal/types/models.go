package types

import "net/http"

// SyncEvent is the struct to handle the incoming
// message from SQS
// messge format
/**
 * {
 * 		"eventType"	:"CUSTOMER_UPDATE",
 * 		"guid"		: "some guid"
 * 		"headers" : {
 * 			##required headers
 * 		}
 * }
 */
type SyncEvent struct {
	// EventType is the flag that gets set at the CTP layer
	EventType string `json:"eventType"`

	// GUID is the identifier of the account the customer/vehicle
	// is associated with
	GUID string `json:"guid"`

	// VIN is the vehicle identifier
	VIN string `json:"vin"`

	// wifi
    	Wifi string `json:"wifi"`

    // wifiAcceptedDate
        WifiAcceptedDate string `json:"wifiAcceptedDate"`

    // wifiDeclinedDate
        WifiDeclinedDate string `json:"wifiDeclinedDate"`

	//Header can contain X-CorrelationID or any related headers
	Header Header `json:"headers"`
}

//Header can contain X-CorrelationID or any related headers
type Header struct {
	XCorrelationID string `json:"X-CorrelationId"`
	ACMHeaders     http.Header
}
