package repository

import (
	"cloudpix/internal/domain/model"
	"context"
	"time"
)

type UploadRepository interface {
	// 直接アップロードする（Base64エンコードされたデータを使用）
	UploadImage(ctx context.Context, imageID string, request *model.UploadRequest) (string, error)

	// プレサインドURLを生成する
	GeneratePresignedURL(ctx context.Context, imageID string, request *model.UploadRequest, expiration time.Duration) (string, string, error)

	// メタデータをDynamoDBに保存する
	SaveMetadata(ctx context.Context, metadata *model.UploadMetadata) error
}
