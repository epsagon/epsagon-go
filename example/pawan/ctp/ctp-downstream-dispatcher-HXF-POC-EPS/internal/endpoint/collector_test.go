package endpoint

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollect_NoEventType__Error(t *testing.T) {
	event := ""

	endpoints, err := Collect(event)
	assert.Error(t, err)
	assert.Equal(t, []string([]string(nil)), endpoints, "expected array of downstream endpoints")
}

func TestCollect_UnknownEventType__Error(t *testing.T) {
	setEnvironmentVariables()
	event := "customer_event"

	endpoints, err := Collect(event)
	assert.Error(t, err, "expected error")
	assert.Equal(t, []string([]string(nil)), endpoints, "expected array of downstream endpoints")

}

func TestCollect_EmptyEndpoints__Error(t *testing.T) {
	event := "SUBSCRIPTION_EVENT"

	//set env var to empty string
	SubscriptionIntegrationSyncURI = ""
	SubscriptionIntegrationLangURI = ""

	endpoints, err := Collect(event)
	assert.Error(t, err, "expected error")
	assert.Equal(t, []string([]string(nil)), endpoints, "expected array of downstream endpoints")

}

func TestCollect_CustomerUpdateEvent_ExpectEndpoints__NoError(t *testing.T) {
	cases := []struct {
		event    string
		expected []string
	}{
		{
			event:    "CUSTOMER_UPDATE",
			expected: []string{"SubscriptionIntegrationSyncURI", "SubscriptionIntegrationLangURI"},
		},
	}

	setEnvironmentVariables()

	for _, c := range cases {
		endpoints, err := Collect(c.event)
		assert.NoError(t, err)
		assert.Equal(t, c.expected, endpoints, "expected array of downstream endpoints")
	}

}

func TestCollect_CustomerCreateEvent_ExpectEndpoints__NoError(t *testing.T) {
	cases := []struct {
		event    string
		expected []string
	}{
		{
			event:    "CUSTOMER_CREATE",
			expected: []string{"SubscriptionIntegrationSyncURI"},
		},
	}

	setEnvironmentVariables()

	for _, c := range cases {
		arns, err := Collect(c.event)
		assert.NoError(t, err)
		assert.Equal(t, c.expected, arns, "expected array of downstream endpoints")
	}

}

func TestCollect_SubscriptionEvent_ExpectEndpoints__NoError(t *testing.T) {
	cases := []struct {
		event    string
		expected []string
	}{
		{
			event:    "Subscription_Event",
			expected: []string{"SubscriptionIntegrationSyncURI"},
		},
	}

	setEnvironmentVariables()

	for _, c := range cases {
		endpoints, err := Collect(c.event)
		assert.NoError(t, err)
		assert.Equal(t, c.expected, endpoints, "expected array of downstream endpoints")
	}

}

func TestCollect_ProvisioningEvent_ExpectEndpoints__NoError(t *testing.T) {
	cases := []struct {
		event    string
		expected []string
	}{
		{
			event:    "Provisioning_Event",
			expected: []string{"SubscriptionIntegrationSyncURI"},
		},
	}

	setEnvironmentVariables()

	for _, c := range cases {
		endpoints, err := Collect(c.event)
		assert.NoError(t, err)
		assert.Equal(t, c.expected, endpoints, "expected array of downstream endpoints")
	}

}

func TestCollect_RemoteAuthEvent_ExpectEndpoints__NoError(t *testing.T) {
	cases := []struct {
		event    string
		expected []string
	}{
		{
			event:    "RemoteAuth_Event",
			expected: []string{"SubscriptionIntegrationSyncURI"},
		},
	}

	setEnvironmentVariables()

	for _, c := range cases {
		endpoints, err := Collect(c.event)
		assert.NoError(t, err)
		assert.Equal(t, c.expected, endpoints, "expected array of downstream endpoints")
	}

}

func TestCollect_HomeDealerEvent_ExpectEndpoints__NoError(t *testing.T) {
	cases := []struct {
		event    string
		expected []string
	}{
		{
			event:    "HomeDealer_Event",
			expected: []string{"SubscriptionIntegrationSyncURI"},
		},
	}

	setEnvironmentVariables()

	for _, c := range cases {
		endpoints, err := Collect(c.event)
		assert.NoError(t, err)
		assert.Equal(t, c.expected, endpoints, "expected array of downstream endpoints")
	}

}

