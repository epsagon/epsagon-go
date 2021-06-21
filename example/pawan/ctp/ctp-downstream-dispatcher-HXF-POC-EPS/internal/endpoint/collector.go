package endpoint

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

var (
	//Target URIs

	//SubscriptionIntegrationLangURI subscription integration language update handler endpoint
	SubscriptionIntegrationLangURI = os.Getenv("SUB_INT_LANG_URI")

	//SubscriptionIntegrationSyncURI subscription integration sync handler endpoint
	SubscriptionIntegrationSyncURI = os.Getenv("SUB_INT_SYNC_URI")

	//Event Triggers

	// CustomerCreateEvent is used for checking the flag set at the
	// CtpAPI layer for any front-end invoking the orchcca API
	CustomerCreateEvent = os.Getenv("CUSTOMER_CREATE")

	// CustomerUpdateEvent is used for checking the flag set at the
	// CtpAPI layer for any front-end invoking the orchcu API
	CustomerUpdateEvent = os.Getenv("CUSTOMER_UPDATE")

	// CustomerEmailUpdateEvent is used for checking the flag set at the
	// CtpAPI layer for any front-end invoking the orchcu API
	CustomerEmailUpdateEvent = os.Getenv("CUSTOMER_EMAIL_UPDATE")

	// CustomerPhoneUpdateEvent is used for checking the flag set at the
	// CtpAPI layer for any front-end invoking the orchcu API
	CustomerPhoneUpdateEvent = os.Getenv("CUSTOMER_PHONE_UPDATE")

	// CustomerAddressUpdateEvent is used for checking the flag set at the
	// CtpAPI layer for any front-end invoking the orchcu API
	CustomerAddressUpdateEvent = os.Getenv("CUSTOMER_ADDRESS_UPDATE")

	//  CustomerLanguageUpdateEvent is used for checking the flag set at the
	// CtpAPI layer for any front-end invoking the orchcu API
	//CustomerLanguageUpdateEvent = os.Getenv("CUSTOMER_LANGUAGE_UPDATE")

	// SubscriptionEvent is used for checking the flag set at the
	// CtpAPI layer for any front-end invoking the orchsub API
	SubscriptionEvent = os.Getenv("SUBSCRIPTION_EVENT")

	// ProvisioningEvent is used for checking the flag set at the
	// CtpAPI layer for any front-end invoking the orchsub API
	ProvisioningEvent = os.Getenv("PROVISIONING_EVENT")

	// RemoteAuthEvent is used for checking the flag set at the
	// CtpAPI layer for any front-end invoking the orchsub API
	RemoteAuthEvent = os.Getenv("REMOTE_AUTH_EVENT")

	// HomeDealerEvent is used for checking the flag set at the
	// CtpAPI layer for any front-end invoking the orchsub API
	HomeDealerEvent = os.Getenv("HOME_DEALER_EVENT")

	// PaymentMethodFailureEvent is used for checking the flag set at the
	// CtpAPI layer for any front-end invoking the orchsub API
	PaymentMethodFailureEvent = os.Getenv("PAYMENT_METHOD_FAILURE_EVENT")

	// PaymentFailureEvent is used for checking the flag set at the
	// CtpAPI layer for any front-end invoking the orchsub API
	PaymentFailureEvent = os.Getenv("PAYMENT_FAILURE_EVENT")

	// SubscriptionExpirationEvent is used for checking the flag set at the
	// CtpAPI layer for any front-end invoking the orchsub API
	SubscriptionExpirationEvent = os.Getenv("SUBSCRIPTION_EXPIRATION_EVENT")

	// WifiConsentEvent is an event raised by customer-consent-service whenever the
	// customer gets wifi service in their car.
    WifiConsentEvent = os.Getenv("WIFI_CONSENT_EVENT")

    // SubAlertPurgeEvent is an event raised by sub-alerts-job which works to create alert for
    // payment_method expiration, subscription expiration and payment method, and
    // payment method failures - when such alerts are created in the db,
    // its purged after 30 days and an event is raised so as to notify the downstream systems.
    SubAlertPurgeEvent = os.Getenv("SUBSCRIPTION_ALERT_PURGE_EVENT")
)

