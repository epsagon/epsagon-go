package lambda

import (
	"net/http"
)

//APIGatewayError lambda api gateway error interface
type APIGatewayError interface {
	StatusCode() int
}

// ResponseError is an error type other than http errors
type ResponseError struct {
	Message string
}

func (re ResponseError) Error() string {
	return re.Message
}

//StatusCode returns the status code
func (ResponseError) StatusCode() int {
	return http.StatusInternalServerError
}

//BadRequestError http bad request error
type BadRequestError struct {
	Message string
}

func (e BadRequestError) Error() string {
	if e.Message == "" {
		return "bad request"
	}
	return e.Message
}

//StatusCode returns the status code
func (BadRequestError) StatusCode() int {
	return http.StatusBadRequest
}

//NotFoundError for 404 errors
type NotFoundError struct {
	Message string
}

func (e NotFoundError) Error() string {
	return e.Message
}

//StatusCode returns the status code
func (NotFoundError) StatusCode() int {
	return http.StatusNotFound
}

//UnauthorizedError 503 errors
type UnauthorizedError struct {
	Message string
}

func (e UnauthorizedError) Error() string {
	return e.Message
}

//StatusCode returns the status code
func (UnauthorizedError) StatusCode() int {
	return http.StatusUnauthorized
}
