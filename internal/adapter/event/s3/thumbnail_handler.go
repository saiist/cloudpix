package s3

import (
	"cloudpix/internal/application/thumbnailmanagement/usecase"
	"cloudpix/internal/logging"
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

// ThumbnailHandler はS3イベントを処理してサムネイルを生成するハンドラー
type ThumbnailHandler struct {
	thumbnailUsecase *usecase.ThumbnailGenerationUsecase
	logger           logging.Logger
}

// NewThumbnailHandler は新しいサムネイルハンドラーを作成します
func NewThumbnailHandler(thumbnailUsecase *usecase.ThumbnailGenerationUsecase, logger logging.Logger) *ThumbnailHandler {
	return &ThumbnailHandler{
		thumbnailUsecase: thumbnailUsecase,
		logger:           logger,
	}
}

// Handle はS3イベントを処理します
func (h *ThumbnailHandler) Handle(ctx context.Context, s3Event events.S3Event) error {
	// イベントの処理
	for _, record := range s3Event.Records {
		// S3バケットとオブジェクトキーを取得
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		h.logger.Info("Processing image from S3 event", map[string]interface{}{
			"bucket": bucket,
			"key":    key,
		})

		// サムネイル処理を実行
		result, err := h.thumbnailUsecase.ProcessImage(ctx, bucket, key)
		if err != nil {
			h.logger.Error(err, "Error processing image", map[string]interface{}{
				"bucket": bucket,
				"key":    key,
			})
			// 次の画像の処理を続行
			continue
		}

		if !result.Success {
			h.logger.Warn("Thumbnail generation was not successful", map[string]interface{}{
				"bucket":  bucket,
				"key":     key,
				"message": result.Message,
			})
			continue
		}

		h.logger.Info("Successfully processed image", map[string]interface{}{
			"imageId":      result.ImageID,
			"thumbnailUrl": result.ThumbnailURL,
			"dimensions":   fmt.Sprintf("%dx%d", result.Width, result.Height),
		})
	}

	return nil
}
