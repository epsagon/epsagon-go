package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/google/uuid"
	"log"
	"os"
)

func ddbHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("Failed to load default aws config")
	}
	cfg.Region = "eu-west-1"
	svc := epsagon.WrapAwsV2Service(dynamodb.New(cfg)).(*dynamodb.Client)
	putItemInput := dynamodb.PutItemInput{
		Item: map[string]dynamodb.AttributeValue{
			"item":    {S: aws.String(uuid.New().String())},
			"request": {S: &request.Body},
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")),
	}
	req := svc.PutItemRequest(&putItemInput)

	resp, err := req.Send(context.Background())
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body: fmt.Sprintf("PutItem Failed: %s\n%s",
				resp.String(), err.Error()),
			StatusCode: 500,
		}, nil
	}

	log.Println("Successfully written item to table")
	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("ddb-test-go-v2", "")
	config.Debug = true
	lambda.Start(epsagon.WrapLambdaHandler(config, ddbHandler))
	log.Println("exit main")
}
