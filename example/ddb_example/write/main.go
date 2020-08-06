package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/wrappers/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
	"log"
	"os"
)

// Item example
type Item struct {
	Item string `json:"item"`
}

func ddbHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("In ddbHandler, received body: ", request.Body)

	session := epsagonawswrapper.WrapSession(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)))

	svc := dynamodb.New(session)

	item := Item{
		Item: uuid.New().String(),
	}

	av, marshalErr := dynamodbattribute.MarshalMap(item)

	if marshalErr != nil {
		marshErrMsg := fmt.Sprintf("Failed to marshal table: %v", marshalErr)
		log.Println(marshErrMsg)
		return events.APIGatewayProxyResponse{Body: marshErrMsg, StatusCode: 500}, marshalErr
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(os.Getenv("TABLE_NAME")),
	}

	_, err := svc.PutItem(input)

	if err != nil {
		fmt.Println("Failed calling PutItem:")
		errMsg := fmt.Sprintf("Failed calling PutItem: %v", err.Error())
		return events.APIGatewayProxyResponse{Body: errMsg, StatusCode: 500}, err
	}

	fmt.Println("Successfully written item to table")
	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("ddb-test-go-v2", "")
	config.Debug = true
	lambda.Start(epsagon.WrapLambdaHandler(config, ddbHandler))
	log.Println("exit main")
}
