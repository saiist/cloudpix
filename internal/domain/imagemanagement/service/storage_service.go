package service

import (
	"context"
	"time"
)

// StorageService は画像ストレージに関するドメインサービスを定義します
type StorageService interface {
	// StoreImage は Base64 エンコードされた画像データを保存します
	StoreImage(ctx context.Context, bucket, key, contentType, base64Data string) (string, error)

	// GenerateImageURL は画像へのプレサインドURLを生成します
	GenerateImageURL(ctx context.Context, bucket, key, contentType string, expiration time.Duration) (uploadURL string, downloadURL string, err error)

	// DeleteImage は画像を削除します
	DeleteImage(ctx context.Context, bucket, key string) error
}
