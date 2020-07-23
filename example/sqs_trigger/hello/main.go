package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/wrappers/aws/aws-sdk-go/aws"
)

func myHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("In myHandler, received body: ", request.Body)

	// sess := session.Must(session.NewSession())
	// epsagon wrapper for aws-sdk-go
	sess := epsagonawswrapper.WrapSession(session.Must(session.NewSession()))

	svcSQS := sqs.New(sess)

	sqsQueueName := os.Getenv("SQS_NAME")
	queueURL, err := svcSQS.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: &sqsQueueName})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get SQS Queue Url: %v", err)
		log.Println(errMsg)
		return events.APIGatewayProxyResponse{Body: errMsg, StatusCode: 500}, err
	}
	_, err = svcSQS.SendMessage(&sqs.SendMessageInput{
		MessageBody: &request.Body,
		QueueUrl:    queueURL.QueueUrl,
	})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to send SQS message: %v", err)
		log.Println(errMsg)
		return events.APIGatewayProxyResponse{Body: errMsg, StatusCode: 500}, err
	}
	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("sqs-test-go", "")
	config.Debug = true
	lambda.Start(epsagon.WrapLambdaHandler(config, myHandler))
}