func TestCollect_PaymentMethodFailureEvent_ExpectEndpoints__NoError(t *testing.T) {
	cases := []struct {
		event    string
		expected []string
	}{
		{
			event:    "PAYMENT_METHOD_FAILURE_EVENT",
			expected: []string{"SubscriptionIntegrationSyncURI"},
		},
	}

	setEnvironmentVariables()

	for _, c := range cases {
		endpoints, err := Collect(c.event)
		assert.NoError(t, err)
		assert.Equal(t, c.expected, endpoints, "expected array of downstream endpoints")
	}

}

func TestCollect_PaymentFailureEvent_ExpectEndpoints__NoError(t *testing.T) {
	cases := []struct {
		event    string
		expected []string
	}{
		{
			event:    "PAYMENT_FAILURE_EVENT",
			expected: []string{"SubscriptionIntegrationSyncURI"},
		},
	}

	setEnvironmentVariables()

	for _, c := range cases {
		endpoints, err := Collect(c.event)
		assert.NoError(t, err)
		assert.Equal(t, c.expected, endpoints, "expected array of downstream endpoints")
	}

}

func TestCollect_SubscriptionExpirationEvent_ExpectEndpoints__NoError(t *testing.T) {
	cases := []struct {
		event    string
		expected []string
	}{
		{
			event:    "SUBSCRIPTION_EXPIRATION_EVENT",
			expected: []string{"SubscriptionIntegrationSyncURI"},
		},
	}

	setEnvironmentVariables()

	for _, c := range cases {
		endpoints, err := Collect(c.event)
		assert.NoError(t, err)
		assert.Equal(t, c.expected, endpoints, "expected array of downstream endpoints")
	}

}

func TestCollect_WifiConsentEvent_ExpectEndpoints__NoError(t *testing.T) {
 	cases := []struct {
 		event    string
 		expected []string
 	}{
 		{
 			event:    "WIFI_CONSENT_EVENT",
 			expected: []string{"SubscriptionIntegrationSyncURI"},
 		},
 	}

 	setEnvironmentVariables()

 	for _, c := range cases {
 		endpoints, err := Collect(c.event)
 		assert.NoError(t, err)
 		assert.Equal(t, c.expected, endpoints, "expected array of downstream endpoints")
 	}

 }

 func TestCollect_SubAlertPurgeEvent_ExpectEndpoints__NoError(t *testing.T) {
 	cases := []struct {
 		event    string
 		expected []string
 	}{
 		{
 			event:    "SUBSCRIPTION_ALERT_PURGE_EVENT",
 			expected: []string{"SubscriptionIntegrationSyncURI"},
 		},
 	}

 	setEnvironmentVariables()

 	for _, c := range cases {
 		endpoints, err := Collect(c.event)
 		assert.NoError(t, err)
 		assert.Equal(t, c.expected, endpoints, "expected array of downstream endpoints")
 	}

 }
func setEnvironmentVariables() {
	SubscriptionIntegrationSyncURI = "SubscriptionIntegrationSyncURI"
	SubscriptionIntegrationLangURI = "SubscriptionIntegrationLangURI"
	CustomerCreateEvent = "CUSTOMER_CREATE"
	CustomerUpdateEvent = "CUSTOMER_UPDATE"
	SubscriptionEvent = "SUBSCRIPTION_EVENT"
	CustomerEmailUpdateEvent = "CUSTOMER_EMAIL_UPDATE"
	CustomerPhoneUpdateEvent = "CUSTOMER_PHONE_UPDATE"
	CustomerAddressUpdateEvent = "CUSTOMER_ADDRESS_UPDATE"
	//CustomerLanguageUpdateEvent = "CUSTOMER_UPDATE"
	ProvisioningEvent = "PROVISIONING_EVENT"
	RemoteAuthEvent = "REMOTEAUTH_EVENT"
	HomeDealerEvent = "HOMEDEALER_EVENT"
	PaymentMethodFailureEvent = "PAYMENT_METHOD_FAILURE_EVENT"
	PaymentFailureEvent = "PAYMENT_FAILURE_EVENT"
	SubscriptionExpirationEvent = "SUBSCRIPTION_EXPIRATION_EVENT"
	WifiConsentEvent = "WIFI_CONSENT_EVENT"
	SubAlertPurgeEvent = "SUBSCRIPTION_ALERT_PURGE_EVENT"
}
