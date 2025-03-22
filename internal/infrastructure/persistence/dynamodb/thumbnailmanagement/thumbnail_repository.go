package thumbnailmanagement

import (
	"cloudpix/internal/domain/thumbnailmanagement/entity"
	"cloudpix/internal/domain/thumbnailmanagement/repository"
	"cloudpix/internal/domain/thumbnailmanagement/valueobject"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// DynamoDBThumbnailItem はDynamoDBのサムネイルアイテム表現
type DynamoDBThumbnailItem struct {
	ImageID      string `json:"ImageID"`
	ThumbnailKey string `json:"ThumbnailKey"`
	ThumbnailURL string `json:"ThumbnailURL"`
	Width        int    `json:"Width"`
	Height       int    `json:"Height"`
	OriginalKey  string `json:"OriginalKey"`
	ContentType  string `json:"ContentType"`
	CreatedAt    string `json:"CreatedAt"`
}

// DynamoDBThumbnailRepository はDynamoDBを使用したサムネイルリポジトリの実装
type DynamoDBThumbnailRepository struct {
	client            *dynamodb.DynamoDB
	metadataTableName string
}

// NewDynamoDBThumbnailRepository は新しいDynamoDBリポジトリを作成します
func NewDynamoDBThumbnailRepository(client *dynamodb.DynamoDB, metadataTableName string) repository.ThumbnailRepository {
	return &DynamoDBThumbnailRepository{
		client:            client,
		metadataTableName: metadataTableName,
	}
}

// Save はサムネイル情報を保存します
func (r *DynamoDBThumbnailRepository) Save(ctx context.Context, thumbnail *entity.Thumbnail) error {
	// DynamoDBアイテムを作成
	item := DynamoDBThumbnailItem{
		ImageID:      thumbnail.ImageID,
		ThumbnailKey: thumbnail.ThumbnailKey,
		ThumbnailURL: thumbnail.ThumbnailURL,
		Width:        thumbnail.GetWidth(),
		Height:       thumbnail.GetHeight(),
		OriginalKey:  thumbnail.OriginalKey,
		ContentType:  thumbnail.ContentType,
		CreatedAt:    thumbnail.CreatedAt.Format(time.RFC3339),
	}

	// マーシャル
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal DynamoDB item: %w", err)
	}

	// DynamoDBに保存
	_, err = r.client.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.metadataTableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to save thumbnail to DynamoDB: %w", err)
	}

	return nil
}

// FindByImageID は指定された画像IDのサムネイルを取得します
func (r *DynamoDBThumbnailRepository) FindByImageID(ctx context.Context, imageID string) (*entity.Thumbnail, error) {
	// DynamoDBから取得
	result, err := r.client.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.metadataTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ImageID": {
				S: aws.String(imageID),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get thumbnail from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("thumbnail not found for image ID: %s", imageID)
	}

	// アンマーシャル
	var item DynamoDBThumbnailItem
	if err := dynamodbattribute.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DynamoDB item: %w", err)
	}

	// 値オブジェクトの作成
	dimensions, _ := valueobject.NewDimensions(item.Width, item.Height)

	// サムネイルエンティティの作成
	createdAt, _ := time.Parse(time.RFC3339, item.CreatedAt)
	thumbnail := &entity.Thumbnail{
		ImageID:      item.ImageID,
		ThumbnailKey: item.ThumbnailKey,
		ThumbnailURL: item.ThumbnailURL,
		Dimensions:   dimensions,
		OriginalKey:  item.OriginalKey,
		ContentType:  item.ContentType,
		CreatedAt:    createdAt,
	}

	return thumbnail, nil
}

// Delete はサムネイルを削除します
func (r *DynamoDBThumbnailRepository) Delete(ctx context.Context, imageID string) error {
	_, err := r.client.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.metadataTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ImageID": {
				S: aws.String(imageID),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete thumbnail from DynamoDB: %w", err)
	}

	return nil
}

// UpdateMetadata はサムネイルのメタデータを更新します
func (r *DynamoDBThumbnailRepository) UpdateMetadata(ctx context.Context, imageID string, thumbnailURL string, width, height int) error {
	_, err := r.client.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.metadataTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ImageID": {
				S: aws.String(imageID),
			},
		},
		UpdateExpression: aws.String("SET ThumbnailURL = :tu, Width = :w, Height = :h"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":tu": {
				S: aws.String(thumbnailURL),
			},
			":w": {
				N: aws.String(fmt.Sprintf("%d", width)),
			},
			":h": {
				N: aws.String(fmt.Sprintf("%d", height)),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update thumbnail metadata: %w", err)
	}

	return nil
}
