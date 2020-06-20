package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/wrappers/aws/aws-sdk-go/aws"
)

func myHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("In myHandler, received body: ", request.Body)

	// sess := session.Must(session.NewSession())
	// epsagon wrapper for aws-sdk-go
	sess := epsagonawswrapper.WrapSession(session.Must(session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
		})))
	message := "test message"

	svcSNS := sns.New(sess)

	topicArn := os.Getenv("SNS_ARN")
	result, err := svcSNS.Publish(&sns.PublishInput{
		Message:  &message,
		TopicArn: &topicArn,
	})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to send SNS message: %v", err)
		log.Println(errMsg)
		return events.APIGatewayProxyResponse{Body: errMsg, StatusCode: 500}, err
	}
	log.Printf("Result message id: %s\n", *result.MessageId)
	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("sns-sqs-test-go", "")
	config.Debug = true
	lambda.Start(epsagon.WrapLambdaHandler(config, myHandler))
}
