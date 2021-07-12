package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/epsagon/epsagon-go/epsagon"
)


func readHandler() {

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic("Failed to load default aws config")
	}
	//cfg.Region = "us-east-1"
	svc := epsagon.WrapAwsV2Service(s3.NewFromConfig(cfg)).(*s3.Client)

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

	myBucket := "BUCKETNAME"
	key := "KEYNAME"


	fmt.Println(ctx)


	respListObj, err := svc.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: &myBucket,
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(respListObj)



	respGetObj, err := svc.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &myBucket,
		Key:    &key,
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(respGetObj)

}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("s3-test-go", "")
	config.MetadataOnly = false
	config.Debug = true
	epsagon.GoWrapper(config, readHandler)()
	log.Println("exit main")
}


