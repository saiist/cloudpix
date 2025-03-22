package usecase

import (
	"cloudpix/internal/application/thumbnailmanagement/dto"
	"cloudpix/internal/domain/shared/event/dispatcher"
	"cloudpix/internal/domain/thumbnailmanagement/entity"
	"cloudpix/internal/domain/thumbnailmanagement/event"
	"cloudpix/internal/domain/thumbnailmanagement/repository"
	"cloudpix/internal/domain/thumbnailmanagement/service"
	"cloudpix/internal/domain/thumbnailmanagement/valueobject"
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// ThumbnailGenerationUsecase はサムネイル生成ユースケース
type ThumbnailGenerationUsecase struct {
	thumbnailRepo     repository.ThumbnailRepository
	storageService    service.StorageService
	processingService service.ImageProcessingService
	eventDispatcher   dispatcher.EventDispatcher
	thumbnailSize     int
	awsRegion         string
}

// NewThumbnailGenerationUsecase は新しいサムネイル生成ユースケースを作成します
func NewThumbnailGenerationUsecase(
	thumbnailRepo repository.ThumbnailRepository,
	storageService service.StorageService,
	processingService service.ImageProcessingService,
	eventDispatcher dispatcher.EventDispatcher,
	thumbnailSize int,
	awsRegion string,
) *ThumbnailGenerationUsecase {
	return &ThumbnailGenerationUsecase{
		thumbnailRepo:     thumbnailRepo,
		storageService:    storageService,
		processingService: processingService,
		eventDispatcher:   eventDispatcher,
		thumbnailSize:     thumbnailSize,
		awsRegion:         awsRegion,
	}
}

// ProcessImage はS3に保存された画像からサムネイルを生成します
func (u *ThumbnailGenerationUsecase) ProcessImage(ctx context.Context, bucket, key string) (*dto.ThumbnailGenerationResponseDTO, error) {
	// アップロードディレクトリ以外は処理しない
	if !strings.HasPrefix(key, "uploads/") {
		return &dto.ThumbnailGenerationResponseDTO{
			Success: false,
			Message: "Non-upload directory images are not processed",
		}, nil
	}

	// S3から画像を取得
	imageData, err := u.storageService.FetchImage(ctx, bucket, key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}

	// 画像をデコード
	width, height, err := u.processingService.DecodeImage(imageData)
	if err != nil {
		return &dto.ThumbnailGenerationResponseDTO{
			Success: false,
			Message: fmt.Sprintf("Unsupported image format: %v", err),
		}, nil
	}

	// TODO: サムネイルを生成
	if _, err := valueobject.NewDimensions(width, height); err != nil {
		return nil, fmt.Errorf("invalid image dimensions: %w", err)
	}

	// サムネイルを生成
	thumbnailData, thumbnailDimensions, err := u.processingService.GenerateThumbnail(imageData, u.thumbnailSize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	// サムネイルのS3キーを生成
	filename := filepath.Base(key)
	thumbnailKey := fmt.Sprintf("thumbnails/%s", filename)

	// サムネイルをS3にアップロード
	err = u.storageService.UploadThumbnail(ctx, bucket, thumbnailKey, thumbnailData)
	if err != nil {
		return nil, fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	// サムネイルのURLを生成
	thumbnailURL := u.storageService.GetObjectURL(bucket, thumbnailKey)

	// 画像IDを抽出
	imageID, err := u.processingService.ExtractImageID(filename)
	if err != nil {
		return &dto.ThumbnailGenerationResponseDTO{
			Success: false,
			Message: "Could not extract image ID from filename",
		}, nil
	}

	// サムネイルエンティティを作成
	thumbnail := entity.NewThumbnail(
		imageID,
		thumbnailKey,
		thumbnailURL,
		thumbnailDimensions,
		key,
		thumbnailData.ContentType,
	)

	// サムネイル情報をリポジトリに保存
	err = u.thumbnailRepo.Save(ctx, thumbnail)
	if err != nil {
		return nil, fmt.Errorf("failed to save thumbnail metadata: %w", err)
	}

	// イベントを発行
	thumbnailEvent := event.NewThumbnailGeneratedEvent(thumbnail)
	u.eventDispatcher.Dispatch(ctx, thumbnailEvent)

	// レスポンスを作成
	return &dto.ThumbnailGenerationResponseDTO{
		Success:      true,
		ImageID:      imageID,
		ThumbnailURL: thumbnailURL,
		Width:        thumbnail.GetWidth(),
		Height:       thumbnail.GetHeight(),
		Message:      "Thumbnail generated successfully",
	}, nil
}
