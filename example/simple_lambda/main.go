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
		"erez-test-go",
		"4e988052-fb55-4c6d-b46f-f17f3dd81b6b",
		"http://dev.tc.epsagon.com",
		false,
		myHandler,
	))
}
