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
