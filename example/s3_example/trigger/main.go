package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
)

func s3Handler(s3Event events.S3Event) {
	for _, record := range s3Event.Records {
		s3 := record.S3
		fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n",
			record.EventSource, record.EventTime, s3.Bucket.Name, s3.Object.Key)
	}
}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("s3-test-go", "")
	config.Debug = true
	lambda.Start(epsagon.WrapLambdaHandler(config, s3Handler))
	log.Println("exit main")
}
