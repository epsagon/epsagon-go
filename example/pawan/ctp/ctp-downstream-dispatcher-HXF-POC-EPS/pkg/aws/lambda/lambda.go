package lambda

import (
	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
)

// Start wraps the standard aws lambda handler function and returns the error message
// within the response.
func Start(fn func(events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error)) {
	lambda.Start(func(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {

		//check if lambda event payload is empty
		if len(req.Body) == 0 {
			err := errors.New("request empty")
			return nil, err
		}

		resp, err := fn(req)
		if err != nil {
			log.Info().Msgf("received error from handler:%+v", err)
			return Error(err)
		}
		return resp, nil
	})
}
