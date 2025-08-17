package app

import (
	"context"
	openapi "user-backend/docs/gen/go"
	infra "user-backend/infra"
)

type OpinionService struct {
	openapi.OpinionAPIService
	db *infra.DynamoDBClient
}

func NewOpinionService(db *infra.DynamoDBClient) *OpinionService {
	return &OpinionService{db: db}
}

// PostUserOpinions - 意見投稿API
func (s *OpinionService) PostUserOpinions(ctx context.Context, opinion openapi.OpinionRequest) (openapi.ImplResponse, error) {
	// DynamoDBに保存する処理
	_, err := s.db.SaveOpinion(
		ctx,
		opinion.MailAddress,
		opinion.Coordinate.Latitude,
		opinion.Coordinate.Longitude,
		opinion.Opinion,
	)
	if err != nil {
		return openapi.Response(500, nil), err
	}

	// 正常時は201を返す
	return openapi.Response(201, nil), nil
}

// PostUserComments - コメント投稿API
func (s *OpinionService) PostUserComments(ctx context.Context, opinionId string, commentRequest openapi.CommentRequest) (openapi.ImplResponse, error) {
	// DynamoDBにコメントを保存する処理
	_, err := s.db.SaveComment(
		ctx,
		opinionId,
		commentRequest.MailAddress,
		commentRequest.Comment,
	)
	if err != nil {
		return openapi.Response(500, nil), err
	}

	// 正常時は201を返す
	return openapi.Response(201, nil), nil
}

// GetUserOpinions - ユーザー意見取得API
func (s *OpinionService) GetUserOpinions(ctx context.Context) (openapi.ImplResponse, error) {
	opinions, err := s.db.GetOpinions(ctx) // DynamoDBから意見を取得する処理
	if err != nil {
		return openapi.Response(500, nil), err
	}

	return openapi.Response(200, opinions), nil // 正常時は200と意見を返す
}
