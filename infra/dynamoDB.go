package infra

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBClient struct {
	Client *dynamodb.Client
}

// ConnectDynamoDBService creates a DynamoDB client
func ConnectDynamoDBService() *DynamoDBClient {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// クライアント作成
	svc := dynamodb.NewFromConfig(cfg)

	return &DynamoDBClient{
		Client: svc,
	}
}
