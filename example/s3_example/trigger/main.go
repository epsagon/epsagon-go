package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"log"
	"runtime"
	"runtime/debug"

	//	"time"
)

func handler(s3Event events.S3Event) {
	for _, record := range s3Event.Records {
		s3 := record.S3
		fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n",
			record.EventSource, record.EventTime, s3.Bucket.Name, s3.Object.Key)
	}
}

func main() {
	log.Println("starting main()")
	var thrown bool
	thrown = true
	defer handleErrors(&thrown)
	handler(nil)
	thrown = false
	log.Println("ending main()")

	//config := epsagon.Config{
	//	ApplicationName: "epsagon-s3-test-go",
	//	CollectorURL:    "http://dev.tc.epsagon.com",
	//	Debug: true}
	//lambda.Start(epsagon.WrapLambdaHandler(&config, handler))
}
