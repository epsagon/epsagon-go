package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
)

func ddbHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic("Failed to load default aws config")
	}
	cfg.Region = "eu-west-1"
	svc := epsagon.WrapAwsV2Service(sts.NewFromConfig(cfg)).(*sts.Client)
	//req := svc.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})

	resp, err := svc.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body: fmt.Sprintf("GetCallerIdentityRequest Failed: %s\n%s",
				aws.ToString(resp.UserId), err.Error()),
			StatusCode: 500,
		}, nil
	}

	log.Println("Successfully got caller identity request")
	return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}, nil
}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("ddb-test-go-v2", "")
	config.Debug = true
	lambda.Start(epsagon.WrapLambdaHandler(config, ddbHandler))
	log.Println("exit main")
}
