package handler

import (
	"cloudpix/internal/application/tagmanagement/dto"
	"cloudpix/internal/application/tagmanagement/usecase"
	"cloudpix/internal/logging"
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-lambda-go/events"
)

// TagHandler はタグ管理のAPIハンドラー
type TagHandler struct {
	tagUsecase *usecase.TagUsecase
}

// NewTagHandler は新しいタグハンドラーを作成します
func NewTagHandler(tagUsecase *usecase.TagUsecase) *TagHandler {
	return &TagHandler{
		tagUsecase: tagUsecase,
	}
}

// Handle はAPI Gatewayからのリクエストを処理します
func (h *TagHandler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := logging.FromContext(ctx)
	logger.Info("Processing tags request", map[string]interface{}{
		"method": request.HTTPMethod,
		"path":   request.Path,
	})

	// パスとメソッドに基づいてルーティング
	if request.Resource == "/tags" {
		if request.HTTPMethod == "GET" {
			// タグの一覧を取得
			return h.listTags(ctx, request)
		} else if request.HTTPMethod == "POST" {
			// タグを追加
			return h.addTags(ctx, request)
		}
	} else if request.Resource == "/tags/{imageId}" {
		if request.HTTPMethod == "GET" {
			// 特定の画像のタグを取得
			return h.getImageTags(ctx, request)
		} else if request.HTTPMethod == "DELETE" {
			// タグを削除
			return h.removeTags(ctx, request)
		}
	}

	// 未対応のパス・メソッド
	return h.errorResponse(404, "Not Found")
}

// listTags はすべてのタグのリストを取得する
func (h *TagHandler) listTags(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := logging.FromContext(ctx)

	// タグの一覧を取得
	response, err := h.tagUsecase.ListAllTags(ctx)
	if err != nil {
		logger.Error(err, "Error listing tags", nil)
		return h.errorResponse(500, "タグの取得に失敗しました")
	}

	// レスポンスをJSON形式で返す
	return h.jsonResponse(200, response)
}

// getImageTags は特定の画像のタグを取得する
func (h *TagHandler) getImageTags(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := logging.FromContext(ctx)

	// パスパラメータから画像IDを取得
	imageID := request.PathParameters["imageId"]
	if imageID == "" {
		return h.errorResponse(400, "画像IDが指定されていません")
	}

	// 画像のタグを取得
	response, err := h.tagUsecase.GetImageTags(ctx, imageID)
	if err != nil {
		if errors.Is(err, usecase.ErrImageNotFound) {
			return h.errorResponse(404, "指定された画像が見つかりません")
		}
		logger.Error(err, "Error getting image tags", map[string]interface{}{
			"imageId": imageID,
		})
		return h.errorResponse(500, "タグの取得に失敗しました")
	}

	// レスポンスをJSON形式で返す
	return h.jsonResponse(200, response)
}

// addTags はタグを追加する
func (h *TagHandler) addTags(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := logging.FromContext(ctx)

	// リクエストボディをパース
	var tagRequest dto.AddTagRequestDTO
	if err := json.Unmarshal([]byte(request.Body), &tagRequest); err != nil {
		logger.Error(err, "Error parsing request body", nil)
		return h.errorResponse(400, "無効なリクエスト形式です")
	}

	// バリデーション
	if tagRequest.ImageID == "" {
		return h.errorResponse(400, "画像IDは必須です")
	}
	if len(tagRequest.Tags) == 0 {
		return h.errorResponse(400, "タグが指定されていません")
	}

	// タグを追加
	response, err := h.tagUsecase.AddTags(ctx, &tagRequest)
	if err != nil {
		if errors.Is(err, usecase.ErrImageNotFound) {
			return h.errorResponse(404, "指定された画像が見つかりません")
		}
		if errors.Is(err, usecase.ErrInvalidTag) {
			return h.errorResponse(400, "無効なタグ形式が含まれています")
		}
		logger.Error(err, "Error adding tags", map[string]interface{}{
			"imageId": tagRequest.ImageID,
			"tags":    tagRequest.Tags,
		})
		return h.errorResponse(500, "タグの追加に失敗しました")
	}

	// レスポンスをJSON形式で返す
	return h.jsonResponse(200, response)
}

// removeTags はタグを削除する
func (h *TagHandler) removeTags(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := logging.FromContext(ctx)

	// パスパラメータから画像IDを取得
	imageID := request.PathParameters["imageId"]
	if imageID == "" {
		return h.errorResponse(400, "画像IDが指定されていません")
	}

	// リクエストボディをパース
	var tagRequest dto.RemoveTagRequestDTO
	tagRequest.ImageID = imageID

	if request.Body != "" {
		if err := json.Unmarshal([]byte(request.Body), &tagRequest); err != nil {
			logger.Error(err, "Error parsing request body", nil)
			return h.errorResponse(400, "無効なリクエスト形式です")
		}
	}

	// タグを削除（ボディが空の場合はすべてのタグを削除）
	response, err := h.tagUsecase.RemoveTags(ctx, &tagRequest)
	if err != nil {
		if errors.Is(err, usecase.ErrImageNotFound) {
			return h.errorResponse(404, "指定された画像が見つかりません")
		}
		logger.Error(err, "Error removing tags", map[string]interface{}{
			"imageId": imageID,
			"tags":    tagRequest.Tags,
		})
		return h.errorResponse(500, "タグの削除に失敗しました")
	}

	// レスポンスをJSON形式で返す
	return h.jsonResponse(200, response)
}

// jsonResponse はJSON形式のレスポンスを作成する
func (h *TagHandler) jsonResponse(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
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
func (h *TagHandler) errorResponse(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	body, _ := json.Marshal(map[string]string{"error": message})
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}
