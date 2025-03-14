package persistence

import (
	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/repository"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type DynamoDBThumbnailRepository struct {
	dynamoDBClient    *dynamodb.DynamoDB
	metadataTableName string
}

func NewDynamoDBThumbnailRepository(
	dynamoDBClient *dynamodb.DynamoDB,
	metadataTableName string,
) repository.ThumbnailRepository {
	return &DynamoDBThumbnailRepository{
		dynamoDBClient:    dynamoDBClient,
		metadataTableName: metadataTableName,
	}
}

// サムネイル情報をDynamoDBに保存
func (r *DynamoDBThumbnailRepository) UpdateMetadata(ctx context.Context, info *model.ThumbnailInfo) error {
	// DynamoDBのメタデータを更新
	_, err := r.dynamoDBClient.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.metadataTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ImageID": {
				S: aws.String(info.ImageID),
			},
		},
		UpdateExpression: aws.String("SET thumbnailKey = :tk, thumbnailUrl = :tu, thumbnailWidth = :tw, thumbnailHeight = :th"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":tk": {
				S: aws.String(info.ThumbnailKey),
			},
			":tu": {
				S: aws.String(info.ThumbnailURL),
			},
			":tw": {
				N: aws.String(fmt.Sprintf("%d", info.Width)),
			},
			":th": {
				N: aws.String(fmt.Sprintf("%d", info.Height)),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to update DynamoDB item: %w", err)
	}

	return nil
}
