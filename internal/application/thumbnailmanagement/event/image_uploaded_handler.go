package event

import (
	upload_event "cloudpix/internal/domain/imagemanagement/event"
	"cloudpix/internal/domain/shared/event/dispatcher"
	"context"
	"fmt"
)

// ThumbnailGenerator はサムネイル生成サービスのインターフェース
type ThumbnailGenerator interface {
	GenerateThumbnail(ctx context.Context, bucket, key string) error
}

// ImageUploadedHandler は画像アップロードイベントを処理するハンドラー
type ImageUploadedHandler struct {
	thumbnailGenerator ThumbnailGenerator
}

// NewImageUploadedHandler は新しいイベントハンドラーを作成します
func NewImageUploadedHandler(thumbnailGenerator ThumbnailGenerator) *ImageUploadedHandler {
	return &ImageUploadedHandler{
		thumbnailGenerator: thumbnailGenerator,
	}
}

// HandleEvent はイベントを処理します
func (h *ImageUploadedHandler) HandleEvent(ctx context.Context, event dispatcher.DomainEvent) error {
	// イベントを適切な型にキャスト
	uploadEvent, ok := event.(*upload_event.ImageUploadedEvent)
	if !ok {
		return fmt.Errorf("expected ImageUploadedEvent, got %T", event)
	}

	// サムネイルを生成
	err := h.thumbnailGenerator.GenerateThumbnail(ctx, uploadEvent.Bucket, uploadEvent.S3ObjectKey)
	if err != nil {
		return fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	return nil
}

// EventType はこのハンドラーが処理するイベントタイプを返します
func (h *ImageUploadedHandler) EventType() string {
	return "image.uploaded"
}
