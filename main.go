package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	openapi "user-backend/docs/gen/go"
)

// ResponseWriterラッパー
type proxyResponseWriter struct {
	headers    http.Header
	body       *bytes.Buffer
	statusCode int
}

// API Gatewayのステージプレフィックス（例: /dev）を除去する関数
func normalizePath(rawPath string) string {
	const stagePrefix = "/dev" // 実際のステージ名に書き換えてください

	if strings.HasPrefix(rawPath, stagePrefix) {
		return strings.TrimPrefix(rawPath, stagePrefix)
	}
	return rawPath
}

func newProxyResponseWriter() *proxyResponseWriter {
	return &proxyResponseWriter{
		headers:    make(http.Header),
		body:       bytes.NewBuffer([]byte{}),
		statusCode: http.StatusOK,
	}
}

func (r *proxyResponseWriter) Header() http.Header {
	return r.headers
}

func (r *proxyResponseWriter) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *proxyResponseWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

// LambdaイベントのHTTP API v2リクエストを *http.Request に変換
func proxyEventToHTTPRequest(req events.APIGatewayV2HTTPRequest) (*http.Request, error) {
	// Pathを正規化（ステージを除去）
	normalizedPath := normalizePath(req.RawPath)

	// URLパース（Path + クエリパラメータ）
	uri := normalizedPath
	if len(req.RawQueryString) > 0 {
		uri += "?" + req.RawQueryString
	}
	parsedURL, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, err
	}

	// Body
	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = strings.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(req.RequestContext.HTTP.Method, parsedURL.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	// Headerをセット
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	return httpReq, nil
}

// Lambdaハンドラー
func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Incoming request: %s %s", req.RequestContext.HTTP.Method, req.RawPath)

	// OpenAPIで生成されたrouterを作成
	opinionAPIService := openapi.NewOpinionAPIService()
	opinionAPIController := openapi.NewOpinionAPIController(opinionAPIService)
	router := openapi.NewRouter(opinionAPIController)

	// Lambdaイベントをhttp.Requestに変換
	httpReq, err := proxyEventToHTTPRequest(req)
	if err != nil {
		log.Printf("Failed to convert event to http.Request: %v", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: 500}, err
	}

	log.Printf("Converted path: %s", httpReq.URL.Path)
	log.Printf("Method: %s", httpReq.Method)

	// ResponseWriterを作成
	respWriter := newProxyResponseWriter()

	// routerを呼び出す（http.HandlerのServeHTTP）
	router.ServeHTTP(respWriter, httpReq)

	log.Printf("Response status: %d", respWriter.statusCode)
	log.Printf("Response body: %s", respWriter.body.String())

	// ここでResponseWriterの内容をAPI Gateway v2レスポンスに変換
	resp := events.APIGatewayV2HTTPResponse{
		StatusCode:      respWriter.statusCode,
		Headers:         map[string]string{},
		Body:            respWriter.body.String(),
		IsBase64Encoded: false,
	}

	// Headersコピー
	for k, v := range respWriter.headers {
		resp.Headers[k] = strings.Join(v, ",")
	}

	return resp, nil
}


func main() {
	lambda.Start(handler)
}
