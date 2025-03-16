package handler

import (
	"cloudpix/internal/adapter/middleware"
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
	authMiddleware  middleware.AuthMiddleware
}

func NewListHandler(metadataUsecase *usecase.MetadataUsecase, authMiddleware middleware.AuthMiddleware) *ListHandler {
	return &ListHandler{
		metadataUsecase: metadataUsecase,
		authMiddleware:  authMiddleware,
	}
}

// Handle はAPIリクエストを処理する
func (h *ListHandler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// 認証ミドルウェアを適用
	handlerWithAuth := WithAuth(h.authMiddleware, h.handleList)
	return handlerWithAuth(ctx, request)
}

// handleList は認証なしの実際のハンドラー処理
func (h *ListHandler) handleList(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
		return h.errorResponse(500, fmt.Sprintf("Failed to retrieve tags: %s", err))
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
