package repository

import "context"

type TagRepository interface {
	// タグ一覧を取得
	ListTags(ctx context.Context) ([]string, error)

	// 特定の画像のタグを取得
	GetImageTags(ctx context.Context, imageID string) ([]string, error)

	// タグを追加
	AddTags(ctx context.Context, imageID string, tags []string) (int, error)

	// 特定のタグを削除
	RemoveTags(ctx context.Context, imageID string, tags []string) (int, error)

	// 画像のすべてのタグを削除
	RemoveAllTags(ctx context.Context, imageID string) (int, error)

	// 画像の存在を確認
	VerifyImageExists(ctx context.Context, imageID string) (bool, error)
}
