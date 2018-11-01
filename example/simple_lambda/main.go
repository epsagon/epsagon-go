package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
)

type Response events.APIGatewayProxyResponse

func myHandler(request events.APIGatewayProxyRequest) (Response, error) {
	log.Println("In myHandler, received body: ", request.Body)
	return Response {Body: "yes", StatusCode: 200}, nil
}

func main() {
	log.Println("enter main")
	config := epsagon.Config{
		ApplicationName: "epsagon-test-go",
		CollectorURL:    "http://dev.tc.epsagon.com"}
	lambda.Start(epsagon.WrapLambdaHandler(&config, myHandler))
}
