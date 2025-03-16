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

type TagHandler struct {
	tagUsecase     *usecase.TagUsecase
	authMiddleware middleware.AuthMiddleware
}

func NewTagHandler(tagUsecase *usecase.TagUsecase, authMiddleware middleware.AuthMiddleware) *TagHandler {
	return &TagHandler{
		tagUsecase:     tagUsecase,
		authMiddleware: authMiddleware,
	}
}

// Handle はAPIリクエストを処理する
func (h *TagHandler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// 認証ミドルウェアを適用
	handlerWithAuth := WithAuth(h.authMiddleware, h.handleTag)
	return handlerWithAuth(ctx, request)
}

// handleTag は認証なしの実際のハンドラー処理
func (h *TagHandler) handleTag(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Processing request: %s %s", request.HTTPMethod, request.Path)

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
	return events.APIGatewayProxyResponse{
		StatusCode: 404,
		Body:       `{"error":"Not Found"}`,
	}, nil
}

// listTags はすべてのタグのリストを取得する
func (h *TagHandler) listTags(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response, err := h.tagUsecase.ListTags(ctx)
	if err != nil {
		log.Printf("Error listing tags: %s", err)
		return h.errorResponse(500, fmt.Sprintf("Failed to retrieve tags: %s", err))
	}

	return h.jsonResponse(200, response)
}

// getImageTags は特定の画像のタグを取得する
func (h *TagHandler) getImageTags(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	imageID := request.PathParameters["imageId"]

	response, err := h.tagUsecase.GetImageTags(ctx, imageID)
	if err != nil {
		log.Printf("Error getting image tags: %s", err)
		return h.errorResponse(500, fmt.Sprintf("Failed to retrieve image tags: %s", err))
	}

	return h.jsonResponse(200, response)
}

// addTags はタグを追加する
func (h *TagHandler) addTags(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// リクエストボディをパース
	var tagRequest model.AddTagRequest
	err := json.Unmarshal([]byte(request.Body), &tagRequest)
	if err != nil {
		log.Printf("Error parsing request: %s", err)
		return h.errorResponse(400, fmt.Sprintf("Invalid request format: %s", err))
	}

	// タグを追加
	addedTags, err := h.tagUsecase.AddTags(ctx, &tagRequest)
	if err != nil {
		if err == usecase.ErrImageNotFound {
			return h.errorResponse(404, "Image not found")
		}
		log.Printf("Error adding tags: %s", err)
		return h.errorResponse(500, fmt.Sprintf("Failed to add tags: %s", err))
	}

	// レスポンスを作成
	return h.jsonResponse(200, map[string]interface{}{
		"message": fmt.Sprintf("Added %d tags to image %s", addedTags, tagRequest.ImageID),
	})
}

// removeTags はタグを削除する
func (h *TagHandler) removeTags(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// パスパラメータから画像IDを取得
	imageID := request.PathParameters["imageId"]

	// ボディが空の場合はすべてのタグを削除
	if request.Body == "" {
		removedTags, err := h.tagUsecase.RemoveAllTags(ctx, imageID)
		if err != nil {
			log.Printf("Error removing all tags: %s", err)
			return h.errorResponse(500, fmt.Sprintf("Failed to remove tags: %s", err))
		}

		return h.jsonResponse(200, map[string]interface{}{
			"message": fmt.Sprintf("Removed all %d tags from image %s", removedTags, imageID),
		})
	}

	// リクエストボディをパース
	var tagRequest model.RemoveTagRequest
	err := json.Unmarshal([]byte(request.Body), &tagRequest)
	if err != nil {
		return h.errorResponse(400, fmt.Sprintf("Invalid request format: %s", err))
	}

	// タグを削除
	removedTags, err := h.tagUsecase.RemoveTags(ctx, &tagRequest)
	if err != nil {
		log.Printf("Error removing tags: %s", err)
		return h.errorResponse(500, fmt.Sprintf("Failed to remove tags: %s", err))
	}

	return h.jsonResponse(200, map[string]interface{}{
		"message": fmt.Sprintf("Removed %d tags from image %s", removedTags, imageID),
	})
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
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: fmt.Sprintf(`{"error":"%s"}`, message),
	}, nil
}
