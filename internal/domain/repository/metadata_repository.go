package repository

import (
	"cloudpix/internal/domain/model"
	"context"
)

type MetadataRepository interface {
	Find(ctx context.Context) ([]model.ImageMetadata, error)
	FindByDate(ctx context.Context, date string) ([]model.ImageMetadata, error)
}
