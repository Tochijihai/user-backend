package infra

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type OpinionItem struct {
	ID          string
	MailAddress string
	Latitude    float64
	Longitude   float64
	Opinion     string
}

const opinionsTableName = "opinions"

func (db *DynamoDBClient) SaveOpinion(ctx context.Context, mailAddress string, latitude, longitude float64, opinion string) (string, error) {
	id := uuid.New().String()

	item := map[string]types.AttributeValue{
		"id":          &types.AttributeValueMemberS{Value: id},
		"mailAddress": &types.AttributeValueMemberS{Value: mailAddress},
		"latitude":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", latitude)},
		"longitude":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", longitude)},
		"opinion":     &types.AttributeValueMemberS{Value: opinion},
	}

	_, err := db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(opinionsTableName),
		Item:      item,
	})
	if err != nil {
		return "", err
	}

	return id, nil
}
