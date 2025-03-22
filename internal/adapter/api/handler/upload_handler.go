package handler

import (
	"cloudpix/internal/application/imagemanagement/dto"
	"cloudpix/internal/application/imagemanagement/usecase"
	"cloudpix/internal/logging"
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// UploadHandler は画像アップロードを処理するハンドラー
type UploadHandler struct {
	uploadUsecase *usecase.UploadUsecase
}

// NewUploadHandler は新しいアップロードハンドラーを作成します
func NewUploadHandler(uploadUsecase *usecase.UploadUsecase) *UploadHandler {
	return &UploadHandler{
		uploadUsecase: uploadUsecase,
	}
}

// Handle はAPI Gatewayからのリクエストを処理します
func (h *UploadHandler) Handle(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := logging.FromContext(ctx)

	// リクエストボディを解析
	var request dto.UploadRequest
	if err := json.Unmarshal([]byte(event.Body), &request); err != nil {
		logger.Error(err, "Failed to unmarshal request", nil)
		return h.createErrorResponse(http.StatusBadRequest, "不正なリクエスト形式")
	}

	// アップロード処理実行
	response, err := h.uploadUsecase.ProcessUpload(ctx, &request)
	if err != nil {
		logger.Error(err, "Upload error", nil)
		return h.createErrorResponse(http.StatusInternalServerError, "アップロード処理中にエラーが発生しました")
	}

	// 成功時のログ記録
	logger.Info("Upload successful", map[string]interface{}{
		"imageId": response.ImageID,
	})

	// レスポンスのJSON変換
	responseJSON, err := json.Marshal(response)
	if err != nil {
		logger.Error(err, "Failed to marshal response", nil)
		return h.createErrorResponse(http.StatusInternalServerError, "レスポンス生成中にエラーが発生しました")
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

// createErrorResponse はエラーレスポンスを作成します
func (h *UploadHandler) createErrorResponse(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	body, _ := json.Marshal(map[string]string{"message": message})

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}
