package service

import (
	"cloudpix/internal/domain/thumbnailmanagement/valueobject"
	"context"
)

// StorageService はサムネイル画像のストレージサービスインターフェース
type StorageService interface {
	// FetchImage はストレージから画像を取得します
	FetchImage(ctx context.Context, bucket, key string) (valueobject.ImageData, error)

	// UploadThumbnail はサムネイルをアップロードします
	UploadThumbnail(ctx context.Context, bucket, key string, data valueobject.ImageData) error

	// GetObjectURL はオブジェクトの公開URLを生成します
	GetObjectURL(bucket, key string) string

	// DeleteThumbnail はサムネイルを削除します
	DeleteThumbnail(ctx context.Context, bucket, key string) error
}
