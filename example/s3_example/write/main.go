package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/wrappers/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
	"log"
	"os"
)

func s3WriteHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("In s3WriteHandler, received body: ", request.Body)

	session := epsagonawswrapper.WrapSession(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)))

	uploader := s3manager.NewUploader(session)

	byteData := []byte("Hello World")
	myBucket := os.Getenv("BUCKET_NAME")
	key := uuid.New().String()

	if len(myBucket) == 0 {
		errMsg := "ERROR: You need to assign BUCKET_NAME env"
		log.Println(errMsg)
		return events.APIGatewayProxyResponse{Body: errMsg, StatusCode: 500}, nil
	}

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(byteData),
	})

	if err != nil {
		errMsg := fmt.Sprintf("Failed calling Upload: %v", err.Error())
		log.Println(errMsg)
		return events.APIGatewayProxyResponse{Body: errMsg, StatusCode: 500}, err
	}

	log.Printf("Succesfully uploaded file to %s\n", result.Location)
	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("s3-test-go", "")
	config.Debug = true
	lambda.Start(epsagon.WrapLambdaHandler(config, s3WriteHandler))
	log.Println("exit main")
}
