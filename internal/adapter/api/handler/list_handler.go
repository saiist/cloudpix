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

// ListHandler は画像一覧を処理するハンドラー
type ListHandler struct {
	listUsecase *usecase.ListUsecase
}

// NewListHandler は新しいListHandlerを作成します
func NewListHandler(listUsecase *usecase.ListUsecase) *ListHandler {
	return &ListHandler{
		listUsecase: listUsecase,
	}
}

// Handle はAPIリクエストを処理します
func (h *ListHandler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := logging.FromContext(ctx)
	logger.Info("Processing list request", map[string]interface{}{
		"method": request.HTTPMethod,
		"path":   request.Path,
	})

	// クエリパラメータからフィルターを取得
	date := request.QueryStringParameters["date"]

	var response *dto.ListResponse
	var err error

	if date != "" {
		logger.Info("Filtering by date", map[string]interface{}{
			"date": date,
		})
		response, err = h.listUsecase.ListByDate(ctx, date)
	} else {
		logger.Info("Listing all images", nil)
		response, err = h.listUsecase.List(ctx)
	}

	if err != nil {
		logger.Error(err, "Error listing images", nil)
		return h.errorResponse(http.StatusInternalServerError, "画像一覧の取得に失敗しました")
	}

	logger.Info("Images retrieved successfully", map[string]interface{}{
		"count": response.Count,
	})

	return h.jsonResponse(http.StatusOK, response)
}

// jsonResponse はJSON形式のレスポンスを作成します
func (h *ListHandler) jsonResponse(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	responseJSON, err := json.Marshal(body)
	if err != nil {
		return h.errorResponse(http.StatusInternalServerError, "レスポンス生成中にエラーが発生しました")
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
func (h *ListHandler) errorResponse(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	body, _ := json.Marshal(map[string]string{"error": message})
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}
