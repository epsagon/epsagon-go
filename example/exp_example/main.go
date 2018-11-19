package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
	"runtime/debug"
)

func expHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("In expHandler, received body: ", request.Body)
	zero := 0
	_ = 1 / zero
	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func wtfokay(err *ExceptionData) (hi string) {
	defer func() {
		r := recover()
		if r != nil {
			err.err = fmt.Sprintf("%v", r)
			err.stackTrace = string(debug.Stack())
			panic(r)
		}
	}()

	zero := 0
	res := 12 / zero
	return fmt.Sprintf("%v",res)
}

type ExceptionData struct {
	err string
	stackTrace string
}

func main() {
	exp := &ExceptionData{}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("i'm here to catch exceptions, %v\n%v", exp.err, exp.stackTrace)
			//panic(r)
		}
	}()

	res := wtfokay(exp)
	log.Printf("\n\n\n\n\nhi there %s!!!!!", string(res))
	return
	log.Println("enter main")
	config := epsagon.Config{
		ApplicationName: "exp-test-go",
		Debug:           true,
	}
	lambda.Start(epsagon.WrapLambdaHandler(&config, expHandler))
	log.Println("exit main")
}
