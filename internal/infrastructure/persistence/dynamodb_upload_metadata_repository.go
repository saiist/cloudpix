package persistence

import (
	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/repository"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type DynamoDBUploadMetadataRepository struct {
	dynamoDBClient    *dynamodb.DynamoDB
	metadataTableName string
}

func NewDynamoDBUploadMetadataRepository(
	dynamoDBClient *dynamodb.DynamoDB,
	metadataTableName string,
) repository.UploadMetadataRepository {
	return &DynamoDBUploadMetadataRepository{
		dynamoDBClient:    dynamoDBClient,
		metadataTableName: metadataTableName,
	}
}

// SaveMetadata はアップロードされた画像のメタデータをDynamoDBに保存します
func (r *DynamoDBUploadMetadataRepository) SaveMetadata(ctx context.Context, metadata *model.UploadMetadata) error {
	// DynamoDBのアイテム形式に変換
	item, err := dynamodbattribute.MarshalMap(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal DynamoDB item: %w", err)
	}

	// DynamoDBにメタデータを保存
	_, err = r.dynamoDBClient.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.metadataTableName),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to save metadata to DynamoDB: %w", err)
	}

	return nil
}
