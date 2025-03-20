package handler

import (
	"cloudpix/internal/domain/model"
	"cloudpix/internal/usecase"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

type ListHandler struct {
	metadataUsecase *usecase.MetadataUsecase
}

func NewListHandler(metadataUsecase *usecase.MetadataUsecase) *ListHandler {
	return &ListHandler{
		metadataUsecase: metadataUsecase,
	}
}

// Handle はAPIリクエストを処理する（ミドルウェア適用済み）
func (h *ListHandler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Processing request: %s %s", request.HTTPMethod, request.Path)

	// クエリパラメータからフィルターを取得
	date := request.QueryStringParameters["date"]

	var (
		response *model.ListResponse
		err      error
	)

	if date != "" {
		response, err = h.metadataUsecase.ListByDate(ctx, date)
	} else {
		response, err = h.metadataUsecase.List(ctx)
	}

	if err != nil {
		log.Printf("Error listing: %s", err)
		return h.errorResponse(500, fmt.Sprintf("Failed to retrieve images: %s", err))
	}

	// 成功レスポンスを返す
	return h.jsonResponse(200, response)
}

// jsonResponse はJSON形式のレスポンスを作成する
func (h *ListHandler) jsonResponse(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	responseJSON, err := json.Marshal(body)
	if err != nil {
		return h.errorResponse(500, "Internal Server Error")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseJSON),
	}, nil
}

// errorResponse はエラーレスポンスを作成する
func (h *ListHandler) errorResponse(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: fmt.Sprintf(`{"error":"%s"}`, message),
	}, nil
}
