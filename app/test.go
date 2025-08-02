package app

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func HandleCreate(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "Hello, World!",
	}, nil
}