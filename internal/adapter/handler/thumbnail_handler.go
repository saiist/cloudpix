package handler

import (
	"cloudpix/internal/usecase"
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

type ThumbnailHandler struct {
	thumbnailUsecase *usecase.ThumbnailUsecase
}

func NewThumbnailHandler(thumbnailUsecase *usecase.ThumbnailUsecase) *ThumbnailHandler {
	return &ThumbnailHandler{
		thumbnailUsecase: thumbnailUsecase,
	}
}

// S3イベントを処理する
func (h *ThumbnailHandler) Handle(ctx context.Context, s3Event events.S3Event) error {
	// イベントの処理
	for _, record := range s3Event.Records {
		// S3バケットとオブジェクトキーを取得
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		log.Printf("Processing image from bucket: %s, key: %s", bucket, key)

		// サムネイル処理を実行
		err := h.thumbnailUsecase.ProcessImage(ctx, bucket, key)

		if err != nil {
			log.Printf("Error processing image: %s", err)
			// 次の画像の処理を続行
			continue
		}

		log.Printf("Successfully processed image: %s", key)
	}

	return nil
}
