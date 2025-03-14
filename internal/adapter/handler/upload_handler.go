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

type UploadHandler struct {
	uploadUsecase *usecase.UploadUsecase
}

func NewUploadHandler(uploadUsecase *usecase.UploadUsecase) *UploadHandler {
	return &UploadHandler{
		uploadUsecase: uploadUsecase,
	}
}

// Handle はAPIリクエストを処理します
func (h *UploadHandler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received request: %s", request.Body)

	// リクエストボディをパース
	var uploadReq model.UploadRequest
	err := json.Unmarshal([]byte(request.Body), &uploadReq)
	if err != nil {
		log.Printf("Error parsing request: %s", err)
		return h.errorResponse(400, fmt.Sprintf("Invalid request format: %s", err))
	}

	// アップロード処理を実行
	response, err := h.uploadUsecase.ProcessUpload(ctx, &uploadReq)
	if err != nil {
		log.Printf("Error processing upload: %s", err)
		return h.errorResponse(500, fmt.Sprintf("Failed to process upload: %s", err))
	}

	// レスポンスをJSON形式で返す
	return h.jsonResponse(200, response)
}

// jsonResponse はJSON形式のレスポンスを作成します
func (h *UploadHandler) jsonResponse(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
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

// errorResponse はエラーレスポンスを作成します
func (h *UploadHandler) errorResponse(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: fmt.Sprintf(`{"error":"%s"}`, message),
	}, nil
}
