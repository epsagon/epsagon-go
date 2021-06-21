# CTP Downstream Dispatcher
This is the CTP Downstream Dispatcher solution for sync'ing data to downstream/external systems triggered by create and update calls from the CTP customer profile and subscription services.

The following is a map of event triggers and their target systems:

	|	CT Event	|	Target System	| Payload      | API Endpoints for Sync            |
	|                       |                       | Format       |                                   |
	|_______________________|_______________________|______________|___________________________________|
	| create account	| 1. Customer Central	| cc-specified | CUSTOMER_CENTRAL_SYNC_URI         |
	| (orchcca service)	| 2. Servco		| combined     | SERVCO_SYNC_URI                   |
	|_______________________|_______________________|______________|___________________________________|
	| update account	| 1. Customer Central	| cc-specified | CUSTOMER_CENTRAL_SYNC_URI         |
	| (orchcu service)	| 2. Servco		| combined     | SERVCO_SYNC_URI                   |
	|			| 3. TC Safety Services	| combined     | TC_SAFETY_SYNC_URI                |
	|_______________________|_______________________|______________|___________________________________|
	| create subscription	| 1. Servco		| combined     | SERVCO_SYNC_URI                   |
	| (orch service)	| 2. TC			| combined     | TC_SYNC_URI                       |
	|_______________________|_______________________|______________|___________________________________|
	| update subscription	| 1. Servco		| combined     | SERVCO_SYNC_URI                   |
	| (orch service)	| 2. TC			| combined     | TC_SYNC_URI                       |
	|_______________________|_______________________|______________|___________________________________|


## Architecture Diagram
![Diagram](/diagram.png)

Refer to #1 and #2 in the diagram.
The data sync is triggered when there's an HTTP request to CTP service APIs, such that:
1. If create-account API is invoked, the orchcca service will send a message to SQS queue, ctp-downstream-data-sync, in the format of 
```{"eventType":"Subscription_Event","guid":"some123guid12345678someguid2d0a2","vin":"SOMETESTVIN000627","headers":{"X-CorrelationId":"12345678-1234-1234-1234-123456789012"}}
```
2. The queue will trigger the Lambda function, ctp-downstream-data-sync-dispatcher (Dispatcher)
3. The Dispatcher will then send the SQS message by invoking the API (microservice*) endpoints designated to perform the sync (see table above).

*** The microservice that handles the sync will:
- call OCPR (orchcd layer) to get customer details
- build HTTP request with customer details as request body
- send HTTP request to target system


## AWS Resource Management
AWS Resource Management for Lambda Function, Lambda DLQ, SQS, SQS DLQ, and Security Groups are specified in serverless.yml resources.


## Prerequisites
[Go installation](https://golang.org/dl/)

[Serverless Framework](https://serverless.com/framework/docs/getting-started/)

[AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html)

[MingGW-w64 for Windows](https://sourceforge.net/projects/mingw-w64/files/mingw-w64/mingw-w64-release/)


## Installation
1. Clone the project to $USERPROFILE/Documents/<your-workspace>
2. Open a terminal, download dependencies, build, and test by running the following commands:
```bash
cd $USERPROFILE/Documents/<your-workspace>/ctp-downstream-dispatcher
go mod tidy
```
```bash
make build
make test
```

## Usage
To add an endpoint for the Dispatcher to invoke:
1. Add a variable in ```env.yml``` for the API URL
```
dev:
    ENV_VAR_NAME_URI: https://sync-handler-endpoint

qa:
    ENV_VAR_NAME_URI: https://sync-handler-endpoint

stg:
    ENV_VAR_NAME_URI: https://sync-handler-endpoint

prod:
    ENV_VAR_NAME_URI: https://sync-handler-endpoint

```
2. In the collector.go (~/ctp-downstream-dispatcher/internal/endpoint/collector.go), add and call the variable you added in the env.yml:
```
line 11
var (
	//----- line 12 skipped -------//
	
	//VarNameURI sync handler endpoint
	VarNameURI = os.Getend("ENV_VAR_NAME_URI")
	
)
```
```
line 34
//Collect function is used to collect endpoints based on CtpAPI event type
func Collect(event string) ([]string, error) {
	var endpoints Endpoints
	var err error

    //-------->> add here the endpoints in the array per event type if applicable -------->>
	customerCreateSyncEndpoints := []string{CustCentralSyncURI, ServcoSyncURI}
	customerUpdateSyncEndpoints := []string{CustCentralSyncURI, ServcoSyncURI, TcSafetySyncURI}
	subscriptionSyncEndpoints := []string{TcSyncURI, ServcoSyncURI}

	//check event type
	switch event {
    //-------->> add here the switch case per event type if applicable -------->>
	case constants.CustomerCreateEvent, strings.ToLower(constants.CustomerCreateEvent):
		endpoints.Add(customerCreateSyncEndpoints)
	case constants.CustomerUpdateEvent, strings.ToLower(constants.CustomerUpdateEvent):
		endpoints.Add(customerUpdateSyncEndpoints)
	case constants.SubscriptionEvent, strings.ToLower(constants.SubscriptionEvent):
		endpoints.Add(subscriptionSyncEndpoints)
	default:
		err = errors.Errorf("undefined CTP event type (%s)", event)
	}
```