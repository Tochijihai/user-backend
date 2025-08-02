package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"user-backend/app"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type routeKey struct {
	Path   string
	Method string
}

// ルーティング定義
var routes = map[routeKey]func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error){
	{Path: "/test", Method: "POST"}: app.HandleCreate,
}

// パスからAPI Gatewayのステージやパスプレフィックスを取り除く
func normalizePath(rawPath string) string {
	if after, ok := strings.CutPrefix(rawPath, "/dev/user"); ok {
		return after
	}
	return rawPath
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	method := req.RequestContext.HTTP.Method
	originalPath := req.RawPath
	path := normalizePath(originalPath)
	if path == "" {
		path = "/"
	}

	log.Printf("[INFO] Method: %s | Path: %s -> %s", method, originalPath, path)

	key := routeKey{
		Path:   path,
		Method: method,
	}

	if routeHandler, ok := routes[key]; ok {
		return routeHandler(ctx, req)
	}

	log.Printf("[WARN] No handler found for %s %s", method, path)

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNotFound,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"message":"Not Found"}`,
	}, nil
}

func main() {
	lambda.Start(handler)
}
