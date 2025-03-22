package service

import "context"

// CleanupService はイメージの自動クリーンアップを担当するドメインサービス
type CleanupService interface {
	// CleanupOldImages は指定された日数より古い画像を処理する
	CleanupOldImages(ctx context.Context, retentionDays int) error

	// ArchiveImage は画像をアーカイブする
	ArchiveImage(ctx context.Context, imageID string) error

	// DeleteImage は画像とすべての関連データを完全に削除する
	DeleteImage(ctx context.Context, imageID string) error
}
