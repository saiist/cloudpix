package usecase

import (
	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/repository"
	"cloudpix/internal/domain/service"
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// サムネイル作成ユースケース
type ThumbnailUsecase struct {
	thumbnailRepo repository.ThumbnailRepository
	storageRepo   repository.StorageRepository
	imageService  service.ImageService
	awsRegion     string
	thumbnailSize int
}

// 新しいサムネイルユースケースを作成
func NewThumbnailUsecase(
	thumbnailRepo repository.ThumbnailRepository,
	storageRepo repository.StorageRepository,
	imageService service.ImageService,
	awsRegion string,
	thumbnailSize int,
) *ThumbnailUsecase {
	return &ThumbnailUsecase{
		thumbnailRepo: thumbnailRepo,
		storageRepo:   storageRepo,
		imageService:  imageService,
		awsRegion:     awsRegion,
		thumbnailSize: thumbnailSize,
	}
}

// S3イベントからサムネイルを生成・保存する
func (u *ThumbnailUsecase) ProcessImage(ctx context.Context, bucket, key string) error {
	// アップロードディレクトリ以外は処理しない
	if !strings.HasPrefix(key, "uploads/") {
		return nil
	}

	// S3から画像を取得
	imageData, err := u.storageRepo.FetchImage(ctx, bucket, key)
	if err != nil {
		return fmt.Errorf("failed to fetch image: %w", err)
	}

	// 画像をデコード
	_, _, err = u.imageService.DecodeImage(imageData)
	if err != nil {
		// サポートされていないフォーマットは無視
		if _, ok := err.(*service.UnsupportedFormatError); ok {
			return nil
		}
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// サムネイルを生成
	thumbnailData, width, height, err := u.imageService.GenerateThumbnail(imageData, u.thumbnailSize)
	if err != nil {
		return fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	// サムネイルのS3キーを生成
	filename := filepath.Base(key)
	thumbnailKey := fmt.Sprintf("thumbnails/%s", filename)

	// サムネイルをS3にアップロード
	err = u.storageRepo.UploadThumbnail(ctx, bucket, thumbnailKey, thumbnailData)
	if err != nil {
		return fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	// 画像IDの抽出
	imageID, err := u.imageService.ExtractImageID(filename)
	if err != nil {
		// IDが抽出できない場合は処理をスキップ
		return nil
	}

	// サムネイル情報を作成
	thumbnailInfo := &model.ThumbnailInfo{
		ImageID:      imageID,
		ThumbnailKey: thumbnailKey,
		ThumbnailURL: fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, u.awsRegion, thumbnailKey),
		Width:        width,
		Height:       height,
		OriginalKey:  key,
		ContentType:  thumbnailData.ContentType,
	}

	// メタデータを更新
	err = u.thumbnailRepo.UpdateMetadata(ctx, thumbnailInfo)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}
