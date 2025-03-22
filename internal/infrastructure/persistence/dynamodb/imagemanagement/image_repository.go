package imagemanagement

import (
	"cloudpix/internal/domain/imagemanagement/aggregate"
	"cloudpix/internal/domain/imagemanagement/entity"
	"cloudpix/internal/domain/imagemanagement/repository"
	"cloudpix/internal/domain/imagemanagement/valueobject"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// DynamoDBImageItem はDynamoDBのイメージアイテム表現
type DynamoDBImageItem struct {
	ImageID         string   `json:"ImageID"`
	FileName        string   `json:"FileName"`
	ContentType     string   `json:"ContentType"`
	Size            int      `json:"Size"`
	UploadDate      string   `json:"UploadDate"`
	S3ObjectKey     string   `json:"S3ObjectKey"`
	DownloadURL     string   `json:"DownloadURL"`
	ThumbnailURL    string   `json:"ThumbnailURL,omitempty"`
	ThumbnailWidth  int      `json:"ThumbnailWidth,omitempty"`
	ThumbnailHeight int      `json:"ThumbnailHeight,omitempty"`
	Tags            []string `json:"Tags,omitempty"`
	CreatedAt       string   `json:"CreatedAt"`
	ModifiedAt      string   `json:"ModifiedAt"`
	HasThumbnail    bool     `json:"HasThumbnail"`
}

// DynamoDBImageRepository はDynamoDBを使用した画像リポジトリの実装
type DynamoDBImageRepository struct {
	client            *dynamodb.DynamoDB
	metadataTableName string
}

// NewDynamoDBImageRepository は新しいDynamoDBリポジトリを作成します
func NewDynamoDBImageRepository(client *dynamodb.DynamoDB, metadataTableName string) repository.ImageRepository {
	return &DynamoDBImageRepository{
		client:            client,
		metadataTableName: metadataTableName,
	}
}

// FindByID は指定されたIDの画像集約を取得します
func (r *DynamoDBImageRepository) FindByID(ctx context.Context, id string) (*aggregate.ImageAggregate, error) {
	// DynamoDBから画像データを取得
	result, err := r.client.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.metadataTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ImageID": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get image from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("image not found: %s", id)
	}

	// DynamoDBアイテムをマッピング
	var item DynamoDBImageItem
	if err := dynamodbattribute.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DynamoDB item: %w", err)
	}

	// 値オブジェクトの作成
	fileName, _ := valueobject.NewFileName(item.FileName)
	contentType, _ := valueobject.NewContentType(item.ContentType)
	size, _ := valueobject.NewImageSize(item.Size)
	uploadDate, _ := valueobject.NewUploadDate(item.UploadDate)

	// エンティティの作成
	image := &entity.Image{
		ID:           item.ImageID,
		FileName:     fileName,
		ContentType:  contentType,
		Size:         size,
		UploadDate:   uploadDate,
		S3ObjectKey:  item.S3ObjectKey,
		DownloadURL:  item.DownloadURL,
		HasThumbnail: item.HasThumbnail,
	}

	// 集約の作成
	imageAggregate := aggregate.NewImageAggregate(image)
	imageAggregate.ThumbnailURL = item.ThumbnailURL
	imageAggregate.ThumbnailWidth = item.ThumbnailWidth
	imageAggregate.ThumbnailHeight = item.ThumbnailHeight
	imageAggregate.Tags = item.Tags

	return imageAggregate, nil
}

// FindByDate は指定された日付の画像を検索します
func (r *DynamoDBImageRepository) FindByDate(ctx context.Context, date valueobject.UploadDate) ([]*entity.Image, error) {
	// フィルター式の作成
	filt := expression.Name("UploadDate").Equal(expression.Value(date.String()))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	// DynamoDBからデータを取得
	input := &dynamodb.ScanInput{
		TableName:                 aws.String(r.metadataTableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	}

	result, err := r.client.ScanWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan DynamoDB: %w", err)
	}

	// 結果をエンティティに変換
	images := make([]*entity.Image, 0)
	for _, item := range result.Items {
		var dbItem DynamoDBImageItem
		if err := dynamodbattribute.UnmarshalMap(item, &dbItem); err != nil {
			continue
		}

		// 値オブジェクトの作成
		fileName, _ := valueobject.NewFileName(dbItem.FileName)
		contentType, _ := valueobject.NewContentType(dbItem.ContentType)
		size, _ := valueobject.NewImageSize(dbItem.Size)
		uploadDate, _ := valueobject.NewUploadDate(dbItem.UploadDate)

		image := &entity.Image{
			ID:           dbItem.ImageID,
			FileName:     fileName,
			ContentType:  contentType,
			Size:         size,
			UploadDate:   uploadDate,
			S3ObjectKey:  dbItem.S3ObjectKey,
			DownloadURL:  dbItem.DownloadURL,
			HasThumbnail: dbItem.HasThumbnail,
		}

		images = append(images, image)
	}

	return images, nil
}

