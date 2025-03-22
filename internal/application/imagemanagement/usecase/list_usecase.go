package usecase

import (
	"cloudpix/internal/application/imagemanagement/dto"
	"cloudpix/internal/domain/imagemanagement/repository"
	"cloudpix/internal/domain/imagemanagement/valueobject"
	"context"
)

// ListUsecase は画像一覧取得のユースケースを実装します
type ListUsecase struct {
	imageRepository repository.ImageRepository
}

// NewListUsecase は新しい一覧取得ユースケースを作成します
func NewListUsecase(imageRepository repository.ImageRepository) *ListUsecase {
	return &ListUsecase{
		imageRepository: imageRepository,
	}
}

// List はすべての画像を取得します
func (u *ListUsecase) List(ctx context.Context) (*dto.ListResponse, error) {
	options := repository.ImageQueryOptions{
		Limit: 100, // デフォルト上限
	}

	// リポジトリから画像を取得
	images, err := u.imageRepository.Find(ctx, options)
	if err != nil {
		return nil, err
	}

	// エンティティをDTOに変換
	imagesDTO := make([]dto.ImageMetadataDTO, len(images))
	for i, img := range images {
		imagesDTO[i] = dto.ImageMetadataDTO{
			ImageID:     img.ID,
			FileName:    img.FileName.String(),
			ContentType: img.ContentType.String(),
			Size:        img.Size.Value(),
			UploadDate:  img.UploadDate.String(),
			DownloadURL: img.DownloadURL,
		}
	}

	return &dto.ListResponse{
		Images: imagesDTO,
		Count:  len(imagesDTO),
	}, nil
}

// ListByDate は指定された日付の画像を取得します
func (u *ListUsecase) ListByDate(ctx context.Context, dateStr string) (*dto.ListResponse, error) {
	// 日付の値オブジェクトを作成
	date, err := valueobject.NewUploadDate(dateStr)
	if err != nil {
		return nil, err
	}

	// リポジトリから画像を取得
	images, err := u.imageRepository.FindByDate(ctx, date)
	if err != nil {
		return nil, err
	}

	// エンティティをDTOに変換
	imagesDTO := make([]dto.ImageMetadataDTO, len(images))
	for i, img := range images {
		imagesDTO[i] = dto.ImageMetadataDTO{
			ImageID:     img.ID,
			FileName:    img.FileName.String(),
			ContentType: img.ContentType.String(),
			Size:        img.Size.Value(),
			UploadDate:  img.UploadDate.String(),
			DownloadURL: img.DownloadURL,
		}
	}

	return &dto.ListResponse{
		Images: imagesDTO,
		Count:  len(imagesDTO),
	}, nil
}
