package usecase

import (
	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/repository"
	"context"
)

type MetadataUsecase struct {
	metadataRepo repository.MetadataRepository
}

func NewMetadataUsecase(metadataRepo repository.MetadataRepository) *MetadataUsecase {
	return &MetadataUsecase{
		metadataRepo: metadataRepo,
	}
}

func (u *MetadataUsecase) List(ctx context.Context) (*model.ListResponse, error) {
	metadataList, err := u.metadataRepo.Find(ctx)
	if err != nil {
		return nil, err
	}

	return &model.ListResponse{
		Images: metadataList,
		Count:  len(metadataList),
	}, nil
}

func (u *MetadataUsecase) ListByDate(ctx context.Context, date string) (*model.ListResponse, error) {
	metadataList, err := u.metadataRepo.FindByDate(ctx, date)
	if err != nil {
		return nil, err
	}

	return &model.ListResponse{
		Images: metadataList,
		Count:  len(metadataList),
	}, nil
}
