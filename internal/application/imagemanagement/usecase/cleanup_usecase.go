package usecase

import (
	"cloudpix/internal/domain/imagemanagement/repository"
	"cloudpix/internal/domain/imagemanagement/service"
	"cloudpix/internal/domain/shared/event/dispatcher"
	tagrepository "cloudpix/internal/domain/tagmanagement/repository"
	"cloudpix/internal/logging"
	"context"
	"time"
)

// CleanupUsecase は画像クリーンアップのユースケースを実装
type CleanupUsecase struct {
	imageRepository repository.ImageRepository
	tagRepository   tagrepository.TagRepository
	storageService  service.StorageService
	cleanupService  service.CleanupService
	eventDispatcher dispatcher.EventDispatcher
	retentionDays   int
	logger          logging.Logger
}

// NewCleanupUsecase は新しいクリーンアップユースケースを作成
func NewCleanupUsecase(
	imageRepository repository.ImageRepository,
	tagRepository tagrepository.TagRepository,
	storageService service.StorageService,
	cleanupService service.CleanupService,
	eventDispatcher dispatcher.EventDispatcher,
	retentionDays int,
	logger logging.Logger,
) *CleanupUsecase {
	return &CleanupUsecase{
		imageRepository: imageRepository,
		tagRepository:   tagRepository,
		storageService:  storageService,
		cleanupService:  cleanupService,
		eventDispatcher: eventDispatcher,
		retentionDays:   retentionDays,
		logger:          logger,
	}
}

// ProcessCleanup は古い画像のクリーンアップ処理を実行
func (u *CleanupUsecase) ProcessCleanup(ctx context.Context) error {
	u.logger.Info("Starting cleanup process", map[string]interface{}{
		"retentionDays": u.retentionDays,
	})

	// 基準日の計算（現在からretentionDays日前）
	cutoffDate := time.Now().AddDate(0, 0, -u.retentionDays)
	dateStr := cutoffDate.Format("2006-01-02")

	// 指定日付より古い画像をクエリ
	options := repository.ImageQueryOptions{
		UploadDateBefore: dateStr,
	}

	oldImages, err := u.imageRepository.Find(ctx, options)
	if err != nil {
		return err
	}

	u.logger.Info("Found old images to process", map[string]interface{}{
		"count": len(oldImages),
	})

	// 各画像を処理
	for _, image := range oldImages {
		err := u.cleanupService.ArchiveImage(ctx, image.ID)
		if err != nil {
			u.logger.Error(err, "Failed to archive image", map[string]interface{}{
				"imageId": image.ID,
			})
			continue
		}

		u.logger.Info("Successfully archived image", map[string]interface{}{
			"imageId": image.ID,
		})
	}

	return nil
}
