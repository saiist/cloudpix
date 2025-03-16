package handler

import (
	"cloudpix/internal/adapter/middleware"
	"cloudpix/internal/domain/model"
	"cloudpix/internal/usecase"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

type UploadHandler struct {
	uploadUsecase  *usecase.UploadUsecase
	authMiddleware middleware.AuthMiddleware
}

func NewUploadHandler(uploadUsecase *usecase.UploadUsecase, authMiddleware middleware.AuthMiddleware) *UploadHandler {
	return &UploadHandler{
		uploadUsecase:  uploadUsecase,
		authMiddleware: authMiddleware,
	}
}

// Handle はAPIリクエストを処理します
func (h *UploadHandler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// 認証ミドルウェアを適用
	handlerWithAuth := WithAuth(h.authMiddleware, h.handleUpload)
	return handlerWithAuth(ctx, request)
}

// handleUpload は認証なしの実際のハンドラー処理
func (h *UploadHandler) handleUpload(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received request: %s", event.Body)

	// リクエストボディを解析
	var request model.UploadRequest
	if err := json.Unmarshal([]byte(event.Body), &request); err != nil {
		log.Printf("Failed to unmarshal request: %v", err)
		return h.authMiddleware.CreateErrorResponse(http.StatusBadRequest, "不正なリクエスト形式"), nil
	}

	// アップロード処理実行
	response, err := h.uploadUsecase.ProcessUpload(ctx, &request)
	if err != nil {
		log.Printf("Upload error: %v", err)
		return h.authMiddleware.CreateErrorResponse(http.StatusInternalServerError, "アップロード処理中にエラーが発生しました"), nil
	}

	// レスポンスのJSON変換
	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return h.authMiddleware.CreateErrorResponse(http.StatusInternalServerError, "レスポンス生成中にエラーが発生しました"), nil
	}

	// 成功レスポンスを返す
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseJSON),
	}, nil
}
