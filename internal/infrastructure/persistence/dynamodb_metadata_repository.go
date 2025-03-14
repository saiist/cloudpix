package persistence

import (
	"cloudpix/internal/domain/model"
	"cloudpix/internal/domain/repository"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type DynamoDBMetadataRepository struct {
	client            *dynamodb.DynamoDB
	metadataTableName string
}

func NewDynamoDBMetadataRepository(client *dynamodb.DynamoDB, metadataTableName string) repository.MetadataRepository {
	return &DynamoDBMetadataRepository{
		client:            client,
		metadataTableName: metadataTableName,
	}
}

// Find は全てのイメージメタデータを取得します
func (r *DynamoDBMetadataRepository) Find(ctx context.Context) ([]model.ImageMetadata, error) {
	// フィルターなしで全件取得するためのスキャン入力を作成
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.metadataTableName),
	}

	// 取得と変換を行う
	return r.scanOperation(ctx, input)
}

// FindByDate は指定された日付のイメージメタデータを取得します
func (r *DynamoDBMetadataRepository) FindByDate(ctx context.Context, date string) ([]model.ImageMetadata, error) {

	// フィルタ式を作成
	filt := expression.Name("UploadDate").Equal(expression.Value(date))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	// フィルター付きスキャン入力を作成
	input := &dynamodb.ScanInput{
		TableName:                 aws.String(r.metadataTableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	}

	// 取得と変換を行う
	return r.scanOperation(ctx, input)
}

// scanOperation はスキャン操作を実行し、結果をモデルに変換する共通処理です
func (r *DynamoDBMetadataRepository) scanOperation(ctx context.Context, input *dynamodb.ScanInput) ([]model.ImageMetadata, error) {
	// スキャンを実行
	result, err := r.client.ScanWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan DynamoDB: %w", err)
	}

	if len(result.Items) == 0 {
		return []model.ImageMetadata{}, nil
	}

	// 結果をアンマーシャル
	var images []model.ImageMetadata
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &images); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DynamoDB results: %w", err)
	}

	return images, nil
}
