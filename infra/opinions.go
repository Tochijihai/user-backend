package infra

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type Coordinate struct {
	Latitude  float64 `json:"Latitude"`
	Longitude float64 `json:"Longitude"`
}

type OpinionItem struct {
	ID          string
	MailAddress string
	Coordinate  Coordinate `json:"Coordinate"`
	Opinion     string
}

type CommentItem struct {
	ID              string // OpinionID
	CommentID       string // コメントID
	MailAddress     string
	Comment         string
	CreatedDateTime time.Time
}

type Reaction struct {
	IsReactioned bool `json:"IsReactioned"`
}

type ReactionInfo struct {
	IsReactioned  bool  `json:"IsReactioned"`
	ReactionCount int32 `json:"ReactionCount"`
}

const opinionsTableName = "opinions"
const commentsTableName = "comments"
const reactionsTableName = "reactions"

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
			opinion.Coordinate.Latitude, _ = strconv.ParseFloat(item["latitude"].(*types.AttributeValueMemberN).Value, 64)
			opinion.Coordinate.Longitude, _ = strconv.ParseFloat(item["longitude"].(*types.AttributeValueMemberN).Value, 64)
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

// SaveComment - コメントをDynamoDBに保存するメソッド
func (db *DynamoDBClient) SaveComment(ctx context.Context, opinionId string, mailAddress string, comment string) (string, error) {
	commentId := uuid.New().String()

	item := map[string]types.AttributeValue{
		"opinionId":       &types.AttributeValueMemberS{Value: opinionId},
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

// GetComment - OpinionIDに紐づくコメントをDynamoDBから取得するメソッド
func (db *DynamoDBClient) GetComment(ctx context.Context, opinionId string) ([]CommentItem, error) {
	var comments []CommentItem
	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		input := &dynamodb.QueryInput{
			TableName:              aws.String(commentsTableName),
			IndexName:              aws.String("opinionId-createdDateTime-index"), // GSI名を指定
			KeyConditionExpression: aws.String("opinionId = :opinionId"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":opinionId": &types.AttributeValueMemberS{Value: opinionId},
			},
			ExclusiveStartKey: lastEvaluatedKey,
		}

		result, err := db.Client.Query(ctx, input)
		if err != nil {
			log.Printf("DynamoDB Query failed: %v", err)
			return nil, err
		}

		// DynamoDBから返された各項目をComment構造体にデコード
		for _, item := range result.Items {
			var comment CommentItem
			comment.ID = item["opinionId"].(*types.AttributeValueMemberS).Value
			comment.CommentID = item["commentId"].(*types.AttributeValueMemberS).Value
			comment.MailAddress = item["mailAddress"].(*types.AttributeValueMemberS).Value
			comment.Comment = item["comment"].(*types.AttributeValueMemberS).Value
			comment.CreatedDateTime, _ = time.Parse(time.RFC3339, item["createdDateTime"].(*types.AttributeValueMemberS).Value)

			comments = append(comments, comment)
		}

		// LastEvaluatedKeyがnilでない場合、再度取得
		if result.LastEvaluatedKey == nil {
			break
		}
		lastEvaluatedKey = result.LastEvaluatedKey
	}

	return comments, nil
}

// SaveReaction - リアクションをDynamoDBに保存(更新)するメソッド
func (db *DynamoDBClient) SaveReaction(ctx context.Context, opinionId string, mailAddress string, isReactioned bool) (Reaction, error) {
	item := map[string]types.AttributeValue{
		"opinionId":    &types.AttributeValueMemberS{Value: opinionId},
		"mailAddress":  &types.AttributeValueMemberS{Value: mailAddress},
		"isReactioned": &types.AttributeValueMemberBOOL{Value: isReactioned},
	}

	_, err := db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(reactionsTableName),
		Item:      item,
	})
	if err != nil {
		return Reaction{}, err
	}
	return Reaction{IsReactioned: isReactioned}, nil
}

// SaveReaction - リアクション情報をDynamoDBから取得するメソッド
func (db *DynamoDBClient) GetReactionInfo(ctx context.Context, opinionId string, mailAddress string) (ReactionInfo, error) {
	// IsReactionedの取得
	isReactionedInput := &dynamodb.GetItemInput{
		TableName: aws.String(reactionsTableName),
		Key: map[string]types.AttributeValue{
			"opinionId":   &types.AttributeValueMemberS{Value: opinionId},
			"mailAddress": &types.AttributeValueMemberS{Value: mailAddress},
		},
	}

	result, err := db.Client.GetItem(ctx, isReactionedInput)
	if err != nil {
		return ReactionInfo{}, err
	}

	var isReactioned bool
	if result.Item != nil {
		isReactioned = result.Item["isReactioned"].(*types.AttributeValueMemberBOOL).Value
	}

	// ReactionCountの取得
	reactionCountInput := &dynamodb.QueryInput{
		TableName:              aws.String(reactionsTableName),
		KeyConditionExpression: aws.String("opinionId = :opinionId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":opinionId": &types.AttributeValueMemberS{Value: opinionId},
		},
	}

	reactionCountResult, err := db.Client.Query(ctx, reactionCountInput)
	if err != nil {
		return ReactionInfo{}, err
	}

	var reactionCount int32
	for _, item := range reactionCountResult.Items {
		if item["isReactioned"].(*types.AttributeValueMemberBOOL).Value {
			reactionCount++
		}
	}

	return ReactionInfo{
		IsReactioned:  isReactioned,
		ReactionCount: reactionCount,
	}, nil
}
