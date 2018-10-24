# Epsagon Go Instrumentation

[![Build Status][1]][2] [![GoDoc][3]][4]

## Installing
```
go get github.com/epsagon/com
```
Or using `dep`:
```
dep ensure -add github.com/epsagon/com
```
## Usage

### To wrap a lambda handler
```
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
		"APPLICATION-NAME", os.Environ("EPSAGON_TOKEN"),
		"http://tc.epsagon.com", false, myHandler,
	))
}
```

`epsagon.WrapLambdaHandler` will wrap your handler with code that will start a tracer that will wait for events and will send them to the collector when the lambda finishes

### Wrapping other libraries
TODO

## Configuration