// Endpoints is a list of downstream endpoints
type Endpoints []string

// Add adds to the Resource Names
func (e *Endpoints) Add(endpoints []string) {
	for _, endpoint := range endpoints {
		if endpoint != "" {
			*e = append(*e, endpoint)
		}
	}
}

//Collect function is used to collect endpoints based on CtpAPI event type
func Collect(event string) ([]string, error) {
	endpoints := Endpoints{}

	var err error

	customerCreateSyncEndpoints := []string{SubscriptionIntegrationSyncURI}
	customerUpdateSyncEndpoints := []string{SubscriptionIntegrationSyncURI, SubscriptionIntegrationLangURI}
	customerEmailUpdateSyncEndpoints := []string{SubscriptionIntegrationSyncURI}
	customerPhoneUpdateSyncEndpoints := []string{SubscriptionIntegrationSyncURI}
	customerAddressUpdateSyncEndpoints := []string{SubscriptionIntegrationSyncURI}
	// customerLanguageUpdateSyncEndpoints := []string{SubscriptionIntegrationSyncURI, SubscriptionIntegrationLangURI}
	subscriptionSyncEndpoints := []string{SubscriptionIntegrationSyncURI}
	provisioningSyncEndpoints := []string{SubscriptionIntegrationSyncURI}
	remoteAuthEventSyncEndpoints := []string{SubscriptionIntegrationSyncURI}
	homeDealerEventSyncEndpoints := []string{SubscriptionIntegrationSyncURI}
	paymentMethodFailureEventSyncEndpoints := []string{SubscriptionIntegrationSyncURI}
	paymentFailureEventSyncEndpoints := []string{SubscriptionIntegrationSyncURI}
	subscriptionExpirationEventSyncEndpoints := []string{SubscriptionIntegrationSyncURI}
	wifiConsentEventEndpoints := []string{SubscriptionIntegrationSyncURI}
	subAlertPurgeEventEndpoints := []string{SubscriptionIntegrationSyncURI}

	endpoints.Add(customerCreateSyncEndpoints)

	evt := strings.ToUpper(event)
	// check event type
	switch evt {
	case CustomerCreateEvent:
		endpoints.Add(customerCreateSyncEndpoints)
	case CustomerUpdateEvent:
		endpoints.Add(customerUpdateSyncEndpoints)
	case CustomerEmailUpdateEvent:
		endpoints.Add(customerEmailUpdateSyncEndpoints)
	case CustomerPhoneUpdateEvent:
		endpoints.Add(customerPhoneUpdateSyncEndpoints)
	case CustomerAddressUpdateEvent:
		endpoints.Add(customerAddressUpdateSyncEndpoints)
	// case CustomerLanguageUpdateEvent:
	//	endpoints.Add(customerLanguageUpdateSyncEndpoints)
	case SubscriptionEvent:
		endpoints.Add(subscriptionSyncEndpoints)
	case ProvisioningEvent:
		endpoints.Add(provisioningSyncEndpoints)
	case RemoteAuthEvent:
		endpoints.Add(remoteAuthEventSyncEndpoints)
	case HomeDealerEvent:
		endpoints.Add(homeDealerEventSyncEndpoints)
	case PaymentMethodFailureEvent:
		endpoints.Add(paymentMethodFailureEventSyncEndpoints)
	case PaymentFailureEvent:
		endpoints.Add(paymentFailureEventSyncEndpoints)
	case SubscriptionExpirationEvent:
		endpoints.Add(subscriptionExpirationEventSyncEndpoints)
	case WifiConsentEvent:
    	endpoints.Add(wifiConsentEventEndpoints)
    case SubAlertPurgeEvent:
        endpoints.Add(subAlertPurgeEventEndpoints)
	default:
		err = errors.Errorf("undefined CTP event type (%s)", event)
	}

	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, err
	}

	if len(endpoints) == 0 {
		err := errors.Errorf("no target endpoints found (%+v)", endpoints)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	log.Info().Msgf("endpoints collected: %+v", endpoints)
	return endpoints, nil
}
