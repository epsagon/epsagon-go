package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
	"reflect"
)

func ddbHandler()  {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic("Failed to load default aws config")
	}
	cfg.Region = "eu-west-1"
	svc := epsagon.WrapAwsV2Service(dynamodb.NewFromConfig(cfg)).(*dynamodb.Client)
	fmt.Println(svc)
	fmt.Println(reflect.TypeOf(svc))

	resp, err := svc.GetItem(context.Background(), &dynamodb.GetItemInput{
		Key:                      map[string]types.AttributeValue{
			"attr_name": &types.AttributeValueMemberS{Value: "example-val"},
		},
		TableName:                aws.String("application-table-dev"),
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp)

	putResp, err := svc.PutItem(context.Background(), &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			"attr_name": &types.AttributeValueMemberS{Value: "attr_value"},
		},
		TableName: aws.String("table_name"),
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(putResp)

}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("ddb-test-go", "")
	config.MetadataOnly = true
	config.Debug = true
	epsagon.GoWrapper(config, ddbHandler)()
	log.Println("exit main")
}
