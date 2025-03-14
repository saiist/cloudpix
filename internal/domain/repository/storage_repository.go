package repository

import (
	"cloudpix/internal/domain/model"
	"context"
)

// ストレージリポジトリのインターフェース
type StorageRepository interface {
	// S3からイメージを取得する
	FetchImage(ctx context.Context, bucket, key string) (*model.ImageData, error)

	// 生成したサムネイルをS3にアップロードする
	UploadThumbnail(ctx context.Context, bucket, key string, data *model.ImageData) error
}
