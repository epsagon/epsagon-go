package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
)

func mySQSHandler(sqsEvents events.SQSEvent) error {
	log.Println("In mySQSHandler, received body: ", sqsEvents)
	return nil
}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("sqs-test-go", "")
	lambda.Start(epsagon.WrapLambdaHandler(config, mySQSHandler))
}
