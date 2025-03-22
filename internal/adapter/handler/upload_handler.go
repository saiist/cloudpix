package handler

import (
	"cloudpix/internal/domain/model"
	"cloudpix/internal/logging"
	"cloudpix/internal/usecase"
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// UploadHandler は画像アップロードを処理するハンドラー
type UploadHandler struct {
	uploadUsecase *usecase.UploadUsecase
}

// NewUploadHandler は新しいUploadHandlerを作成する
func NewUploadHandler(uploadUsecase *usecase.UploadUsecase) *UploadHandler {
	return &UploadHandler{
		uploadUsecase: uploadUsecase,
	}
}

// Handle はAPI Gatewayからのリクエストを処理する
func (h *UploadHandler) Handle(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := logging.FromContext(ctx)

	// リクエストボディを解析
	var request model.UploadRequest
	if err := json.Unmarshal([]byte(event.Body), &request); err != nil {
		logger.Error(err, "Failed to unmarshal request", nil)
		return h.createErrorResponse(http.StatusBadRequest, "不正なリクエスト形式")
	}

	// アップロード前のコンテキスト情報
	uploadCtx := logging.UploadContext{
		FileName:    request.FileName,
		ContentType: request.ContentType,
	}

	logger.Info("Processing upload", map[string]interface{}{
		"upload": uploadCtx,
	})

	// アップロード処理実行
	response, err := h.uploadUsecase.ProcessUpload(ctx, &request)
	if err != nil {
		logger.Error(err, "Upload error", map[string]interface{}{
			"upload": uploadCtx,
		})
		return h.createErrorResponse(http.StatusInternalServerError, "アップロード処理中にエラーが発生しました")
	}

	// 成功時のコンテキスト情報
	uploadCtx.ImageID = response.ImageID

	logger.Info("Upload successful", map[string]interface{}{
		"upload":  uploadCtx,
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

// createErrorResponse はエラーレスポンスを作成する
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
