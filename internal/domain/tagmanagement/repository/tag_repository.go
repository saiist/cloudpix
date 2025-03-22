package repository

import (
	"cloudpix/internal/domain/tagmanagement/entity"
	"cloudpix/internal/domain/tagmanagement/valueobject"
	"context"
)

// TagRepository はタグ情報の永続化を担当するインターフェース
type TagRepository interface {
	// FindAllTags は全てのユニークなタグを取得します
	FindAllTags(ctx context.Context) ([]valueobject.Tag, error)

	// FindTaggedImage は指定された画像IDのタグ情報を取得します
	FindTaggedImage(ctx context.Context, imageID string) (*entity.TaggedImage, error)

	// FindImagesByTag は指定されたタグを持つ画像IDのリストを取得します
	FindImagesByTag(ctx context.Context, tag valueobject.Tag) ([]string, error)

	// Save はタグ付き画像情報を保存します
	Save(ctx context.Context, taggedImage *entity.TaggedImage) error

	// Delete はタグ付き画像情報を削除します
	Delete(ctx context.Context, imageID string) error

	// ImageExists は画像が存在するか確認します
	ImageExists(ctx context.Context, imageID string) (bool, error)
}