// Find は条件に一致する画像を検索します
func (r *DynamoDBImageRepository) Find(ctx context.Context, options repository.ImageQueryOptions) ([]*entity.Image, error) {
	// 検索条件の構築
	var filterBuilder expression.ConditionBuilder
	var filterSet bool

	// デフォルトではアーカイブされていない画像のみを返す
	filterBuilder = expression.Name("ImageStatus").AttributeNotExists().
		Or(expression.Name("ImageStatus").NotEqual(expression.Value("ARCHIVED")))
	filterSet = true

	// 日付フィルター
	if options.UploadDateBefore != "" {
		dateFilter := expression.Name("UploadDate").LessThanEqual(expression.Value(options.UploadDateBefore))
		if filterSet {
			filterBuilder = filterBuilder.And(dateFilter)
		} else {
			filterBuilder = dateFilter
			filterSet = true
		}
	}

	// スキャン入力の作成
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(r.metadataTableName),
	}

	// フィルター式があれば適用
	if filterSet {
		expr, err := expression.NewBuilder().WithFilter(filterBuilder).Build()
		if err != nil {
			return nil, fmt.Errorf("failed to build expression: %w", err)
		}
		scanInput.ExpressionAttributeNames = expr.Names()
		scanInput.ExpressionAttributeValues = expr.Values()
		scanInput.FilterExpression = expr.Filter()
	}

	// 結果数の制限（指定されていれば）
	if options.Limit > 0 {
		scanInput.Limit = aws.Int64(int64(options.Limit))
	}

	// DynamoDBからデータを取得
	result, err := r.client.ScanWithContext(ctx, scanInput)
	if err != nil {
		return nil, fmt.Errorf("failed to scan DynamoDB: %w", err)
	}

	// 結果をエンティティに変換
	images := make([]*entity.Image, 0)
	for _, item := range result.Items {
		var dbItem DynamoDBImageItem
		if err := dynamodbattribute.UnmarshalMap(item, &dbItem); err != nil {
			continue
		}

		// 値オブジェクトの作成
		fileName, _ := valueobject.NewFileName(dbItem.FileName)
		contentType, _ := valueobject.NewContentType(dbItem.ContentType)
		size, _ := valueobject.NewImageSize(dbItem.Size)
		uploadDate, _ := valueobject.NewUploadDate(dbItem.UploadDate)

		image := &entity.Image{
			ID:           dbItem.ImageID,
			FileName:     fileName,
			ContentType:  contentType,
			Size:         size,
			UploadDate:   uploadDate,
			S3ObjectKey:  dbItem.S3ObjectKey,
			DownloadURL:  dbItem.DownloadURL,
			HasThumbnail: dbItem.HasThumbnail,
		}

		images = append(images, image)
	}

	return images, nil
}

// Save は画像集約を保存します
func (r *DynamoDBImageRepository) Save(ctx context.Context, imageAggregate *aggregate.ImageAggregate) error {
	// 集約から必要なデータを取得
	image := imageAggregate.Image

	// DynamoDBアイテムを作成
	item := DynamoDBImageItem{
		ImageID:         image.ID,
		FileName:        image.FileName.String(),
		ContentType:     image.ContentType.String(),
		Size:            image.Size.Value(),
		UploadDate:      image.UploadDate.String(),
		S3ObjectKey:     image.S3ObjectKey,
		DownloadURL:     image.DownloadURL,
		ThumbnailURL:    imageAggregate.ThumbnailURL,
		ThumbnailWidth:  imageAggregate.ThumbnailWidth,
		ThumbnailHeight: imageAggregate.ThumbnailHeight,
		Tags:            imageAggregate.Tags,
		CreatedAt:       image.CreatedAt.Format(time.RFC3339),
		ModifiedAt:      image.ModifiedAt.Format(time.RFC3339),
		HasThumbnail:    image.HasThumbnail,
	}

	// DynamoDBのアイテム形式に変換
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
		return fmt.Errorf("failed to save item to DynamoDB: %w", err)
	}

	return nil
}

// Delete は画像集約を削除します
func (r *DynamoDBImageRepository) Delete(ctx context.Context, id string) error {
	// DynamoDBから画像を削除
	_, err := r.client.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.metadataTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ImageID": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete image from DynamoDB: %w", err)
	}

	return nil
}

// Exists は画像が存在するかどうかを確認します
func (r *DynamoDBImageRepository) Exists(ctx context.Context, id string) (bool, error) {
	// DynamoDBから画像データを取得（最小限の属性のみ）
	result, err := r.client.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.metadataTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ImageID": {
				S: aws.String(id),
			},
		},
		ProjectionExpression: aws.String("ImageID"),
	})
	if err != nil {
		return false, fmt.Errorf("failed to check image existence: %w", err)
	}

	// アイテムが存在する場合はtrueを返す
	return result.Item != nil, nil
}
