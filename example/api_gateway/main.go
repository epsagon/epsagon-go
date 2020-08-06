package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/wrappers/net/http"
	"io/ioutil"
	"log"
	"net/http"
)

// Response is an API gateway response type
type Response events.APIGatewayProxyResponse

func myHandler(request events.APIGatewayProxyRequest) (Response, error) {
	log.Println("In myHandler, received body: ", request.Body)
	client := epsagonhttp.Wrap(http.Client{})
	resp, err := client.Get("https://api.randomuser.me")
	var body string
	if err == nil {
		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			body = string(bodyBytes)
		}
	} else {
		log.Printf("Error in getting response: %+v\n", err)
	}
	return Response{Body: "yes: random user: " + body, StatusCode: 200}, nil
}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("epsagon-test-go", "")
	lambda.Start(epsagon.WrapLambdaHandler(config, myHandler))
}
