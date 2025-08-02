package app

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func HandleCreate(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Body:       "Hello, World!",
	}, nil
}