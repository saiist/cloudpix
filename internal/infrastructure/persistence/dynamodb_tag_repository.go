package persistence

import (
	"cloudpix/internal/domain/repository"
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type DynamoDBTagRepository struct {
	client            *dynamodb.DynamoDB
	tagsTableName     string
	metadataTableName string
}

func NewDynamoDBTagRepository(client *dynamodb.DynamoDB, tagsTableName, metadataTableName string) repository.TagRepository {
	return &DynamoDBTagRepository{
		client:            client,
		tagsTableName:     tagsTableName,
		metadataTableName: metadataTableName,
	}
}

// タグ一覧を取得
func (r *DynamoDBTagRepository) ListTags(ctx context.Context) ([]string, error) {

	// タグの一覧を取得（重複なし）
	result, err := r.client.ScanWithContext(ctx, &dynamodb.ScanInput{
		TableName:            aws.String(r.tagsTableName),
		ProjectionExpression: aws.String("TagName"),
	})

	if err != nil {
		return nil, err
	}

	// ユニークなタグを抽出
	uniqueTags := make(map[string]bool)
	for _, item := range result.Items {
		if tagName, ok := item["TagName"]; ok {
			uniqueTags[*tagName.S] = true
		}
	}

	// マップからスライスに変換
	tags := make([]string, 0, len(uniqueTags))
	for tag := range uniqueTags {
		tags = append(tags, tag)
	}

	// レスポンスを作成
	return tags, nil
}

// 特定の画像のタグを取得
func (r *DynamoDBTagRepository) GetImageTags(ctx context.Context, imageID string) ([]string, error) {

	// 指定された画像IDのタグを検索
	result, err := r.client.QueryWithContext(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tagsTableName),
		IndexName:              aws.String("ImageIDIndex"),
		KeyConditionExpression: aws.String("ImageID = :imageId"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":imageId": {
				S: aws.String(imageID),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	// タグのリストを抽出
	var tags []string
	for _, item := range result.Items {
		if tagName, ok := item["TagName"]; ok {
			tags = append(tags, *tagName.S)
		}
	}

	return tags, nil
}

// タグを追加
func (r *DynamoDBTagRepository) AddTags(ctx context.Context, imageID string, tags []string) (int, error) {

	// タグを追加
	addedTags := 0
	for _, tag := range tags {

		// タグをDynamoDBに追加
		_, err := r.client.PutItemWithContext(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(r.tagsTableName),
			Item: map[string]*dynamodb.AttributeValue{
				"TagName": {
					S: aws.String(tag),
				},
				"ImageID": {
					S: aws.String(imageID),
				},
			},
			// 同じタグが既に存在する場合は上書きしない
			ConditionExpression: aws.String("attribute_not_exists(TagName) AND attribute_not_exists(ImageID)"),
		})

		// 既存のタグエラーは無視（冪等性を確保）
		if err != nil && !strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			continue
		}

		addedTags++
	}

	return addedTags, nil
}

// 特定のタグを削除
func (r *DynamoDBTagRepository) RemoveTags(ctx context.Context, imageID string, tags []string) (int, error) {
	removedTags := 0

	for _, tag := range tags {
		// タグをDynamoDBから削除
		_, err := r.client.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String(r.tagsTableName),
			Key: map[string]*dynamodb.AttributeValue{
				"TagName": {
					S: aws.String(tag),
				},
				"ImageID": {
					S: aws.String(imageID),
				},
			},
		})

		if err != nil {
			continue
		}

		removedTags++
	}

	return removedTags, nil

}

// 画像のすべてのタグを削除
func (r *DynamoDBTagRepository) RemoveAllTags(ctx context.Context, imageID string) (int, error) {
	// 画像のすべてのタグを検索
	result, err := r.client.QueryWithContext(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tagsTableName),
		IndexName:              aws.String("ImageIDIndex"),
		KeyConditionExpression: aws.String("ImageID = :imageId"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":imageId": {
				S: aws.String(imageID),
			},
		},
	})

	if err != nil {
		return 0, err
	}

	// すべてのタグを削除
	removedTags := 0
	for _, item := range result.Items {
		tagName := item["TagName"].S

		_, err := r.client.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String(r.tagsTableName),
			Key: map[string]*dynamodb.AttributeValue{
				"TagName": {
					S: tagName,
				},
				"ImageID": {
					S: aws.String(imageID),
				},
			},
		})

		if err != nil {
			continue
		}

		removedTags++
	}

	return removedTags, nil
}

// 画像の存在を確認
func (r *DynamoDBTagRepository) VerifyImageExists(ctx context.Context, imageID string) (bool, error) {
	result, err := r.client.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.metadataTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ImageID": {
				S: aws.String(imageID),
			},
		},
	})

	if err != nil {
		return false, err
	}

	// 画像が存在するかどうかを確認
	return len(result.Item) > 0, nil
}
