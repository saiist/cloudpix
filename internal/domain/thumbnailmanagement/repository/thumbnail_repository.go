package repository

import (
	"cloudpix/internal/domain/thumbnailmanagement/entity"
	"context"
)

// ThumbnailRepository はサムネイル情報の永続化を担当するインターフェース
type ThumbnailRepository interface {
	// Save はサムネイル情報を保存します
	Save(ctx context.Context, thumbnail *entity.Thumbnail) error

	// FindByImageID は指定された画像IDのサムネイルを取得します
	FindByImageID(ctx context.Context, imageID string) (*entity.Thumbnail, error)

	// Delete はサムネイルを削除します
	Delete(ctx context.Context, imageID string) error

	// UpdateMetadata はサムネイルのメタデータを更新します
	UpdateMetadata(ctx context.Context, imageID string, thumbnailURL string, width, height int) error
}
