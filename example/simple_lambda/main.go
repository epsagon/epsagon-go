package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
)

// Response is an API gateway response type
type Response events.APIGatewayProxyResponse

func myHandler(request events.APIGatewayProxyRequest) (Response, error) {
	log.Println("In myHandler, received body: ", request.Body)
	return Response{Body: "yes", StatusCode: 200}, nil
}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("epsagon-test-go", "")
	lambda.Start(epsagon.WrapLambdaHandler(config, myHandler))
}
