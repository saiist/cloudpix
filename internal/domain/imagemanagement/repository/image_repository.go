package repository

import (
	"cloudpix/internal/domain/imagemanagement/aggregate"
	"cloudpix/internal/domain/imagemanagement/entity"
	"cloudpix/internal/domain/imagemanagement/valueobject"
	"context"
)

// ImageQueryOptions は画像検索のオプションを表す構造体
type ImageQueryOptions struct {
	UploadDate string
	Tags       []string
	Limit      int
	Offset     int
}

// ImageRepository は画像集約の永続化を担当するインターフェース
type ImageRepository interface {
	// FindByID は指定されたIDの画像集約を取得します
	FindByID(ctx context.Context, id string) (*aggregate.ImageAggregate, error)

	// FindByDate は指定された日付の画像を検索します
	FindByDate(ctx context.Context, date valueobject.UploadDate) ([]*entity.Image, error)

	// Find は条件に一致する画像を検索します
	Find(ctx context.Context, options ImageQueryOptions) ([]*entity.Image, error)

	// Save は画像集約を保存します
	Save(ctx context.Context, imageAggregate *aggregate.ImageAggregate) error

	// Delete は画像集約を削除します
	Delete(ctx context.Context, id string) error

	// Exists は画像が存在するかどうかを確認します
	Exists(ctx context.Context, id string) (bool, error)
}
