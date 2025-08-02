package main

import (
    "context"
    "log"
    "net/http"
    "user-backend/app"
	"strings"

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

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// CloudWatch Logsに出力
	log.Printf("Request: %s %s", req.RequestContext.HTTP.Method, req.RawPath)
	// API Gateway v2のパスプロキシから実際のパスを抽出
	// /dev/user/test -> /test
	originalPath := req.RawPath
	path := req.RawPath
	if strings.HasPrefix(path, "/dev/user") {
		path = strings.TrimPrefix(path, "/dev/user")
	}
	if path == "" {
		path = "/"
	}
	
	method := req.RequestContext.HTTP.Method
	
	key := routeKey{
		Path:   path,
		Method: method,
	}
	if routeHandler, ok := routes[key]; ok {
		// API Gateway v1のリクエストに変換してハンドラーを呼び出し
		v1Req := events.APIGatewayProxyRequest{
			HTTPMethod: method,
			Path:       originalPath,
			Body:       req.Body,
			Headers:    req.Headers,
		}
		v1Resp, err := routeHandler(ctx, v1Req)
		if err != nil {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "Internal Server Error",
			}, err
		}
		
		// API Gateway v2のレスポンスに変換
		return events.APIGatewayV2HTTPResponse{
			StatusCode: v1Resp.StatusCode,
			Body:       v1Resp.Body,
			Headers:    v1Resp.Headers,
		}, nil
	}
	
	// デバッグ情報を含む404レスポンス
	debugInfo := "Not Found - Original Path: " + originalPath + ", Processed Path: " + path + ", Method: " + method
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNotFound,
		Headers: map[string]string{
            "Content-Type": "application/json",
        },
		Body: debugInfo,
	}, nil
}

func main() {
    lambda.Start(handler)
}
