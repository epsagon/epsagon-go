package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
)

func ddbHandler(ddbEvent events.DynamoDBEvent) error {
	log.Println("In mySQSHandler, received body: ", ddbEvent)
	return nil
}

func main() {
	log.Println("enter main")
	config := epsagon.Config{
		ApplicationName: "ddb-test-go",
		Debug: true,
	}
	lambda.Start(epsagon.WrapLambdaHandler(&config, ddbHandler))
	log.Println("exit main")
}
