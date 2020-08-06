package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/google/uuid"
)

func s3WriteHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("In s3WriteHandler, received body: ", request.Body)

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("Failed to load default aws config")
	}
	cfg.Region = "eu-west-1"
	svc := epsagon.WrapAwsV2Service(s3.New(cfg)).(*s3.Client)

	// Create a context with a timeout that will abort the upload if it takes
	// more than the passed in timeout.
	ctx := context.Background()
	var cancelFn func()
	timeout := time.Minute * 2
	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
	}
	// Ensure the context is canceled to prevent leaking.
	// See context package for more information, https://golang.org/pkg/context/
	defer cancelFn()

	data := "Hello World"
	myBucket := os.Getenv("BUCKET_NAME")
	key := uuid.New().String()

	if len(myBucket) == 0 {
		errMsg := "ERROR: You need to assign BUCKET_NAME env"
		log.Println(errMsg)
		return events.APIGatewayProxyResponse{Body: errMsg, StatusCode: 500}, nil
	}

	req := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: &myBucket,
		Key:    &key,
		Body:   strings.NewReader(data),
	})

	if _, err := req.Send(ctx); err != nil {
		var cerr *aws.RequestCanceledError
		var errMsg string
		if errors.As(err, &cerr) {
			errMsg = fmt.Sprintf("upload caceled due to timeout, %v", err)
		} else {
			errMsg = fmt.Sprintf("Failed to upload object: %v", err)
		}
		log.Println(errMsg)
		return events.APIGatewayProxyResponse{Body: errMsg, StatusCode: 500}, err
	}

	log.Printf("Succesfully uploaded file.")
	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("s3-test-go", "")
	config.Debug = true
	lambda.Start(epsagon.WrapLambdaHandler(config, s3WriteHandler))
	log.Println("exit main")
}
