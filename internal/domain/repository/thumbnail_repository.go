package repository

import (
	"cloudpix/internal/domain/model"
	"context"
)

// サムネイルリポジトリのインターフェース
type ThumbnailRepository interface {
	// サムネイル情報をDynamoDBのメタデータに保存する
	UpdateMetadata(ctx context.Context, info *model.ThumbnailInfo) error
}
