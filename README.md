# Epsagon Go Instrumentation

[![Build Status][1]][2] [![GoDoc][3]][4]

[1]: https://travis-ci.com/epsagon/epsagon-go.svg?branch=master
[2]: https://travis-ci.com/epsagon/epsagon-go
[3]: https://godoc.org/github.com/epsagon/epsagon-go?status.svg
[4]: https://godoc.org/github.com/epsagon/epsagon-go

## Installing
```
go get github.com/epsagon/epsagon-go
```
Or using `dep`:
```
dep ensure -add github.com/epsagon/epsagon-go
```
## Usage

### To wrap a lambda handler
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
        &epsagon.Config{ApplicationName: "APPLICATION-NAME"},
        myHandler))
}
```

`epsagon.WrapLambdaHandler` will wrap your handler with code that will start a tracer that will wait for events and will send them to the collector when the lambda finishes

### Wrapping other libraries
#### aws-sdk-go
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

## Configuration
The epsagon.Config structure that is sent to `WraphLambdaHandler` has some fields to customize epsagons behaviour:
- `MetadataOnly`: Boolean flag to make sure to only send metadata information and not the data itself
- `Debug`: Will print debug information form epsagon
- `CollectorURL`: Set the collector's address instead of getting it from `EPSAGON_COLLECTOR_URL` or using epsagon's default in the same aws region
- `Token`: Your epsagon token (The default is to attempt to read this from `EPSAGON_TOKEN` environment variable)
