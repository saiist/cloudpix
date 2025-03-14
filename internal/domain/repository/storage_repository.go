package repository

import (
	"cloudpix/internal/domain/model"
	"context"
	"time"
)

type StorageRepository interface {
	// イメージを取得する
	FetchImage(ctx context.Context, bucket, key string) (*model.ImageData, error)

	// 生成したサムネイルをS3にアップロードする
	UploadThumbnail(ctx context.Context, bucket, key string, data *model.ImageData) error

	// Base64エンコードされた画像データをS3にアップロードする
	UploadImage(ctx context.Context, bucket, key, contentType, base64Data string) (string, error)

	// アップロード用プレサインドURLを生成する
	GeneratePresignedURL(ctx context.Context, bucket, key, contentType string, expiration time.Duration) (uploadURL string, downloadURL string, err error)
}
