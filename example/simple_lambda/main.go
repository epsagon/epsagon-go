package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
)

type Response events.APIGatewayProxyResponse

func Handler(ctx context.Context, event interface{}) (Response, error) {
	log.Println("In myHandler, received body: ", ctx)
	log.Println("In myHandler, received body: ", event)
	return Response {Body: "yes", StatusCode: 200}, nil
}

func main() {
	log.Println("enter main")
	config := epsagon.Config{
		ApplicationName: "erez-test-go",
		CollectorURL:    "http://dev.tc.epsagon.com"}
	lambda.Start(epsagon.WrapLambdaHandler(&config, Handler))
}
