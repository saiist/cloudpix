package event

import (
	"cloudpix/internal/application/thumbnailmanagement/usecase"
	imagemanagement_evnet "cloudpix/internal/domain/imagemanagement/event"
	"cloudpix/internal/domain/shared/event/dispatcher"
	"cloudpix/internal/logging"
	"context"
	"fmt"
)

// ImageUploadedHandler は画像アップロードイベントを処理するハンドラー
type ImageUploadedHandler struct {
	thumbnailUsecase *usecase.ThumbnailGenerationUsecase
	logger           logging.Logger
}

// NewImageUploadedHandler は新しいイベントハンドラーを作成します
func NewImageUploadedHandler(thumbnailUsecase *usecase.ThumbnailGenerationUsecase, logger logging.Logger) *ImageUploadedHandler {
	return &ImageUploadedHandler{
		thumbnailUsecase: thumbnailUsecase,
		logger:           logger,
	}
}

// HandleEvent はイベントを処理します
func (h *ImageUploadedHandler) HandleEvent(ctx context.Context, event dispatcher.DomainEvent) error {
	// イベントを適切な型にキャスト
	uploadEvent, ok := event.(*imagemanagement_evnet.ImageUploadedEvent)
	if !ok {
		return fmt.Errorf("expected ImageUploadedEvent, got %T", event)
	}

	h.logger.Info("Processing image upload event", map[string]interface{}{
		"imageId":     uploadEvent.ImageID,
		"bucket":      uploadEvent.Bucket,
		"s3ObjectKey": uploadEvent.S3ObjectKey,
	})

	// サムネイルを生成
	result, err := h.thumbnailUsecase.ProcessImage(ctx, uploadEvent.Bucket, uploadEvent.S3ObjectKey)
	if err != nil {
		h.logger.Error(err, "Failed to generate thumbnail", map[string]interface{}{
			"imageId": uploadEvent.ImageID,
		})
		return fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	if !result.Success {
		h.logger.Warn("Thumbnail generation was not successful", map[string]interface{}{
			"imageId": uploadEvent.ImageID,
			"message": result.Message,
		})
		return nil
	}

	h.logger.Info("Thumbnail generated successfully", map[string]interface{}{
		"imageId":      uploadEvent.ImageID,
		"thumbnailUrl": result.ThumbnailURL,
		"dimensions":   fmt.Sprintf("%dx%d", result.Width, result.Height),
	})

	return nil
}

// EventType はこのハンドラーが処理するイベントタイプを返します
func (h *ImageUploadedHandler) EventType() string {
	return "image.uploaded"
}
