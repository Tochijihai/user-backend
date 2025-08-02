package main

import (
    "context"
    "log"
    "net/http"
    "user-backend/app"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)

type routeKey struct {
    Path   string
    Method string
}

var routes = map[routeKey]func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error){
    {Path: "/test", Method: "POST"}: app.HandleCreate,
    // {Path: "/test", Method: "GET"}:  app.HandleGet,
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    // CloudWatch Logsに出力
    log.Printf("Request: %s %s", req.HTTPMethod, req.Path)
    
    key := routeKey{
        Path:   req.Path,
        Method: req.HTTPMethod,
    }
    
    if routeHandler, ok := routes[key]; ok {
        return routeHandler(ctx, req)
    }
    
    // API Gatewayが期待するレスポンス形式
    return events.APIGatewayProxyResponse{
        StatusCode: http.StatusNotFound,
        Headers: map[string]string{
            "Content-Type": "application/json",
        },
        Body: `{"error":"Not Found"}`,
    }, nil
}

func main() {
    lambda.Start(handler)
}