package repository

import (
	"cloudpix/internal/domain/model"
	"context"
)

type UploadMetadataRepository interface {
	// メタデータをDynamoDBに保存する
	SaveMetadata(ctx context.Context, metadata *model.UploadMetadata) error
}
