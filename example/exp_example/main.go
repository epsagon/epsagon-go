package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
)

func expHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("In expHandler, received body: ", request.Body)
	zero := 0
	_ = 1 / zero
	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("exp-test-go", "")
	lambda.Start(epsagon.WrapLambdaHandler(config, expHandler))
	log.Println("exit main")
}
