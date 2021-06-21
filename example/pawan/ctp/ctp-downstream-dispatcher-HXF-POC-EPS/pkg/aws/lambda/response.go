package lambda

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog/log"
)

var (
	responseEmpty = events.APIGatewayProxyResponse{}
)

//Error takes an error and returns a standard Gateway appropriate with valid status codes
func Error(err error) (*events.APIGatewayProxyResponse, error) {
	log.Error().Err(err).Msg("")
	code := http.StatusInternalServerError
	if gwErr, ok := err.(APIGatewayError); ok {
		code = gwErr.StatusCode()
	}
	return &events.APIGatewayProxyResponse{StatusCode: code, Body: err.Error()}, nil
}

//JSON takes a resp and returns an API Gateway proxy response with appropriate status code
func JSON(resp string) (*events.APIGatewayProxyResponse, error) {
	return &events.APIGatewayProxyResponse{StatusCode: http.StatusOK, Body: string(resp)}, nil
}
