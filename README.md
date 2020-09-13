<p align="center">
  <a href="https://epsagon.com" target="_blank" align="center">
    <img src="https://cdn2.hubspot.net/hubfs/4636301/Positive%20RGB_Logo%20Horizontal%20-01.svg" width="300">
  </a>
  <br />
</p>

[![Build Status](https://travis-ci.com/epsagon/epsagon-go.svg?token=wsveVqcNtBtmq6jpZfSf&branch=master)](https://travis-ci.com/epsagon/epsagon-go)
[![GoDoc](https://godoc.org/github.com/epsagon/epsagon-go?status.svg)](https://godoc.org/github.com/epsagon/epsagon-go)

# Epsagon Tracing for Go

This package provides tracing to Go applications for the collection of distributed tracing and performance metrics in [Epsagon](https://app.epsagon.com/?utm_source=github).


## Contents

- [Installation](#installation)
- [Frameworks](#frameworks)
- [Integrations](#integrations)
- [Configuration](#configuration)
- [Getting Help](#getting-help)
- [Opening Issues](#opening-issues)
- [License](#license)


## Installation

To install Epsagon, simply run:
```sh
go get github.com/epsagon/epsagon-go
```

Or using `dep`:
```sh
dep ensure -add github.com/epsagon/epsagon-go
```

## Frameworks

The following frameworks are supported by Epsagon:

|Framework                               |Supported Version          |
|----------------------------------------|---------------------------|
|[AWS Lambda](#aws-lambda)               |All                        |
|[Generic Function](#generic)            |All                        |


### AWS Lambda

Tracing Lambda functions can be done in the following method:

```go
package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
)

func myHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("In myHandler, received body: ", request.Body)
	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func main() {
	log.Println("enter main")
	lambda.Start(epsagon.WrapLambdaHandler(
        epsagon.NewTracerConfig("app-name-stage","epsagon-token"),
        myHandler))
}
```

### Generic

You can instrument a single function, this function can use go routines inside and their operations will still be traced.

```go
func doTask(a int, b string) (int, error) {
	log.Printf("inside doTask: b = %s", b)
	return a + 1, fmt.Errorf("boom")
}
func main() {
	// With Epsagon instrumentation
	config := epsagon.NewTracerConfig("generic-go-wrapper", "")
	config.Debug = true
	response := epsagon.GoWrapper(config, doTask)(5, "hello")
	res2 := response[0].Int()
	errInterface := response[1].Interface()
}
```
Optionally, you can pass a custom name for your wrapped function. In the epsagon dashboard,
your wrapped function will be displayed with your configured name in all the relevant screens:
traces search, service map and more.
```
	response := epsagon.GoWrapper(config, doTask, "<MyInstrumentedFuncName>")(5, "hello")
```

### Concurrent Generic
In order to support more than one function being traced in the same environment (using different goroutines), use this wrapper as shown in the example below. The wrapped function has to receive a context as its first parameter, and pass it to the relevant wrapped operations.


```go
func doTask(ctx context.Context, a int, b string, wg *sync.WaitGroup) (int, error) {
	defer wg.Done()
	log.Printf("inside doTask: b = %s", b)
	client := epsagonhttp.Wrap(http.Client{}, ctx)
	client.Get("https://epsagon.com/")
	return a + 1, fmt.Errorf("boom")
}

func main() {
	config := epsagon.NewTracerConfig("generic-go-wrapper", "")
	config.Debug = true
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		go epsagon.ConcurrentGoWrapper(config, doTask)(i, "hello", &wg)
	}
	wg.Wait()
	time.Sleep(2 * time.Second)
}
```
Optionally, you can pass a custom name for your wrapped function. In the epsagon dashboard,
your wrapped function will be displayed with your configured name in all the relevant screens:
traces search, service map and more.
```
		go epsagon.ConcurrentGoWrapper(config, doTask, "<MyInstrumentedFuncName>")(i, "hello", &wg)
```

## Integrations

Epsagon provides out-of-the-box instrumentation (tracing) for many popular frameworks and libraries.

|Library             |Supported Version          |
|--------------------|---------------------------|
|net/http            |Fully supported            |
|aws-sdk-go          |`>=1.10.0`                 |
|aws-sdk-go-v2       |`>=0.23.0`                 |

### net/http

Wrapping http requests using the net/http library can be done by wrapping the client:

```go
import (
	"github.com/epsagon/epsagon-go/wrappers/net/http"
...
	client := epsagonhttp.Wrap(http.Client{})
	resp, err := client.Get(anyurl)
```

If you want to disable data collection only for the calls made by this client set  `client.MetadataOnly = true`

### aws-sdk-go

Wrapping of aws-sdk-go is done through the Session object that has to be created to communicate with AWS:

```go
import (
...
	"github.com/epsagon/epsagon-go/wrappers/aws/aws-sdk-go/aws"
)
    ...
	sess := epsagonawswrapper.WrapSession(session.Must(session.NewSession()))
	svcSQS := sqs.New(sess)
```


### aws-sdk-go-v2

Wrapping of aws-sdk-go-v2 is done through the service object:
```go
import (
...
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/epsagon/epsagon-go/epsagon"
)

	svc := epsagon.WrapAwsV2Service(dynamodb.New(cfg)).(*dynamodb.Client)
    ...
```

## Configuration

Advanced options can be configured as a parameter to the `Config` struct to the `WrapLambdaHandler` or as environment variables.

|Parameter             |Environment Variable          |Type   |Default      |Description                                                                        |
|----------------------|------------------------------|-------|-------------|-----------------------------------------------------------------------------------|
|Token                 |EPSAGON_TOKEN                 |String |-            |Epsagon account token                                                              |
|ApplicationName       |-                             |String |-            |Application name that will be set for traces                                       |
|MetadataOnly          |EPSAGON_METADATA              |Boolean|`true`       |Whether to send only the metadata (`True`) or also the payloads (`False`)          |
|CollectorURL          |EPSAGON_COLLECTOR_URL         |String |-            |The address of the trace collector to send trace to                                |
|Debug                 |EPSAGON_DEBUG                 |Boolean|`False`      |Enable debug prints for troubleshooting                                            |
|SendTimeout           |EPSAGON_SEND_TIMEOUT_SEC      |String |`1s`         |The timeout duration to send the traces to the trace collector                     |



## Getting Help

If you have any issue around using the library or the product, please don't hesitate to:

* Use the [documentation](https://docs.epsagon.com).
* Use the help widget inside the product.
* Open an issue in GitHub.


## Opening Issues

If you encounter a bug with the Epsagon library for Go, we want to hear about it.

When opening a new issue, please provide as much information about the environment:
* Library version, Go runtime version, dependencies, etc.
* Snippet of the usage.
* A reproducible example can really help.

The GitHub issues are intended for bug reports and feature requests.
For help and questions about Epsagon, use the help widget inside the product.

## License

Provided under the MIT license. See LICENSE for details.

Copyright 2020, Epsagon
