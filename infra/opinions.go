package infra

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time" // 追加

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
const commentsTableName = "comments"

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

// SaveComment - コメントをDynamoDBに保存するメソッド
func (db *DynamoDBClient) SaveComment(ctx context.Context, opinionId string, mailAddress string, comment string) (string, error) {
	commentId := uuid.New().String()

	item := map[string]types.AttributeValue{
		"opinionid":       &types.AttributeValueMemberS{Value: opinionId},
		"commentId":       &types.AttributeValueMemberS{Value: commentId},
		"mailAddress":     &types.AttributeValueMemberS{Value: mailAddress},
		"comment":         &types.AttributeValueMemberS{Value: comment},
		"createdDateTime": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
	}

	_, err := db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(commentsTableName),
		Item:      item,
	})
	if err != nil {
		return "", err
	}

	return commentId, nil
}

// GetOpinions - ユーザーの意見を取得するメソッド
func (db *DynamoDBClient) GetOpinions(ctx context.Context) ([]OpinionItem, error) {
	var allOpinions []OpinionItem
	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:         aws.String(opinionsTableName),
			ExclusiveStartKey: lastEvaluatedKey,
		}

		result, err := db.Client.Scan(ctx, input)
		if err != nil {
			log.Printf("DynamoDB Scan failed: %v", err)
			return nil, err
		}

		// DynamoDBから返された各項目をOpinion構造体にデコード
		var opinionsPage []OpinionItem
		for _, item := range result.Items {
			var opinion OpinionItem
			opinion.ID = item["id"].(*types.AttributeValueMemberS).Value
			opinion.MailAddress = item["mailAddress"].(*types.AttributeValueMemberS).Value
			opinion.Latitude, _ = strconv.ParseFloat(item["latitude"].(*types.AttributeValueMemberN).Value, 64)
			opinion.Longitude, _ = strconv.ParseFloat(item["longitude"].(*types.AttributeValueMemberN).Value, 64)
			opinion.Opinion = item["opinion"].(*types.AttributeValueMemberS).Value
			if err != nil {
				log.Printf("Failed to unmarshal DynamoDB item: %v", err)
				return nil, err
			}
			opinionsPage = append(opinionsPage, opinion)
		}
		if err != nil {
			log.Printf("Failed to unmarshal DynamoDB items: %v", err)
			return nil, err
		}

		// 取得した項目をすべてOpinionsリストに追加
		allOpinions = append(allOpinions, opinionsPage...)

		// LastEvaluatedKeyがnilでない場合、再度取得
		if result.LastEvaluatedKey == nil {
			break
		}
		lastEvaluatedKey = result.LastEvaluatedKey
	}

	return allOpinions, nil
}
